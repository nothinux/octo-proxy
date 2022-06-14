package config

import (
	"bytes"
	"reflect"
	"strings"
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

func TestReadContent(t *testing.T) {
	t.Run("Test doesn't exists file", func(t *testing.T) {
		_, err := readContent("/tmp/zzzzzs")
		if err != nil {
			if !strings.Contains(err.Error(), "open /tmp/zzzzzs: no such file or directory") {
				t.Fatal(err)
			}
		}
	})

	t.Run("Test read content", func(t *testing.T) {
		b, err := readContent("../testdata/file.txt")
		if err != nil {
			t.Fatal(err)
		}

		r := bytes.Compare(b, []byte{104, 101, 108, 108, 111})
		if r != 0 {
			t.Fatalf("got %v, want %v", b, []byte{104, 101, 108, 108, 111})
		}
	})
}
