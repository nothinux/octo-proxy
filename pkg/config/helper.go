package config

import (
	"errors"
	"io"
	"net"
	"os"
	"strconv"
)

func hostIPIsValid(h string) bool {
	return net.ParseIP(h) != nil
}

func portIsValid(p string) bool {
	valid := true

	port, err := strconv.Atoi(p)
	if err != nil {
		valid = false
	}

	if port >= 1 && port <= 65535 {
	} else {
		valid = false
	}

	return valid
}

func isExist(path string) bool {
	_, err := os.Stat(path)

	return !errors.Is(err, os.ErrNotExist)
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
