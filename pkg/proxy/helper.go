package proxy

import (
	"net"

	goerrors "errors"
)

func errCopy(err error) error {
	if err != nil {
		if goerrors.Is(err, net.ErrClosed) {
			// use of closed network connection
		} else {
			return err
		}
	}

	return nil
}
