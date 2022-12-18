package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

func AddGaugeVec(name, help string) *prometheus.GaugeVec {
	return promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: name,
		Help: help,
	}, []string{"name"})
}

func AddGaugeVecMultiLabels(name, help string) *prometheus.GaugeVec {
	return promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: name,
		Help: help,
	}, []string{"host", "port"})
}

func AddCounterVec(name, help string) *prometheus.CounterVec {
	return promauto.NewCounterVec(prometheus.CounterOpts{
		Name: name,
		Help: help,
	}, []string{"name"})
}

func AddCounterVecMultiLabels(name, help string) *prometheus.CounterVec {
	return promauto.NewCounterVec(prometheus.CounterOpts{
		Name: name,
		Help: help,
	}, []string{"host", "port"})
}
