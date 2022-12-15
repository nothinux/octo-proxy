package proxy

import (
	"crypto/tls"
	"io"
	"math/rand"
	"net"
	"reflect"
	"time"

	"github.com/nothinux/octo-proxy/pkg/config"
	"github.com/nothinux/octo-proxy/pkg/errors"
	"github.com/nothinux/octo-proxy/pkg/metrics"
	"github.com/nothinux/octo-proxy/pkg/tlsconn"
	"github.com/rs/zerolog/log"
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
		tlsConf, err := tlsconn.GetTLSConfig(hc.TLSConfig)
		if err != nil {
			return nil, err
		}

		log.Debug().
			Str("host", hc.Host).
			Str("port", hc.Port).
			Msg("called tls target")

		return tls.DialWithDialer(d, "tcp", net.JoinHostPort(hc.Host, hc.Port), tlsConf.Config)
	}

	return d.Dial("tcp", net.JoinHostPort(hc.Host, hc.Port))
}

func dialTargets(hcs []config.HostConfig) (net.Conn, error) {
	for _, hc := range hcs {
		c, err := dialTarget(hc)
		if err == nil {
			if hc.Timeout > 0 {
				deadline := time.Now().Add(time.Second * time.Duration(hc.Timeout))
				if err := c.SetDeadline(deadline); err != nil {
					log.Error().Err(err).Msg("failed to set target conn deadline")
				}
			}
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
			log.Warn().
				Err(err).
				Str("host", c.Mirror.Host).
				Str("port", c.Mirror.Port).
				Msg("can't dial mirror backend")
		}
		if m != nil {
			if c.Mirror.Timeout > 0 {
				deadline := time.Now().Add(time.Second * time.Duration(c.Mirror.Timeout))
				if err := m.SetDeadline(deadline); err != nil {
					log.Error().Err(err).Msg("failed to set mirror conn deadline")
				}
			}
		}
	}

	if m == nil {
		return []net.Conn{t}, io.MultiWriter(t), nil
	}

	return []net.Conn{t, m}, io.MultiWriter(t, m), nil
}
