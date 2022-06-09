package metrics

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	*http.Server
}

func New() *Metrics {
	r := http.NewServeMux()
	r.Handle("/metrics", promhttp.Handler())

	srv := &http.Server{
		Handler:      r,
		Addr:         ":9123",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	return &Metrics{srv}
}

func (m *Metrics) Run() error {
	if err := m.ListenAndServe(); err != nil {
		return err
	}

	return nil
}
