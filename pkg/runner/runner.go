package runner

import (
	"context"
	"os"
	"os/signal"
	"reflect"
	"sync"
	"syscall"

	"github.com/nothinux/octo-proxy/pkg/config"
	"github.com/nothinux/octo-proxy/pkg/metrics"
	"github.com/nothinux/octo-proxy/pkg/proxy"
	"github.com/okzk/sdnotify"
	"github.com/rs/zerolog/log"
)

// Octo hold list proxy server information
type Octo struct {
	sync.Mutex
	Proxies map[string]*proxy.Proxy
}

type Server config.Config

func Run(c *config.Config, cPath string) error {
	ss := &Server{
		ServerConfigs: c.ServerConfigs,
		MetricsConfig: c.MetricsConfig,
	}

	proxies := ss.runProxy()

	octo := &Octo{
		Proxies: proxies,
	}

	var metricsServer *metrics.Metrics

	if !reflect.DeepEqual(c.MetricsConfig, config.HostConfig{}) {
		var err error
		metricsServer, err = runMetrics(c.MetricsConfig)
		if err != nil {
			return err
		}
	}

	sigTerm := make(chan os.Signal, 1)
	sigReload := make(chan os.Signal, 1)

	signal.Notify(sigTerm, os.Interrupt, syscall.SIGTERM)
	signal.Notify(sigReload, syscall.SIGUSR1, syscall.SIGUSR2)

	sdnotify.Ready()

alive:
	for {
		select {
		case <-sigTerm:
			log.Warn().Msg("octo-proxy interrupted")

			octo.Lock()
			shutdown(octo.Proxies, metricsServer)
			octo.Unlock()

			break alive
		case <-sigReload:
			log.Info().Msg("octo-proxy reload triggered")

			if err := reloadProxy(cPath, octo); err != nil {
				log.Error().Err(err).Msg("octo-proxy reload failed")
				continue
			}

			log.Info().Msg("octo-proxy reloaded")
		}
	}

	return nil
}

func reloadProxy(cPath string, octo *Octo) error {
	c, err := config.New(cPath)
	if err != nil {
		return err
	}

	ss := &Server{
		ServerConfigs: c.ServerConfigs,
		MetricsConfig: c.MetricsConfig,
	}

	proxies := ss.runProxy()

	octo.Lock()
	// shutdown old listener in
	oldOcto := octo.Proxies
	shutdown(oldOcto, nil)

	// set new listener
	octo.Proxies = proxies

	// done
	octo.Unlock()

	return nil
}

// runProxy
func (s *Server) runProxy() map[string]*proxy.Proxy {
	proxies := make(map[string]*proxy.Proxy)

	for _, serverConfig := range s.ServerConfigs {
		sc := serverConfig

		//
		proxy := proxy.New(sc.Name)
		proxies[sc.Name] = proxy

		// run proxy
		proxy.Wg.Add(1)
		go func() {
			proxy.Run(sc)
			proxy.Wg.Done()
		}()
	}

	return proxies
}

func runMetrics(c config.HostConfig) (*metrics.Metrics, error) {
	m := metrics.New(c)

	go func() {
		log.Info().
			Str("host", c.Host).
			Str("port", c.Port).
			Msg("starting metrics server")
		m.Run()
	}()

	return m, nil
}

func shutdown(proxies map[string]*proxy.Proxy, m *metrics.Metrics) {
	log.Info().Msg("shutdown octo-proxy")

	for _, p := range proxies {
		p.Shutdown()
	}

	if m != nil {
		if err := m.Shutdown(context.Background()); err != nil {
			log.Error().Err(err).Msg("shutdown failed")
		}
	}
}
