package proxy

import (
	"testing"
	"time"

	"github.com/nothinux/octo-proxy/pkg/config"
)

func TestTimeoutIsZero(t *testing.T) {
	tests := []struct {
		Name     string
		Config   config.HostConfig
		Response bool
	}{
		{
			Name: "Timeout is zero",
			Config: config.HostConfig{
				ConnectionConfig: config.ConnectionConfig{
					TimeoutDuration: time.Duration(0) * time.Second,
				},
			},
			Response: true,
		},
		{
			Name: "Timeout is not zero",
			Config: config.HostConfig{
				ConnectionConfig: config.ConnectionConfig{
					TimeoutDuration: time.Duration(100) * time.Second,
				},
			},
			Response: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			r := timeoutIsZero(tt.Config)

			if r != tt.Response {
				t.Fatalf("got %v, want %v", r, tt.Response)
			}
		})
	}
}
