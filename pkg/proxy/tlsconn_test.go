package proxy

import (
	"crypto/tls"
	"testing"
)

func TestNewTLS(t *testing.T) {
	pTLS := newTLS()

	if pTLS.MinVersion != tls.VersionTLS12 {
		t.Fatalf("got %v, want %v", pTLS.MinVersion, tls.VersionTLS12)
	}
}
