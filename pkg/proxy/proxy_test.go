package proxy

import (
	"bytes"
	"errors"
	"io"
	"log"
	"strings"
	"sync"
	"syscall"
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
	cfg, err := config.GenerateConfig("127.0.0.1:9000", []string{backend}, "")
	if err != nil {
		t.Fatal(err)
	}
	p := New("test-proxy")
	go func() {
		p.Run(cfg.ServerConfigs[0])
	}()

	time.Sleep(1 * time.Second)

	// send data to octo-proxy
	if err := SendData(cfg.ServerConfigs[0].Listener, messageByte, false); err != nil {
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

func TestProxyWithMultipleTargets(t *testing.T) {
	var wg sync.WaitGroup
	result := make(chan []byte)

	// start target server
	backend := testhelper.RunTestServer(&wg, result)

	// start octo proxy
	cfg, err := config.GenerateConfig("127.0.0.1:9000", []string{"localhost:9001", backend}, "")
	if err != nil {
		t.Fatal(err)
	}
	p := New("test-proxy")
	go func() {
		p.Run(cfg.ServerConfigs[0])
	}()

	time.Sleep(1 * time.Second)

	// send data to octo-proxy
	if err := SendData(cfg.ServerConfigs[0].Listener, messageByte, false); err != nil {
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
	cfg, err := config.GenerateConfig("127.0.0.1:9000", []string{backend}, "")
	if err != nil {
		t.Fatal(err)
	}
	cfg.ServerConfigs[0].Mirror = config.HostConfig{
		Host:            strings.Split(mirror, ":")[0],
		Port:            strings.Split(mirror, ":")[1],
		TimeoutDuration: 10 * time.Second,
	}

	p := New("test-proxy")
	go func() {
		p.Run(cfg.ServerConfigs[0])
	}()

	time.Sleep(1 * time.Second)

	// send data to octo-proxy
	if err := SendData(cfg.ServerConfigs[0].Listener, messageByte, false); err != nil {
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
		mr := <-mirrorResult

		r := bytes.Compare(messageByte, mr)
		if r != 0 {
			t.Fatalf("got %v, want %v", mr, messageByte)
		}
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
	cfg, err := config.GenerateConfig("127.0.0.1:9000", []string{backend}, "")
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
	if err := SendData(hc, messageByte, false); err != nil {
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
	cfg, err := config.GenerateConfig("127.0.0.1:9000", []string{backend}, "")
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
	hc := config.HostConfig{
		Host: "127.0.0.1",
		Port: "9000",
		TLSConfig: config.TLSConfig{
			Mode:   "mutual",
			CaCert: "../testdata/ca-cert.pem",
			Cert:   "../testdata/client.pem",
			Key:    "../testdata/client-key.pem",
		},
	}
	if err := SendData(hc, messageByte, false); err != nil {
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

func TestProxyMutualTLSWhenClientUsingInvalidCertificate(t *testing.T) {
	var wg sync.WaitGroup
	result := make(chan []byte)

	// start target server
	backend := testhelper.RunTestServer(&wg, result)

	// prepare configuration for octo proxy
	cfg, err := config.GenerateConfig("127.0.0.1:9000", []string{backend}, "")
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
	hc := config.HostConfig{
		Host: "127.0.0.1",
		Port: "9000",
		TLSConfig: config.TLSConfig{
			Mode:   "mutual",
			CaCert: "../testdata/ca-cert.pem",
			Cert:   "../testdata/localhost.pem",
			Key:    "../testdata/localhost-key.pem",
		},
	}

	// check data received by test server
	t.Run("test get response from octo-proxy", func(t *testing.T) {
		if err := SendData(hc, messageByte, true); err != nil {
			if !strings.Contains(err.Error(), "tls: bad certificate") {
				t.Fatalf("got %v, want response contain tls: bad certificate", err)
			}
		}
	})

	// shutdown octo-proxy
	p.Shutdown()
}

func TestProxyMutualTLSWhenClientNotProvideCA(t *testing.T) {
	var wg sync.WaitGroup
	result := make(chan []byte)

	// start target server
	backend := testhelper.RunTestServer(&wg, result)

	// prepare configuration for octo proxy
	cfg, err := config.GenerateConfig("127.0.0.1:9000", []string{backend}, "")
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
			Mode: "mutual",
			Cert: "../testdata/localhost.pem",
			Key:  "../testdata/localhost-key.pem",
		},
	}

	// check data received by test server
	t.Run("test get response from octo-proxy", func(t *testing.T) {
		if err := SendData(hc, messageByte, true); err != nil {
			if !strings.Contains(err.Error(), "certificate signed by unknown authority") {
				t.Fatalf("got %v, want response contain certificate signed by unknown authority", err)
			}
		}
	})

	// shutdown octo-proxy
	p.Shutdown()
}

func TestProxyMutualTLSWhenClientUseWrongCA(t *testing.T) {
	var wg sync.WaitGroup
	result := make(chan []byte)

	// start target server
	backend := testhelper.RunTestServer(&wg, result)

	// prepare configuration for octo proxy
	cfg, err := config.GenerateConfig("127.0.0.1:9000", []string{backend}, "")
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
			Mode:   "mutual",
			CaCert: "../testdata/wrong-ca.pem",
			Cert:   "../testdata/localhost.pem",
			Key:    "../testdata/localhost-key.pem",
		},
	}

	// check data received by test server
	t.Run("test get response from octo-proxy", func(t *testing.T) {
		if err := SendData(hc, messageByte, true); err != nil {
			if !strings.Contains(err.Error(), "certificate signed by unknown authority") {
				t.Fatalf("got %v, want response contain certificate signed by unknown authority", err)
			}
		}
	})

	// shutdown octo-proxy
	p.Shutdown()
}

func TestProxyWithTargetSimpleTLS(t *testing.T) {
	var wg sync.WaitGroup
	result := make(chan []byte)

	// start target server
	tC := config.TLSConfig{
		Mode: "simple",
		Cert: "../testdata/cert.pem",
		Key:  "../testdata/cert-key.pem",
	}
	backend := RunTestTLSServer(&wg, tC, result)

	// prepare configuration for octo proxy
	cfg, err := config.GenerateConfig("127.0.0.1:9000", []string{backend}, "")
	if err != nil {
		t.Fatal(err)
	}
	cfg.ServerConfigs[0].Targets[0].TLSConfig = config.TLSConfig{
		Mode:   "simple",
		CaCert: "../testdata/ca-cert.pem",
	}

	p := New("test-proxy")
	go func() {
		p.Run(cfg.ServerConfigs[0])
	}()

	time.Sleep(1 * time.Second)

	// send data to octo-proxy with only tcp
	hc := config.HostConfig{
		Host: "127.0.0.1",
		Port: "9000",
	}
	if err := SendData(hc, messageByte, false); err != nil {
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

func TestProxyWithTargetMutualTLS(t *testing.T) {
	var wg sync.WaitGroup
	result := make(chan []byte)

	// start target server
	tC := config.TLSConfig{
		Mode:   "mutual",
		CaCert: "../testdata/ca-cert.pem",
		Cert:   "../testdata/cert.pem",
		Key:    "../testdata/cert-key.pem",
		Role: config.Role{
			Server: true,
		},
	}
	backend := RunTestTLSServer(&wg, tC, result)

	// prepare configuration for octo proxy
	cfg, err := config.GenerateConfig("127.0.0.1:9000", []string{backend}, "")
	if err != nil {
		t.Fatal(err)
	}
	cfg.ServerConfigs[0].Targets[0].TLSConfig = config.TLSConfig{
		Mode:   "mutual",
		CaCert: "../testdata/ca-cert.pem",
		Cert:   "../testdata/client.pem",
		Key:    "../testdata/client-key.pem",
	}

	p := New("test-proxy")
	go func() {
		p.Run(cfg.ServerConfigs[0])
	}()

	time.Sleep(1 * time.Second)

	// send data to octo-proxy with only tcp
	hc := config.HostConfig{
		Host: "127.0.0.1",
		Port: "9000",
	}
	if err := SendData(hc, messageByte, false); err != nil {
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
	cfg, err := config.GenerateConfig("127.0.0.1:9000", []string{"127.0.0.1:10"}, "")
	if err != nil {
		t.Fatal(err)
	}
	p := New("test-proxy")
	go func() {
		p.Run(cfg.ServerConfigs[0])
	}()

	time.Sleep(1 * time.Second)

	t.Run("test read from server", func(t *testing.T) {
		hc := config.HostConfig{
			Host: "127.0.0.1",
			Port: "9000",
		}
		d, err := dialTarget(hc)
		if err != nil {
			t.Fatal(err)
		}

		buf := make([]byte, 5)
		_, err = d.Read(buf)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				t.Fatalf("got %v, want %v", err, io.EOF)
			}
		}
	})

	// shutdown octo-proxy
	p.Shutdown()
}

func TestUnreachableMirror(t *testing.T) {
	var wg sync.WaitGroup
	result := make(chan []byte)

	// start target server
	backend := testhelper.RunTestServer(&wg, result)

	// prepare octo proxy configuration
	cfg, err := config.GenerateConfig("127.0.0.1:9000", []string{backend}, "")
	if err != nil {
		t.Fatal(err)
	}
	cfg.ServerConfigs[0].Mirror = config.HostConfig{
		Host: "127.0.0.1",
		Port: "10",
	}

	p := New("test-proxy")
	go func() {
		p.Run(cfg.ServerConfigs[0])
	}()

	time.Sleep(1 * time.Second)

	// send data to octo-proxy
	if err := SendData(cfg.ServerConfigs[0].Listener, messageByte, false); err != nil {
		t.Fatal(err)
	}

	// check data received by test server
	// the data must be available even the mirror can't be reached
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

func TestProxyConcurrent(t *testing.T) {
	var wg sync.WaitGroup
	connCount := 50

	// start target server
	backend := testhelper.RunTestServerWithResponse(&wg, connCount)

	// start octo proxy
	cfg, err := config.GenerateConfig("127.0.0.1:9000", []string{backend}, "")
	if err != nil {
		t.Fatal(err)
	}
	p := New("test-proxy")
	go func() {
		p.Run(cfg.ServerConfigs[0])
	}()

	time.Sleep(1 * time.Second)

	for i := 0; i < connCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			t.Run("test read from server", func(t *testing.T) {
				hc := config.HostConfig{
					Host: "127.0.0.1",
					Port: "9000",
				}
				d, err := dialTarget(hc)
				if err != nil {
					t.Error(err)
				}

				buf := make([]byte, 6)
				_, err = d.Read(buf)
				if err != nil {
					t.Error(err)
				}

				if string(buf) != "holaaa" {
					t.Errorf("got %v, want holaaa", string(buf))
				}
			})
		}()
	}

	wg.Wait()
	// shutdown octo-proxy
	p.Shutdown()
}

func TestProxyWithSlowTarget(t *testing.T) {
	var wg sync.WaitGroup
	// start target server
	backend := testhelper.RunTestServerSlowMode(&wg, 1)

	// start octo proxy
	cfg, err := config.GenerateConfig("127.0.0.1:9000", []string{backend}, "")
	if err != nil {
		t.Fatal(err)
	}
	// set timeout for target
	cfg.ServerConfigs[0].Targets[0].TimeoutDuration = 3 * time.Second

	p := New("test-proxy")
	go func() {
		p.Run(cfg.ServerConfigs[0])
	}()

	time.Sleep(1 * time.Second)

	t.Run("test read from server and test if timeout is work", func(t *testing.T) {
		hc := config.HostConfig{
			Host: "127.0.0.1",
			Port: "9000",
		}
		d, err := dialTarget(hc)
		if err != nil {
			t.Error(err)
		}

		buf := make([]byte, 26)
		b := make([]byte, 1)

		for i := 0; i < 26; i++ {
			n, err := d.Read(b)
			if err != nil {
				if err != io.EOF {
					log.Println(err)
				}

				if errors.Is(err, syscall.EPIPE) {
					log.Println(err)
					break
				}
			}

			log.Println(string(b[:n]))

			copy(buf[i:], b[:n])
		}

		r := bytes.Compare(buf, []byte("lorem ipsum dolor sit amet"))
		if r != -1 {
			t.Fatalf("the response length must match")
		}
	})

	wg.Wait()
	// shutdown octo-proxy
	p.Shutdown()
}
