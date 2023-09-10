package proxy

import (
	"errors"
	"strings"
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

func TestGetCertKeyPair(t *testing.T) {
	tests := []struct {
		Name      string
		Config    config.TLSConfig
		wantError error
	}{
		{
			Name: "Check if cert file not found",
			Config: config.TLSConfig{
				Cert: "../testdata/zzzz",
			},
			wantError: errors.New("no such file or directory"),
		},
		{
			Name: "Check if key file not found",
			Config: config.TLSConfig{
				Cert: "../testdata/cert.pem",
				Key:  "../testdata/zzzz",
			},
			wantError: errors.New("no such file or directory"),
		},
		{
			Name: "Check if cert and key is not match",
			Config: config.TLSConfig{
				Cert: "../testdata/cert.pem",
				Key:  "../testdata/ca-key.pem",
			},
			wantError: errors.New("can't parse public & private key pair tls: private key does not match public key"),
		},
		{
			Name: "Check if cert and key is match",
			Config: config.TLSConfig{
				Cert: "../testdata/cert.pem",
				Key:  "../testdata/cert-key.pem",
			},
			wantError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			_, err := getCertKeyPair(tt.Config)
			if err != tt.wantError {
				if !strings.Contains(err.Error(), tt.wantError.Error()) {
					t.Fatalf("got %v, want %v", err, tt.wantError)
				}
			}
		})
	}
}

func TestGetCACertificate(t *testing.T) {
	tests := []struct {
		Name      string
		Config    config.TLSConfig
		wantError error
	}{
		{
			Name: "Test if cert file not found",
			Config: config.TLSConfig{
				CaCert: "../testdata/zzzz",
			},
			wantError: errors.New("no such file or directory"),
		},
		{
			Name: "Test invalid certificate file",
			Config: config.TLSConfig{
				CaCert: "../testdata/file",
			},
			wantError: errors.New("can't decode CA file"),
		},
		{
			Name: "Test valid certificate file",
			Config: config.TLSConfig{
				CaCert: "../testdata/ca-cert.pem",
			},
			wantError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			_, err := getCACertificate(tt.Config)
			if err != tt.wantError {
				if !strings.Contains(err.Error(), tt.wantError.Error()) {
					t.Fatalf("got %v, want %v", err, tt.wantError)
				}
			}
		})
	}
}

func TestGetCACRL(t *testing.T) {
	tests := []struct {
		Name      string
		Config    config.TLSConfig
		wantError error
	}{
		{
			Name: "Test if cert file not found",
			Config: config.TLSConfig{
				CRL: "../testdata/zzzz",
			},
			wantError: errors.New("no such file or directory"),
		},
		{
			Name: "Test invalid certificate file",
			Config: config.TLSConfig{
				CRL: "../testdata/file",
			},
			wantError: errors.New("can't decode CRL file"),
		},
		{
			Name: "Test valid certificate file",
			Config: config.TLSConfig{
				CRL: "../testdata/ca-crl.pem",
			},
			wantError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			_, err := getCACRL(tt.Config)
			if err != tt.wantError {
				if !strings.Contains(err.Error(), tt.wantError.Error()) {
					t.Fatalf("got %v, want %v", err, tt.wantError)
				}
			}
		})
	}
}
