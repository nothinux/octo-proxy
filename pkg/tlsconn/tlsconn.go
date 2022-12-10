package tlsconn

import (
	"crypto/tls"
	"crypto/x509"
	"os"
	"strings"

	"github.com/nothinux/octo-proxy/pkg/config"
	"github.com/nothinux/octo-proxy/pkg/errors"
	"github.com/rs/zerolog/log"
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

// GetTLSConfig returns a TLS Config for the simple and mutual tls
func GetTLSConfig(c config.TLSConfig) (*ProxyTLS, error) {
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

				log.Debug().
					Int("n-certificates", len(cs.PeerCertificates)).
					Msg("verifying peer certificates")

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
	cacert, err := os.ReadFile(c.CaCert)
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
