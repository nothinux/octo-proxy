package proxy

import (
	"crypto/tls"
	"io"
	"log"
	"net"
	"time"

	"github.com/nothinux/octo-proxy/pkg/config"
)

func newDial() *net.Dialer {
	return &net.Dialer{
		Timeout: 5 * time.Second,
	}
}

func dialTarget(hc config.HostConfig) (net.Conn, error) {
	d := newDial()

	if hc.IsSimple() || hc.IsMutual() {
		tlsConf, err := getTLSConfig(hc.TLSConfig)
		if err != nil {
			return nil, err
		}

		log.Println("called tls target")
		return tls.DialWithDialer(d, "tcp", net.JoinHostPort(hc.Host, hc.Port), tlsConf.Config)
	}

	return d.Dial("tcp", net.JoinHostPort(hc.Host, hc.Port))
}

func getTargets(c config.ServerConfig) ([]net.Conn, io.Writer, error) {
	t, err := dialTarget(c.Target)
	if err != nil {
		log.Printf("can't dial backend %s:%s %v", c.Target.Host, c.Target.Port, err)
		//srcConn.Close()
		return nil, nil, err
	}
	t.SetDeadline(time.Now().Add(time.Second * time.Duration(c.Listener.Timeout)))

	var m net.Conn

	if (config.HostConfig{}) != c.Mirror {
		m, err = dialTarget(c.Mirror)
		if err != nil {
			log.Printf("can't dial mirror backend %s:%s %v", c.Mirror.Host, c.Mirror.Port, err)
		}
		if m != nil {
			m.SetDeadline(time.Now().Add(time.Second * time.Duration(c.Mirror.Timeout)))
		}
	}

	if m == nil {
		return []net.Conn{t}, io.MultiWriter(t), nil
	}

	return []net.Conn{t, m}, io.MultiWriter(t, m), nil
}