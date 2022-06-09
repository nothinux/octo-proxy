package testhelper

import (
	"errors"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"syscall"
	"time"
)

// RunTestServer run tcp server on random port for testing purpose
func RunTestServer(wg *sync.WaitGroup, result chan []byte) string {
	l, err := net.Listen("tcp", "127.0.0.1:")
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

func RunTestServerWithResponse(wg *sync.WaitGroup, connCount int) string {
	l, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < connCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			c, err := l.Accept()
			if err != nil {
				log.Println(err)
			}

			_, err = c.Write([]byte("holaaa"))
			if err != nil {
				log.Println(err)
			}

			c.Close()
		}()
	}

	return l.Addr().String()
}

func RunTestServerSlowMode(wg *sync.WaitGroup, connCount int) string {
	l, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		log.Fatal(err)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < connCount; i++ {

			c, err := l.Accept()
			if err != nil {
				log.Println(err)
			}

			message := strings.NewReader("lorem ipsum dolor sit amet")

			for i := 0; i < 26; i++ {
				_, err := io.CopyN(c, message, 1)
				if err != nil {
					if errors.Is(err, syscall.EPIPE) {
						log.Println(err)
						break
					}
					log.Println(err)
				}
				time.Sleep(1 * time.Second)
			}

			c.Close()
		}
	}()

	return l.Addr().String()
}
