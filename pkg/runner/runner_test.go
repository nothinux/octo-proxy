package runner

import (
	"bytes"
	"log"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/nothinux/octo-proxy/pkg/config"
	"github.com/nothinux/octo-proxy/pkg/testhelper"
)

func TestRunningRunner(t *testing.T) {
	// start backend server
	var wg sync.WaitGroup
	result := make(chan []byte)
	backend := testhelper.RunTestServer(&wg, result)

	// initialize config
	c, err := config.New("../testdata/run-config.yaml")
	if err != nil {
		t.Fatal(err)
	}
	// replace port in target with backend create previously
	c.ServerConfigs[0].Targets[0].Port = strings.Split(backend, ":")[1]

	// starting octo proxy
	ss := &Server{ServerConfigs: c.ServerConfigs}
	proxies := ss.runProxy()

	// store created proxy data
	octo := &Octo{
		Proxies: proxies,
	}

	// wait till proxy running
	time.Sleep(1 * time.Second)

	// trying to send data
	if err := SendData(
		net.JoinHostPort(c.ServerConfigs[0].Listener.Host, c.ServerConfigs[0].Listener.Port),
		[]byte("hello"),
	); err != nil {
		t.Fatal(err)
	}

	t.Run("test backend server receive data", func(t *testing.T) {
		res := <-result

		r := bytes.Compare(res, []byte("hello"))
		if r != 0 {
			t.Fatalf("got %v, want %v", res, []byte("hello"))
		}
	})

	t.Run("test proxy name is set", func(t *testing.T) {
		if octo.Proxies["test-server"].Name != "test-server" {
			t.Fatalf("got %v, want test-server", octo.Proxies["test-server"].Name)
		}
	})

	// trying to reload with invalid configuration
	if err := reloadProxy("../testdata/err-config.yaml", octo); err != nil {
		if !strings.Contains(err.Error(), "host in servers.[0].target.host not specified") {
			t.Fatalf("reload must be error")
		}
	}

	if err := reloadProxy("../testdata/rel-config.yaml", octo); err != nil {
		if err != nil {
			t.Fatal(err)
		}
	}

	// wait till listener up
	time.Sleep(1 * time.Second)

	// shutdown octo-proxy
	shutdown(octo.Proxies, nil)

}

func TestRunMetrics(t *testing.T) {
	tests := []struct {
		Name   string
		Config config.HostConfig
	}{
		{
			Name: "Test metrics without port",
			Config: config.HostConfig{
				Host: "127.0.0.1",
			},
		},
		{
			Name: "Test metrics",
			Config: config.HostConfig{
				Host: "127.0.0.1",
				Port: "9125",
			},
		},
	}

	for _, tt := range tests {
		m, err := runMetrics(tt.Config)
		if err != nil {
			t.Fatal(err)
		}

		time.Sleep(1 * time.Second)
		shutdown(nil, m)
	}
}

func SendData(address string, message []byte) error {
	d, err := net.DialTimeout("tcp", address, 3*time.Second)
	if err != nil {
		log.Println(err)
		return err
	}

	_, err = d.Write(message)
	if err != nil {
		log.Println(err)
		return err
	}
	d.Close()

	return nil
}
