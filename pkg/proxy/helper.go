package proxy

import (
	"io"
	"net"
	"os"

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

func readContent(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	b, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return b, nil
}
