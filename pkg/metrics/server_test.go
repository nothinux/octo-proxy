package metrics

import (
	"testing"

	"github.com/nothinux/octo-proxy/pkg/config"
)

func TestNew(t *testing.T) {
	tests := []struct {
		Name         string
		Config       config.HostConfig
		expectedAddr string
	}{
		{
			Name: "Test valid metrics configuration",
			Config: config.HostConfig{
				Host: "127.0.0.1",
				Port: "9127",
			},
			expectedAddr: "127.0.0.1:9127",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			m := New(tt.Config)

			if m.Addr != tt.expectedAddr {
				t.Fatalf("got %v, want %v", m.Addr, tt.expectedAddr)
			}
		})
	}
}
