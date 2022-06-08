package proxy

import (
	"bytes"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/nothinux/octo-proxy/pkg/config"
	"github.com/nothinux/octo-proxy/pkg/testhelper"
)

var (
	messageByte = []byte("hello")
)

func TestProxy(t *testing.T) {
	var wg sync.WaitGroup
	result := make(chan []byte)

	// start target server
	backend := testhelper.RunTestServer(&wg, result)

	// start octo proxy
	cfg, err := config.GenerateConfig("127.0.0.1:9000", backend)
	if err != nil {
		t.Fatal(err)
	}
	p := New("test-proxy")
	go func() {
		p.Run(cfg.ServerConfigs[0])
	}()

	time.Sleep(1 * time.Second)

	// send data to octo-proxy
	if err := SendData(cfg.ServerConfigs[0].Listener, messageByte); err != nil {
		t.Fatal(err)
	}

	// check data received by test server
	t.Run("test message received is same", func(t *testing.T) {
		res := <-result

		r := bytes.Compare(messageByte, res)
		if r != 0 {
			t.Fatalf("got %v, want %v", res, messageByte)
		}
	})

	// shutdown octo-proxy
	p.Shutdown()
}

func TestProxyWithMirror(t *testing.T) {
	var wg sync.WaitGroup
	result := make(chan []byte)
	mirrorResult := make(chan []byte)

	// start target server
	backend := testhelper.RunTestServer(&wg, result)
	mirror := testhelper.RunTestServer(&wg, mirrorResult)

	// prepare octo proxy configuration
	cfg, err := config.GenerateConfig("127.0.0.1:9000", backend)
	if err != nil {
		t.Fatal(err)
	}
	cfg.ServerConfigs[0].Mirror = config.HostConfig{
		Host: strings.Split(mirror, ":")[0],
		Port: strings.Split(mirror, ":")[1],
	}

	p := New("test-proxy")
	go func() {
		p.Run(cfg.ServerConfigs[0])
	}()

	time.Sleep(1 * time.Second)

	// send data to octo-proxy
	if err := SendData(cfg.ServerConfigs[0].Listener, messageByte); err != nil {
		t.Fatal(err)
	}

	// check data received by test server
	t.Run("test message received is same", func(t *testing.T) {
		res := <-result

		r := bytes.Compare(messageByte, res)
		if r != 0 {
			t.Fatalf("got %v, want %v", res, messageByte)
		}
	})

	t.Run("test message received is same in mirror server", func(t *testing.T) {
		// TODO check
	})

	// shutdown octo-proxy
	p.Shutdown()
}

func TestProxyWithSimpleTLS(t *testing.T) {
	var wg sync.WaitGroup
	result := make(chan []byte)

	// start target server
	backend := testhelper.RunTestServer(&wg, result)

	// prepare configuration for octo proxy
	cfg, err := config.GenerateConfig("127.0.0.1:9000", backend)
	if err != nil {
		t.Fatal(err)
	}
	cfg.ServerConfigs[0].Listener.TLSConfig = config.TLSConfig{
		Mode: "simple",
		Cert: "../testdata/cert.pem",
		Key:  "../testdata/cert-key.pem",
	}

	p := New("test-proxy")
	go func() {
		p.Run(cfg.ServerConfigs[0])
	}()

	time.Sleep(1 * time.Second)

	// send data to octo-proxy
	hc := config.HostConfig{
		Host: "127.0.0.1",
		Port: "9000",
		TLSConfig: config.TLSConfig{
			Mode:   "simple",
			CaCert: "../testdata/ca-cert.pem",
		},
	}
	if err := SendData(hc, messageByte); err != nil {
		t.Fatal(err)
	}

	// check data received by test server
	t.Run("test message received is same", func(t *testing.T) {
		res := <-result

		r := bytes.Compare(messageByte, res)
		if r != 0 {
			t.Fatalf("got %v, want %v", res, messageByte)
		}
	})

	// shutdown octo-proxy
	p.Shutdown()
}

func TestProxyWithMutualTLS(t *testing.T) {
	var wg sync.WaitGroup
	result := make(chan []byte)

	// start target server
	backend := testhelper.RunTestServer(&wg, result)

	// prepare configuration for octo proxy
	cfg, err := config.GenerateConfig("127.0.0.1:9000", backend)
	if err != nil {
		t.Fatal(err)
	}
	cfg.ServerConfigs[0].Listener.TLSConfig = config.TLSConfig{
		Mode:   "mutual",
		Cert:   "../testdata/cert.pem",
		Key:    "../testdata/cert-key.pem",
		CaCert: "../testdata/ca-cert.pem",
		Role:   config.Role{Server: true},
	}

	p := New("test-proxy")
	go func() {
		p.Run(cfg.ServerConfigs[0])
	}()

	time.Sleep(1 * time.Second)

	// send data to octo-proxy
	// send data to octo-proxy
	hc := config.HostConfig{
		Host: "127.0.0.1",
		Port: "9000",
		TLSConfig: config.TLSConfig{
			Mode:   "simple",
			CaCert: "../testdata/ca-cert.pem",
			Cert:   "../testdata/client.pem",
			Key:    "../testdata/client-key.pem",
		},
	}
	if err := SendData(hc, messageByte); err != nil {
		t.Fatal(err)
	}

	// check data received by test server
	t.Run("test message received is same", func(t *testing.T) {
		res := <-result

		r := bytes.Compare(messageByte, res)
		if r != 0 {
			t.Fatalf("got %v, want %v", res, messageByte)
		}
	})

	// shutdown octo-proxy
	p.Shutdown()
}

func TestUnreachableTarget(t *testing.T) {
	// start octo proxy
	cfg, err := config.GenerateConfig("127.0.0.1:9000", "127.0.0.1:4")
	if err != nil {
		t.Fatal(err)
	}
	p := New("test-proxy")
	go func() {
		p.Run(cfg.ServerConfigs[0])
	}()

	time.Sleep(1 * time.Second)

	// send data to octo-proxy
	if err := SendData(cfg.ServerConfigs[0].Listener, messageByte); err != nil {
		t.Fatal(err)
	}

	// shutdown octo-proxy
	p.Shutdown()
}
