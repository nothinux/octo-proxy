package proxy

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	reuseport "github.com/kavu/go_reuseport"
	"github.com/nothinux/octo-proxy/pkg/config"
	"github.com/nothinux/octo-proxy/pkg/metrics"
)

var (
	activeConn      = metrics.AddGauge("octo_conn_active", "current active connection")
	activeConnTotal = metrics.AddCounter("octo_conn_active_total", "total active connection")
	connErrTotal    = metrics.AddCounter("octo_conn_error_total", "total connection error. include tcp and tls")
	tlsConnErrTotal = metrics.AddCounter("octo_conn_tls_error_total", "total tls connection error")
)

// Proxy hold running proxy data
type Proxy struct {
	Name     string
	Listener net.Listener
	Quit     chan interface{}
	Wg       sync.WaitGroup
	sync.Mutex
}

// New initialize new proxy
func New(name string) *Proxy {
	return &Proxy{
		Name: name,
		Quit: make(chan interface{}),
	}
}

// Run initialize tcp or tls listener
func (p *Proxy) Run(c config.ServerConfig) {
	l, err := reuseport.Listen("tcp", net.JoinHostPort(c.Listener.Host, c.Listener.Port))
	if err != nil {
		log.Fatal(err)
	}

	ts := []string{}

	for _, target := range c.Targets {
		ts = append(ts, fmt.Sprintf("%s:%s", target.Host, target.Port))
	}

	log.Printf(
		"listening %s on %s:%s -> %v",
		c.Name,
		c.Listener.Host, c.Listener.Port,
		ts,
	)

	tc := c.Listener.TLSConfig
	if tc.IsSimple() || tc.IsMutual() {
		tlsConf, err := getTLSConfig(tc)
		if err != nil {
			log.Fatal(err)
		}

		tlsListen := tls.NewListener(l, tlsConf.Config)
		p.Lock()
		p.Listener = tlsListen
		p.Unlock()

		log.Printf("running %s with tls mode %s", c.Name, c.Listener.TLSConfig.Mode)
	} else {
		p.Lock()
		p.Listener = l
		p.Unlock()
	}

	p.handleConn(c)
}

// handleConn accept incoming connection and forward it
func (p *Proxy) handleConn(c config.ServerConfig) {
	for {
		srcConn, err := p.Listener.Accept()
		if err != nil {
			select {
			case <-p.Quit:
				return
			default:
				log.Println(err)
				connErrTotal.Inc()
			}
		}

		activeConn.Inc()
		defer activeConn.Dec()
		activeConnTotal.Inc()

		srcConn.SetDeadline(time.Now().Add(time.Second * time.Duration(c.Listener.Timeout)))

		if err := isTLSConn(srcConn); err != nil {
			log.Println(err)
			srcConn.Close()
			tlsConnErrTotal.Inc()
			continue
		}

		p.Wg.Add(1)
		go func() {
			p.forwardConn(c, srcConn)
			p.Wg.Done()
		}()
	}
}

// forwardConn forward source connection to rarget or destination
func (p *Proxy) forwardConn(c config.ServerConfig, srcConn net.Conn) {
	targetConn, targetWr, err := getTargets(c)
	if err != nil {
		log.Println(err)
		srcConn.Close()
		return
	}

	defer srcConn.Close()
	defer closeConn(targetConn)

	p.Wg.Add(1)
	go func() {
		defer srcConn.Close()
		defer closeConn(targetConn)

		_, err := io.Copy(srcConn, targetConn[0])
		errCopy(err)

		p.Wg.Done()
	}()

	_, err = io.Copy(targetWr, srcConn)
	errCopy(err)
}

func (p *Proxy) Shutdown() {
	p.Lock()
	close(p.Quit)
	if p.Listener != nil {
		p.Listener.Close()
	}
	p.Wg.Wait()
	p.Unlock()
}
