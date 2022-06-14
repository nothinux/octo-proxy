package proxy

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"reflect"
	"time"

	"github.com/nothinux/octo-proxy/pkg/config"
	"github.com/nothinux/octo-proxy/pkg/errors"
	"github.com/nothinux/octo-proxy/pkg/metrics"
)

var (
	targetErr = metrics.AddCounter("octo_target_error_total", "total target or backend error")
	mirrorErr = metrics.AddCounter("octo_mirror_error_total", "total mirror error")
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
		targetErr.Inc()
		return nil, nil, errors.New("target", fmt.Sprintf("can't dial backend %s:%s %v", c.Target.Host, c.Target.Port, err))
	}
	t.SetDeadline(time.Now().Add(time.Second * time.Duration(c.Target.Timeout)))

	var m net.Conn

	if !reflect.DeepEqual(config.HostConfig{}, c.Mirror) {
		m, err = dialTarget(c.Mirror)
		if err != nil {
			mirrorErr.Inc()
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
