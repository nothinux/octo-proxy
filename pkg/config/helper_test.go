package config

import (
	"reflect"
	"testing"
)

func TestHostIpIsValid(t *testing.T) {
	t.Run("test valid ip address", func(t *testing.T) {
		if !hostIPIsValid("127.0.0.1") {
			t.Fatalf("ip must be valid")
		}
	})
	t.Run("test invalid ip address", func(t *testing.T) {
		if hostIPIsValid("127.0.0.o") {
			t.Fatalf("ip must be invalid")
		}
	})
}

func TestParseSubjectAltNames(t *testing.T) {
	tests := []struct {
		Name        string
		sans        []string
		expectedSAN *SubjectAltName
	}{
		{
			Name: "test ip san",
			sans: []string{"127.0.0.1", "192.168.1.1"},
			expectedSAN: &SubjectAltName{
				IPAddress: []string{"127.0.0.1", "192.168.1.1"},
			},
		},
		{
			Name: "test ip and uri san",
			sans: []string{"127.0.0.1", "192.168.1.1", "http://localhost"},
			expectedSAN: &SubjectAltName{
				IPAddress: []string{"127.0.0.1", "192.168.1.1"},
				Uri:       []string{"http://localhost"},
			},
		},
		{
			Name: "test ip, uri, and dns san",
			sans: []string{"127.0.0.1", "192.168.1.1", "spiffe://example.org/ns/spire/sa/spire-agent", "github.com"},
			expectedSAN: &SubjectAltName{
				IPAddress: []string{"127.0.0.1", "192.168.1.1"},
				Uri:       []string{"spiffe://example.org/ns/spire/sa/spire-agent"},
				DNS:       []string{"github.com"},
			},
		},
		{
			Name: "test dns san",
			sans: []string{"github.com", "gitlab.com"},
			expectedSAN: &SubjectAltName{
				DNS: []string{"github.com", "gitlab.com"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			san := parseSubjectAltNames(tt.sans)

			if !reflect.DeepEqual(san, tt.expectedSAN) {
				t.Fatalf("got %v, want %v", san, tt.expectedSAN)
			}
		})
	}
}

func TestPortIsValid(t *testing.T) {
	t.Run("test valid port", func(t *testing.T) {
		if !portIsValid("80") {
			t.Fatalf("port must be valid")
		}
	})
	t.Run("test invalid port", func(t *testing.T) {
		if portIsValid("127.0.0.o") {
			t.Fatalf("port must be invalid")
		}
	})
}
