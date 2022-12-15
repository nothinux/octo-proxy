package proxy

import (
	"crypto/tls"
	"log"
	"sync"

	"github.com/nothinux/octo-proxy/pkg/config"
	"github.com/nothinux/octo-proxy/pkg/tlsconn"
)

func SendData(hc config.HostConfig, message []byte, readResponse bool) error {
	d, err := dialTarget(hc)
	if err != nil {
		log.Println(err)
		return err
	}

	_, err = d.Write(message)
	if err != nil {
		log.Println(err)
		return err
	}

	if readResponse {
		buf := make([]byte, 5)
		_, err = d.Read(buf)
		if err != nil {
			log.Println(err)
			return err
		}
	}

	d.Close()

	return nil
}

func RunTestTLSServer(wg *sync.WaitGroup, c config.TLSConfig, result chan []byte) string {
	tlsConfig, err := tlsconn.GetTLSConfig(c)
	if err != nil {
		log.Fatal(err)
	}

	l, err := tls.Listen("tcp", "127.0.0.1:", tlsConfig.Config)
	if err != nil {
		log.Fatal(err)
	}

	wg.Add(1)
	go func(wg *sync.WaitGroup, r chan []byte) {
		defer wg.Done()

		c, err := l.Accept()
		if err != nil {
			return
		}

		buf := make([]byte, 5)
		_, err = c.Read(buf)
		if err != nil {
			log.Println(err)
		}

		r <- buf
	}(wg, result)

	return l.Addr().String()
}
