package proxy

import (
	"crypto/tls"
	"crypto/x509"
	"io"
	"net"
	"os"
	"strings"

	goerrors "errors"

	"github.com/nothinux/octo-proxy/pkg/config"
	"github.com/nothinux/octo-proxy/pkg/errors"
)

type ProxyTLS struct {
	*tls.Config
}

func newTLS() *ProxyTLS {
	return &ProxyTLS{
		Config: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}
}

func getTLSConfig(c config.TLSConfig) (*ProxyTLS, error) {
	var caPool *x509.CertPool
	var pair tls.Certificate
	var err error

	ptls := newTLS()

	if c.CaCert != "" {
		caPool, err = getCACertPool(c)
		if err != nil {
			return nil, err
		}

		ptls.RootCAs = caPool
	}

	if c.Cert != "" && c.Key != "" {
		pair, err = getCertKeyPair(c)
		if err != nil {
			return nil, err
		}

		ptls.Certificates = []tls.Certificate{pair}
	}

	if c.IsMutual() {
		if c.Role.Server {
			ptls.GetConfigForClient = func(h *tls.ClientHelloInfo) (*tls.Config, error) {
				return &tls.Config{
					ClientCAs:    caPool,
					Certificates: []tls.Certificate{pair},
					ClientAuth:   tls.RequireAndVerifyClientCert,
					VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
						opts := x509.VerifyOptions{
							Roots:   caPool,
							DNSName: strings.Split(h.Conn.RemoteAddr().String(), ":")[0],
						}

						_, err := verifiedChains[0][0].Verify(opts)

						return err
					},
				}, nil
			}
		} else {
			ptls.VerifyConnection = func(cs tls.ConnectionState) error {
				opts := x509.VerifyOptions{
					Roots:         caPool,
					DNSName:       cs.ServerName,
					Intermediates: x509.NewCertPool(),
				}

				for _, cert := range cs.PeerCertificates[1:] {
					opts.Intermediates.AddCert(cert)
				}

				_, err := cs.PeerCertificates[0].Verify(opts)

				return err
			}
		}
	}

	return ptls, nil
}

func getCACertPool(c config.TLSConfig) (*x509.CertPool, error) {
	cacert, err := readContent(c.CaCert)
	if err != nil {
		return nil, err
	}

	pool := x509.NewCertPool()
	if ok := pool.AppendCertsFromPEM(cacert); !ok {
		return nil, errors.New("tlsConfig", "can't add CA to pool "+err.Error())
	}

	return pool, nil
}

func getCertKeyPair(c config.TLSConfig) (tls.Certificate, error) {
	pcert, err := readContent(c.Cert)
	if err != nil {
		return tls.Certificate{}, err
	}

	pkey, err := readContent(c.Key)
	if err != nil {
		return tls.Certificate{}, err
	}

	cert, err := tls.X509KeyPair(pcert, pkey)
	if err != nil {
		return tls.Certificate{}, errors.New("tlsConfig", "can't parse public & private key pair "+err.Error())
	}

	return cert, nil
}

func isTLSConn(nc net.Conn) error {
	if _, ok := nc.(*tls.Conn); ok {
		if err := nc.(*tls.Conn).Handshake(); err != nil {
			return err
		}

		if !nc.(*tls.Conn).ConnectionState().HandshakeComplete {
			nc.Close()
			return goerrors.New("handshake failed")
		}
	}

	return nil
}

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
