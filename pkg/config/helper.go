package config

import (
	"io"
	"net"
	"os"
	"regexp"
	"strconv"
)

var ipRegex = regexp.MustCompile("(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])")

func hostIsValid(h string) bool {
	if ipRegex.MatchString(h) {
		return hostIPIsValid(h)
	}

	// if not match with regex, then assume its a hostname
	return true
}

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

func readContent(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	return readContentFile(f)
}

func readContentFile(r io.Reader) ([]byte, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return b, nil
}
