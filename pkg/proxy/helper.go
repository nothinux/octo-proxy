package proxy

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"net"
	"os"

	goerrors "errors"

	"github.com/nothinux/octo-proxy/pkg/config"
	"github.com/nothinux/octo-proxy/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

func errCopy(err error, tConf config.HostConfig) error {
	if err != nil {
		if goerrors.Is(err, net.ErrClosed) {
			// use of closed network connection
		} else {
			upstreamConnErr.With(prometheus.Labels{"host": tConf.Host, "port": tConf.Port}).Inc()
			return err
		}
	}

	return nil
}

func timeoutIsZero(c config.HostConfig) bool {
	return c.ConnectionConfig.TimeoutDuration == 0
}

func getCACertPool(c config.TLSConfig) (*x509.CertPool, error) {
	cacert, err := os.ReadFile(c.CaCert)
	if err != nil {
		return nil, err
	}

	pool := x509.NewCertPool()
	if ok := pool.AppendCertsFromPEM(cacert); !ok {
		return nil, errors.New("tlsConfig", "can't add CA to pool")
	}

	return pool, nil
}

func getCACertificate(c config.TLSConfig) (*x509.Certificate, error) {
	cacert, err := os.ReadFile(c.CaCert)
	if err != nil {
		return nil, err
	}

	b, _ := pem.Decode(cacert)
	if b == nil {
		return nil, goerrors.New("error: can't decode CA file")
	}

	return x509.ParseCertificate(b.Bytes)
}

func getCACRL(c config.TLSConfig) (*x509.RevocationList, error) {
	rawCRL, err := os.ReadFile(c.CRL)
	if err != nil {
		return nil, err
	}

	b, _ := pem.Decode(rawCRL)
	if b == nil {
		return nil, goerrors.New("error: can't decode CRL file")
	}

	return x509.ParseRevocationList(b.Bytes)
}

func getCertKeyPair(c config.TLSConfig) (tls.Certificate, error) {
	pcert, err := os.ReadFile(c.Cert)
	if err != nil {
		return tls.Certificate{}, err
	}

	pkey, err := os.ReadFile(c.Key)
	if err != nil {
		return tls.Certificate{}, err
	}

	cert, err := tls.X509KeyPair(pcert, pkey)
	if err != nil {
		return tls.Certificate{}, errors.New("tlsConfig", "can't parse public & private key pair "+err.Error())
	}

	return cert, nil
}
