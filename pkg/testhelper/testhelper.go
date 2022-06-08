package testhelper

import (
	"log"
	"net"
	"sync"
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
