package proxy

import (
	"crypto/tls"
	"io"
	"log"
	"math/rand"
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

func dialTargets(hcs []config.HostConfig) (net.Conn, error) {
	for _, hc := range hcs {
		c, err := dialTarget(hc)
		if err == nil {
			c.SetDeadline(time.Now().Add(time.Second * time.Duration(hc.Timeout)))
			return c, nil
		}

		targetErr.Inc()
	}

	return nil, errors.New("targets", "no backends could be reached")
}

func getTargets(c config.ServerConfig) ([]net.Conn, io.Writer, error) {
	targets := c.Targets

	rand.Shuffle(len(targets), func(i, j int) {
		targets[i], targets[j] = targets[j], targets[i]
	})

	t, err := dialTargets(targets)
	if err != nil {
		return nil, nil, errors.New(c.Name, err.Error())
	}

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
