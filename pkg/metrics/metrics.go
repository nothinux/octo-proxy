package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

func AddGauge(name, help string) prometheus.Gauge {
	return promauto.NewGauge(prometheus.GaugeOpts{
		Name: name,
		Help: help,
	})
}

func AddCounter(name, help string) prometheus.Counter {
	return promauto.NewCounter(prometheus.CounterOpts{
		Name: name,
		Help: help,
	})
}
