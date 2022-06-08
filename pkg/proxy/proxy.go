package proxy

import (
	"crypto/tls"
	"io"
	"log"
	"net"
	"sync"
	"time"

	reuseport "github.com/kavu/go_reuseport"
	"github.com/nothinux/octo-proxy/pkg/config"
)

// Proxy hold running proxy data
type Proxy struct {
	Name     string
	Listener net.Listener
	Quit     chan interface{}
	Wg       sync.WaitGroup
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

	log.Printf(
		"listening %s on %s:%s -> %s:%s",
		c.Name,
		c.Listener.Host, c.Listener.Port,
		c.Target.Host, c.Target.Port,
	)

	tc := c.Listener.TLSConfig
	if tc.IsSimple() || tc.IsMutual() {
		tlsConf, err := getTLSConfig(tc)
		if err != nil {
			log.Fatal(err)
		}

		tlsListen := tls.NewListener(l, tlsConf.Config)
		p.Listener = tlsListen

		log.Printf("running %s with tls mode %s", c.Name, c.Listener.TLSConfig.Mode)
	} else {
		p.Listener = l
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
			}
		}

		srcConn.SetDeadline(time.Now().Add(time.Second * time.Duration(c.Listener.Timeout)))

		if err := isTLSConn(srcConn); err != nil {
			log.Println(err)
			srcConn.Close()
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
	close(p.Quit)
	p.Listener.Close()
	p.Wg.Wait()
}
