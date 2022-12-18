package metrics

import (
	"reflect"
	"testing"

	pcm "github.com/prometheus/client_model/go"

	"github.com/prometheus/client_golang/prometheus"
)

func TestAddGaugeVec(t *testing.T) {
	tests := []struct {
		Name            string
		MetricsName     string
		MetricsHelp     string
		Labels          prometheus.Labels
		ExpectedMetrics *prometheus.Desc
	}{
		{
			Name:        "Test metrics is correct",
			MetricsName: "example_metrics",
			MetricsHelp: "help",
			Labels:      prometheus.Labels{"name": "www"},
			ExpectedMetrics: prometheus.NewDesc(
				"example_metrics",
				"help",
				[]string{"name"},
				prometheus.Labels{},
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			m := AddGaugeVec(tt.MetricsName, tt.MetricsHelp)

			if !reflect.DeepEqual(m.With(tt.Labels).Desc(), tt.ExpectedMetrics) {
				t.Fatalf("got %v, want %v", m.With(tt.Labels).Desc(), tt.ExpectedMetrics)
			}

			metrics := &pcm.Metric{}

			t.Run("Test increase value", func(t *testing.T) {
				m.With(tt.Labels).Inc()
				m.With(tt.Labels).Write(metrics)

				if metrics.Gauge.GetValue() != 1 {
					t.Fatalf("got %v, want %v", metrics.Gauge.GetValue(), 1)
				}
			})

			t.Run("Test decrease value", func(t *testing.T) {
				m.With(tt.Labels).Dec()
				m.With(tt.Labels).Write(metrics)

				if metrics.Gauge.GetValue() != 0 {
					t.Fatalf("got %v, want %v", metrics.Gauge.GetValue(), 0)
				}
			})
		})
	}
}

func TestAddCounterVec(t *testing.T) {
	tests := []struct {
		Name            string
		MetricsName     string
		MetricsHelp     string
		Labels          prometheus.Labels
		ExpectedMetrics *prometheus.Desc
	}{
		{
			Name:        "Test metrics is correct",
			MetricsName: "example_metrics_counter",
			MetricsHelp: "help",
			Labels:      prometheus.Labels{"name": "www"},
			ExpectedMetrics: prometheus.NewDesc(
				"example_metrics_counter",
				"help",
				[]string{"name"},
				prometheus.Labels{},
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			m := AddCounterVec(tt.MetricsName, tt.MetricsHelp)

			if !reflect.DeepEqual(m.With(tt.Labels).Desc(), tt.ExpectedMetrics) {
				t.Fatalf("got %v, want %v", m.With(tt.Labels).Desc(), tt.ExpectedMetrics)
			}

			metrics := &pcm.Metric{}

			t.Run("Test increase value", func(t *testing.T) {
				m.With(tt.Labels).Inc()
				m.With(tt.Labels).Write(metrics)

				if metrics.Counter.GetValue() != 1 {
					t.Fatalf("got %v, want %v", metrics.Counter.GetValue(), 1)
				}
			})
		})
	}
}

func TestAddGaugeVecMultiLabels(t *testing.T) {
	tests := []struct {
		Name            string
		MetricsName     string
		MetricsHelp     string
		Labels          prometheus.Labels
		ExpectedMetrics *prometheus.Desc
	}{
		{
			Name:        "Test metrics is correct",
			MetricsName: "example_metrics_m",
			MetricsHelp: "help",
			Labels:      prometheus.Labels{"host": "127.0.0.1", "port": "8080"},
			ExpectedMetrics: prometheus.NewDesc(
				"example_metrics_m",
				"help",
				[]string{"host", "port"},
				prometheus.Labels{},
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			m := AddGaugeVecMultiLabels(tt.MetricsName, tt.MetricsHelp)

			if !reflect.DeepEqual(m.With(tt.Labels).Desc(), tt.ExpectedMetrics) {
				t.Fatalf("got %v, want %v", m.With(tt.Labels).Desc(), tt.ExpectedMetrics)
			}

			metrics := &pcm.Metric{}

			t.Run("Test increase value", func(t *testing.T) {
				m.With(tt.Labels).Inc()
				m.With(tt.Labels).Write(metrics)

				if metrics.Gauge.GetValue() != 1 {
					t.Fatalf("got %v, want %v", metrics.Gauge.GetValue(), 1)
				}
			})

			t.Run("Test decrease value", func(t *testing.T) {
				m.With(tt.Labels).Dec()
				m.With(tt.Labels).Write(metrics)

				if metrics.Gauge.GetValue() != 0 {
					t.Fatalf("got %v, want %v", metrics.Gauge.GetValue(), 0)
				}
			})
		})
	}
}

func TestAddCounterVecMultiLabels(t *testing.T) {
	tests := []struct {
		Name            string
		MetricsName     string
		MetricsHelp     string
		Labels          prometheus.Labels
		ExpectedMetrics *prometheus.Desc
	}{
		{
			Name:        "Test metrics is correct",
			MetricsName: "example_metrics_counter_m",
			MetricsHelp: "help",
			Labels:      prometheus.Labels{"host": "127.0.0.1", "port": "6379"},
			ExpectedMetrics: prometheus.NewDesc(
				"example_metrics_counter_m",
				"help",
				[]string{"host", "port"},
				prometheus.Labels{},
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			m := AddCounterVecMultiLabels(tt.MetricsName, tt.MetricsHelp)

			if !reflect.DeepEqual(m.With(tt.Labels).Desc(), tt.ExpectedMetrics) {
				t.Fatalf("got %v, want %v", m.With(tt.Labels).Desc(), tt.ExpectedMetrics)
			}

			metrics := &pcm.Metric{}

			t.Run("Test increase value", func(t *testing.T) {
				m.With(tt.Labels).Inc()
				m.With(tt.Labels).Write(metrics)

				if metrics.Counter.GetValue() != 1 {
					t.Fatalf("got %v, want %v", metrics.Counter.GetValue(), 1)
				}
			})
		})
	}
}
