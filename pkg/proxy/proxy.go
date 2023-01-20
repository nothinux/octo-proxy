package proxy

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	reuseport "github.com/kavu/go_reuseport"
	"github.com/nothinux/octo-proxy/pkg/config"
	"github.com/nothinux/octo-proxy/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
)

var (
	downstreamConnActive = metrics.AddGaugeVec("octo_downstream_conn_active", "current active connection in downstream")
	downstreamConnTotal  = metrics.AddCounterVec("octo_downstream_conn_total", "total downstream connection")
	downstreamConnErr    = metrics.AddCounterVec("octo_downstream_conn_error", "total downsream connection error. include tcp and tls")

	upstreamConnActive = metrics.AddGaugeVecMultiLabels("octo_upstream_conn_active", "current active connection in upstreamn")
	upstreamConnTotal  = metrics.AddCounterVecMultiLabels("octo_upstream_conn_total", "total upstream connection")
	upstreamConnErr    = metrics.AddCounterVecMultiLabels("octo_upstream_conn_error", "total upstream connection error. include tcp and tls")
)

// Proxy hold running proxy data
type Proxy struct {
	Name     string
	Listener net.Listener
	Quit     context.CancelFunc
	Wg       sync.WaitGroup
	sync.Mutex
}

// New initialize new proxy
func New(name string) *Proxy {
	return &Proxy{
		Name: name,
	}
}

// Run initialize tcp or tls listener
func (p *Proxy) Run(c config.ServerConfig) {
	p.Lock()
	if p.Quit != nil {
		p.Quit()
	}

	ctx, cancel := context.WithCancel(context.Background())
	p.Quit = cancel
	p.Unlock()

	l, err := reuseport.Listen("tcp", net.JoinHostPort(c.Listener.Host, c.Listener.Port))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to listen")
	}

	ts := []string{}

	for _, target := range c.Targets {
		ts = append(ts, fmt.Sprintf("%s:%s", target.Host, target.Port))
	}

	log.Info().
		Str("name", c.Name).
		Str("host", c.Listener.Host).
		Str("port", c.Listener.Port).
		Strs("targets", ts).
		Msg("running server")

	tc := c.Listener.TLSConfig
	if tc.IsSimple() || tc.IsMutual() {
		tlsConf, err := getTLSConfig(tc)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get TLS config")
		}

		tlsListen := tls.NewListener(l, tlsConf.Config)
		p.Lock()
		p.Listener = tlsListen
		p.Unlock()

		log.Info().
			Str("name", c.Name).
			Str("mode", c.Listener.TLSConfig.Mode).
			Msg("running in TLS mode")
	} else {
		p.Lock()
		p.Listener = l
		p.Unlock()
	}

	p.handleConn(ctx, c)
}

// handleConn accept incoming connection and forward it
func (p *Proxy) handleConn(ctx context.Context, c config.ServerConfig) {
	for {
		srcConn, err := p.Listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				log.Error().
					Err(err).
					Str("name", c.Name).
					Msg("connection error")
				downstreamConnErr.With(prometheus.Labels{"name": p.Name}).Inc()
			}
		}

		downstreamConnActive.With(prometheus.Labels{"name": p.Name}).Inc()
		downstreamConnTotal.With(prometheus.Labels{"name": p.Name}).Inc()

		if !timeoutIsZero(c.Listener) {
			srcConn.SetDeadline(time.Now().Add(c.Listener.TimeoutDuration))
		}

		if err := isTLSConn(srcConn); err != nil {
			log.Error().Err(err).Msg("connection error")
			srcConn.Close()
			downstreamConnErr.With(prometheus.Labels{"name": p.Name}).Inc()
			continue
		}

		p.Wg.Add(1)
		go func() {
			p.forwardConn(ctx, c, srcConn)
			p.Wg.Done()
			downstreamConnActive.With(prometheus.Labels{"name": p.Name}).Dec()
		}()
	}
}

// forwardConn forward source connection to rarget or destination
func (p *Proxy) forwardConn(ctx context.Context, c config.ServerConfig, srcConn net.Conn) {
	targetConn, targetWr, tConf, err := getTargets(c)
	if err != nil {
		log.Error().
			Err(err).
			Str("name", c.Name).
			Msg("failed to get targets")
		srcConn.Close()
		return
	}

	// Close long-lived connections to targets that have no timeout configured forcefully on shutdown.
	if timeoutIsZero(tConf) {
		p.Wg.Add(1)
		go func() {
			<-ctx.Done()
			closeConn(targetConn)

			p.Wg.Done()
		}()
	}

	defer srcConn.Close()
	defer closeConn(targetConn)
	defer upstreamConnActive.With(prometheus.Labels{"host": tConf.Host, "port": tConf.Port}).Dec()

	p.Wg.Add(1)
	go func() {
		defer srcConn.Close()
		defer closeConn(targetConn)

		_, err := io.Copy(srcConn, targetConn[0])
		errCopy(err, tConf)

		p.Wg.Done()
	}()

	upstreamConnActive.With(prometheus.Labels{"host": tConf.Host, "port": tConf.Port}).Inc()
	upstreamConnTotal.With(prometheus.Labels{"host": tConf.Host, "port": tConf.Port}).Inc()

	_, err = io.Copy(targetWr, srcConn)
	errCopy(err, tConf)
}

func (p *Proxy) Shutdown() {
	p.Lock()
	if p.Quit != nil {
		p.Quit()
		p.Quit = nil
	}
	if p.Listener != nil {
		p.Listener.Close()
	}
	p.Wg.Wait()
	p.Unlock()
}
