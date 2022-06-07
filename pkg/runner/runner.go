package runner

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/nothinux/octo-proxy/pkg/config"
	"github.com/nothinux/octo-proxy/pkg/proxy"
)

// Octo hold list proxy server information
type Octo struct {
	sync.Mutex
	Proxies map[string]*proxy.Proxy
}

type Server config.Config

func Run(c *config.Config) error {
	ss := &Server{c.ServerConfigs}
	proxies := ss.runProxy()

	guardian := &Octo{
		Proxies: proxies,
	}

	sigTerm := make(chan os.Signal, 1)
	sigReload := make(chan os.Signal, 1)

	signal.Notify(sigTerm, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	signal.Notify(sigReload, syscall.SIGUSR1, syscall.SIGUSR2)

alive:
	for {
		select {
		case <-sigTerm:
			guardian.Lock()
			shutdown(guardian.Proxies)
			guardian.Unlock()
			break alive
		case <-sigReload:
		}
	}

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

func shutdown(proxies map[string]*proxy.Proxy) {
	for _, p := range proxies {
		if p.Listener != nil {
			close(p.Quit)
			p.Listener.Close()
			p.Wg.Wait()
		}
	}
}
