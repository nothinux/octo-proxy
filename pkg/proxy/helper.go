package proxy

import (
	"net"

	goerrors "errors"

	"github.com/nothinux/octo-proxy/pkg/config"
	"github.com/prometheus/client_golang/prometheus"
)

func errCopy(err error, tConf config.HostConfig) error {
	if err != nil {
		if goerrors.Is(err, net.ErrClosed) {
			// use of closed network connection
		} else {
			upstreamConnErr.With(prometheus.Labels{"host": tConf.Host, "port": tConf.Port}).Inc()
			return err
		}
	}

	return nil
}
