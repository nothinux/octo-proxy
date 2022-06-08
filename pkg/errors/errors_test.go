package errors

import (
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	err := New("server", "no listener found")
	if !strings.Contains(err.Error(), "[server] no listener found") {
		t.Fatal("the error must be [server] no listener found")
	}
}
