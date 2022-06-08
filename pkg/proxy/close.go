package proxy

import "net"

func closeConn(conn []net.Conn) {
	for i := 0; i < len(conn); i++ {
		if conn[i] != nil {
			conn[i].Close()
		}
	}
}
