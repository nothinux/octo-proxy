package metrics

import (
	"fmt"
	"net/http"
	"time"

	"github.com/nothinux/octo-proxy/pkg/config"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	*http.Server
}

func New(c config.HostConfig) (*Metrics, error) {
	r := http.NewServeMux()
	r.Handle("/metrics", promhttp.Handler())

	srv := &http.Server{
		Handler:      r,
		Addr:         fmt.Sprintf("%s:%s", c.Host, c.Port),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	return &Metrics{srv}, nil
}

func (m *Metrics) Run() error {
	if err := m.ListenAndServe(); err != nil {
		return err
	}

	return nil
}
