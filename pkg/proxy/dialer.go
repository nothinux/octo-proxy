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
	"github.com/prometheus/client_golang/prometheus"

	"github.com/rs/zerolog/log"
)

var (
	upstreamDialErr = metrics.AddCounterVecMultiLabels("octo_upstream_dial_error", "total dial error when calling an upstream")
	mirrorDialErr   = metrics.AddCounterVecMultiLabels("octo_mirror_dial_error", "total dial error when calling an mirror upstream")
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

		log.Debug().
			Str("host", hc.Host).
			Str("port", hc.Port).
			Msg("called tls target")

		return tls.DialWithDialer(d, "tcp", net.JoinHostPort(hc.Host, hc.Port), tlsConf.Config)
	}

	return d.Dial("tcp", net.JoinHostPort(hc.Host, hc.Port))
}

func dialTargets(hcs []config.HostConfig) (net.Conn, config.HostConfig, error) {
	tConf := &config.HostConfig{}

	for _, hc := range hcs {
		*tConf = hc
		c, err := dialTarget(hc)
		if err == nil {
			if !timeoutIsZero(hc) {
				c.SetDeadline(time.Now().Add(hc.TimeoutDuration))
			}
			return c, *tConf, nil
		}
	}

	return nil, *tConf, errors.New("targets", "no backends could be reached")
}

func getTargets(c config.ServerConfig) ([]net.Conn, io.Writer, config.HostConfig, error) {
	targets := c.Targets

	rand.Shuffle(len(targets), func(i, j int) {
		targets[i], targets[j] = targets[j], targets[i]
	})

	t, tc, err := dialTargets(targets)
	if err != nil {
		upstreamDialErr.With(prometheus.Labels{"host": tc.Host, "port": tc.Port}).Inc()
		return nil, nil, config.HostConfig{}, errors.New(c.Name, err.Error())
	}

	var m net.Conn

	if !reflect.DeepEqual(config.HostConfig{}, c.Mirror) {
		m, err = dialTarget(c.Mirror)
		if err != nil {
			mirrorDialErr.With(prometheus.Labels{"host": c.Mirror.Host, "port": c.Mirror.Port}).Inc()
			log.Warn().
				Err(err).
				Str("host", c.Mirror.Host).
				Str("port", c.Mirror.Port).
				Msg("can't dial mirror backend")
		}
		if m != nil {
			if !timeoutIsZero(c.Mirror) {
				m.SetDeadline(time.Now().Add(c.Mirror.TimeoutDuration))
			}
		}
	}

	if m == nil {
		return []net.Conn{t}, io.MultiWriter(t), tc, nil
	}

	return []net.Conn{t, m}, io.MultiWriter(t, m), tc, nil
}
