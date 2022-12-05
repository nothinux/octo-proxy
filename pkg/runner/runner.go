package runner

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/nothinux/octo-proxy/pkg/config"
	"github.com/nothinux/octo-proxy/pkg/metrics"
	"github.com/nothinux/octo-proxy/pkg/proxy"
	"github.com/okzk/sdnotify"
)

// Octo hold list proxy server information
type Octo struct {
	sync.Mutex
	Proxies map[string]*proxy.Proxy
}

type Server config.Config

func Run(c *config.Config, cPath string) error {
	ss := &Server{c.ServerConfigs}
	proxies := ss.runProxy()

	octo := &Octo{
		Proxies: proxies,
	}

	srv := runMetrics()

	sigTerm := make(chan os.Signal, 1)
	sigReload := make(chan os.Signal, 1)

	signal.Notify(sigTerm, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	signal.Notify(sigReload, syscall.SIGUSR1, syscall.SIGUSR2)

	sdnotify.Ready()

alive:
	for {
		select {
		case <-sigTerm:
			log.Println("octo-proxy interrupted")

			octo.Lock()
			shutdown(octo.Proxies, srv)
			octo.Unlock()

			break alive
		case <-sigReload:
			log.Println("octo-proxy reload trigerred")

			if err := reloadProxy(cPath, octo); err != nil {
				log.Println(err)
				log.Println("octo-proxy reload failed")
				continue
			}

			log.Println("octo-proxy reloaded")
		}
	}

	return nil
}

func reloadProxy(cPath string, octo *Octo) error {
	c, err := config.New(cPath)
	if err != nil {
		return err
	}

	ss := &Server{c.ServerConfigs}
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

func runMetrics() *metrics.Metrics {
	m := metrics.New()

	go func() {
		log.Println("listening metrics server on :9123")
		m.Run()
	}()

	return m
}

func shutdown(proxies map[string]*proxy.Proxy, m *metrics.Metrics) {
	log.Println("shutdown octo-proxy")

	for _, p := range proxies {
		p.Shutdown()
	}

	if m != nil {
		if err := m.Shutdown(context.Background()); err != nil {
			log.Println(err)
		}
	}
}
