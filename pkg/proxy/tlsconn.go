// TODO: separate this usecase to new package
package proxy

import (
	"crypto/tls"
	"crypto/x509"
	goerrors "errors"
	"net"
	"strings"
	"time"

	"github.com/nothinux/octo-proxy/pkg/config"
	"github.com/rs/zerolog/log"
)

type ProxyTLS struct {
	*tls.Config
	RevocationList *x509.RevocationList
}

type VerifyOpts struct {
	CaCert *x509.Certificate
	Cert   *x509.Certificate
	CRL    *x509.RevocationList
}

func newTLS() *ProxyTLS {
	return &ProxyTLS{
		Config: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}
}

// getTLSConfig returns a TLS Config for the simple and mutual tls
func getTLSConfig(c config.TLSConfig) (*ProxyTLS, error) {
	var caPool *x509.CertPool
	var caCert *x509.Certificate
	var crlVerification bool
	var pair tls.Certificate
	var err error

	ptls := newTLS()
	if c.SNI != "" {
		ptls.ServerName = c.SNI
	}

	if c.CaCert != "" {
		caPool, err = getCACertPool(c)
		if err != nil {
			return nil, err
		}

		ptls.RootCAs = caPool

		caCert, err = getCACertificate(c)
		if err != nil {
			return nil, err
		}
	}

	if c.Cert != "" && c.Key != "" {
		pair, err = getCertKeyPair(c)
		if err != nil {
			return nil, err
		}

		ptls.Certificates = []tls.Certificate{pair}
	}

	if c.CRL != "" {
		crl, err := getCACRL(c)
		if err != nil {
			return nil, err
		}

		ptls.RevocationList = crl
		crlVerification = true
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
						if err != nil {
							return err
						}

						if crlVerification {
							revocationOpts := VerifyOpts{
								CaCert: caCert,
								Cert:   verifiedChains[0][0],
								CRL:    ptls.RevocationList,
							}

							_, err = isCertificateRevoked(revocationOpts)
							return err
						}

						return nil
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
				if err != nil {
					return err
				}

				if crlVerification {
					revocationOpts := VerifyOpts{
						CaCert: caCert,
						Cert:   cs.PeerCertificates[0],
						CRL:    ptls.RevocationList,
					}

					_, err := isCertificateRevoked(revocationOpts)
					return err
				}

				return nil
			}
		}
	}

	return ptls, nil
}

func isCertificateRevoked(opts VerifyOpts) (bool, error) {
	err := opts.CRL.CheckSignatureFrom(opts.CaCert)
	if err != nil {
		return false, err
	}

	if opts.CRL.NextUpdate.Before(time.Now()) {
		return false, goerrors.New("crl file is outdated")
	}

	for _, sn := range opts.CRL.RevokedCertificateEntries {
		if sn.SerialNumber.Cmp(opts.Cert.SerialNumber) == 0 {
			return false, goerrors.New("certificate was revoked and no longer valid - CN:" + opts.Cert.Subject.CommonName)
		}
	}

	return true, nil
}

func isTLSConn(nc net.Conn) error {
	if _, ok := nc.(*tls.Conn); ok {
		if err := nc.(*tls.Conn).Handshake(); err != nil {
			return err
		}

		if !nc.(*tls.Conn).ConnectionState().HandshakeComplete {
			return goerrors.New("handshake failed")
		}
	}

	return nil
}
