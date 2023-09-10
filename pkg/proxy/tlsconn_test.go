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

func TestIsCertificateRevoked(t *testing.T) {
	caCert, err := getCACertificate(config.TLSConfig{CaCert: "../testdata/ca-cert.pem"})
	if err != nil {
		t.Fatal(err)
	}

	invalidCaCert, err := getCACertificate(config.TLSConfig{CaCert: "../testdata/wrong-ca.pem"})
	if err != nil {
		t.Fatal(err)
	}

	emptyCaCRL, err := getCACRL(config.TLSConfig{CRL: "../testdata/ca-crl.pem"})
	if err != nil {
		t.Fatal(err)
	}

	caCRL, err := getCACRL(config.TLSConfig{CRL: "../testdata/ca-crl-20230910095828.pem"})
	if err != nil {
		t.Fatal(err)
	}

	cert, err := getCACertificate(config.TLSConfig{CaCert: "../testdata/client.pem"})
	if err != nil {
		t.Fatal(err)
	}

	invalidCaCRL, err := getCACRL(config.TLSConfig{CRL: "../testdata/ca-crl-20230910100117.pem"})
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		Name      string
		Opts      VerifyOpts
		wantError error
	}{
		{
			Name: "Test valid certificate",
			Opts: VerifyOpts{
				CaCert: caCert,
				Cert:   cert,
				CRL:    emptyCaCRL,
			},
			wantError: nil,
		},
		{
			Name: "Test revoked certificate",
			Opts: VerifyOpts{
				CaCert: caCert,
				Cert:   cert,
				CRL:    caCRL,
			},
			wantError: errors.New("certificate was revoked and no longer valid - CN:certify"),
		},
		{
			Name: "Test invalid crl",
			Opts: VerifyOpts{
				CaCert: caCert,
				Cert:   cert,
				CRL:    invalidCaCRL,
			},
			wantError: errors.New("crl file is outdated"),
		},
		{
			Name: "Test invalid cacert",
			Opts: VerifyOpts{
				CaCert: invalidCaCert,
				Cert:   cert,
				CRL:    caCRL,
			},
			wantError: errors.New("x509: ECDSA verification failure"),
		},
	}

	for _, tt := range tests {
		_, err := isCertificateRevoked(tt.Opts)
		if err != tt.wantError {
			if !strings.Contains(err.Error(), tt.wantError.Error()) {
				t.Fatalf("got %v, want %v", err.Error(), tt.wantError)
			}
		}
	}
}
