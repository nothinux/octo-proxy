package proxy

import (
	"crypto/tls"
	"errors"
	"strings"
	"testing"

	"github.com/nothinux/octo-proxy/pkg/config"
)

func TestNewTLS(t *testing.T) {
	pTLS := newTLS()

	if pTLS.MinVersion != tls.VersionTLS12 {
		t.Fatalf("got %v, want %v", pTLS.MinVersion, tls.VersionTLS12)
	}
}

func TestGetCACertPool(t *testing.T) {
	tests := []struct {
		Name      string
		Config    config.TLSConfig
		wantError error
	}{
		{
			Name:      "Check ca-cert file",
			Config:    config.TLSConfig{CaCert: "../testdata/ca-cert.pem"},
			wantError: nil,
		},
		{
			Name:      "Check if ca-cert file not found",
			Config:    config.TLSConfig{CaCert: "/tmp/zzz"},
			wantError: errors.New("no such file or directory"),
		},
		{
			Name:      "Check not ca-cert file",
			Config:    config.TLSConfig{CaCert: "../testdata/file"},
			wantError: errors.New("can't add CA to pool"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			_, err := getCACertPool(tt.Config)
			if err != tt.wantError {
				if !strings.Contains(err.Error(), tt.wantError.Error()) {
					t.Fatalf("got %v, want %v", err, tt.wantError)
				}
			}
		})
	}
}
