package config

import (
	"bytes"
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
