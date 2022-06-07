package config

import (
	"reflect"
	"strings"
	"testing"
)

var (
	validConfig = `servers:
- name: proxy-1
  listener:
    host: 127.0.0.1
    port: 8080
  target:
    host: 127.0.0.1
    port: 80`

	invalidConfig = `server{}`
)

func TestReadConfig(t *testing.T) {
	tests := []struct {
		Name           string
		Config         string
		expectedConfig *Config
		expectedError  string
	}{
		{
			Name:   "valid yaml file",
			Config: validConfig,
			expectedConfig: &Config{
				[]ServerConfig{
					{
						Name: "proxy-1",
						Listener: HostConfig{
							Host: "127.0.0.1",
							Port: "8080",
						},
						Target: HostConfig{
							Host: "127.0.0.1",
							Port: "80",
						},
					},
				},
			},
		},
		{
			Name:           "invalid yaml file",
			Config:         invalidConfig,
			expectedConfig: nil,
			expectedError:  "cannot unmarshal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			config, err := readConfig(strings.NewReader(tt.Config))
			if err != nil {
				if !strings.Contains(err.Error(), tt.expectedError) {
					t.Fatalf("got %v, want error contains %s", err, tt.expectedError)
				}
			}

			if !reflect.DeepEqual(config, tt.expectedConfig) {
				t.Fatalf("got %v, want %v", config, tt.expectedConfig)
			}
		})
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		Name           string
		Config         *Config
		expectedConfig *Config
		expectedError  string
	}{
		{
			Name: "valid config",
			Config: &Config{
				[]ServerConfig{
					{
						Name: "proxy-1",
						Listener: HostConfig{
							Host: "127.0.0.1",
							Port: "8080",
						},
						Target: HostConfig{
							Host: "127.0.0.1",
							Port: "80",
						},
					},
				},
			},
			expectedConfig: &Config{
				[]ServerConfig{
					{
						Name: "proxy-1",
						Listener: HostConfig{
							Host:    "127.0.0.1",
							Port:    "8080",
							Timeout: 300,
							TLSConfig: TLSConfig{
								Role: Role{
									Server: true,
								},
							},
						},
						Target: HostConfig{
							Host:    "127.0.0.1",
							Port:    "80",
							Timeout: 300,
						},
					},
				},
			},
		},
		{
			Name:           "no config",
			Config:         nil,
			expectedConfig: nil,
			expectedError:  "error no configuration found",
		},
		{
			Name: "no listener config",
			Config: &Config{
				[]ServerConfig{
					{
						Name: "proxy-1",
						Target: HostConfig{
							Host: "127.0.0.1",
							Port: "80",
						},
					},
				},
			},
			expectedConfig: nil,
			expectedError:  "no listener configuration in servers.[0]",
		},
		{
			Name: "no target config",
			Config: &Config{
				[]ServerConfig{
					{
						Name: "proxy-1",
						Listener: HostConfig{
							Host: "127.0.0.1",
							Port: "80",
						},
					},
				},
			},
			expectedConfig: nil,
			expectedError:  "no target configuration in servers.[0]",
		},
		{
			Name: "no host in listener",
			Config: &Config{
				[]ServerConfig{
					{
						Name: "proxy-1",
						Listener: HostConfig{
							Port: "8080",
						},
						Target: HostConfig{
							Host: "127.0.0.1",
							Port: "80",
						},
					},
				},
			},
			expectedConfig: nil,
			expectedError:  "host in servers.[0].listener.host not specified",
		},
		{
			Name: "invalid host in listener",
			Config: &Config{
				[]ServerConfig{
					{
						Name: "proxy-1",
						Listener: HostConfig{
							Host: "local.local",
							Port: "8080",
						},
						Target: HostConfig{
							Host: "127.0.0.1",
							Port: "80",
						},
					},
				},
			},
			expectedConfig: nil,
			expectedError:  "host in servers.[0].listener.host is not valid ip address",
		},
		{
			Name: "no port in listener",
			Config: &Config{
				[]ServerConfig{
					{
						Name: "proxy-1",
						Listener: HostConfig{
							Host: "127.0.0.1",
						},
						Target: HostConfig{
							Host: "127.0.0.1",
							Port: "80",
						},
					},
				},
			},
			expectedConfig: nil,
			expectedError:  "port in servers.[0].listener.port not specified",
		},
		{
			Name: "invalid port in listener",
			Config: &Config{
				[]ServerConfig{
					{
						Name: "proxy-1",
						Listener: HostConfig{
							Host: "127.0.0.1",
							Port: "8o",
						},
						Target: HostConfig{
							Host: "127.0.0.1",
							Port: "80",
						},
					},
				},
			},
			expectedConfig: nil,
			expectedError:  "port in servers.[0].listener.port is not valid port number",
		},
		{
			Name: "no host in target",
			Config: &Config{
				[]ServerConfig{
					{
						Name: "proxy-1",
						Listener: HostConfig{
							Host: "127.0.0.1",
							Port: "8080",
						},
						Target: HostConfig{
							Port: "80",
						},
					},
				},
			},
			expectedConfig: nil,
			expectedError:  "host in servers.[0].target.host not specified",
		},
		{
			Name: "invalid host in target",
			Config: &Config{
				[]ServerConfig{
					{
						Name: "proxy-1",
						Listener: HostConfig{
							Host: "local.local",
							Port: "8080",
						},
						Target: HostConfig{
							Host: "127.0.0.1",
							Port: "80",
						},
					},
				},
			},
			expectedConfig: nil,
			expectedError:  "host in servers.[0].listener.host is not valid ip address",
		},
		{
			Name: "invalid port in target",
			Config: &Config{
				[]ServerConfig{
					{
						Name: "proxy-1",
						Listener: HostConfig{
							Host: "127.0.0.1",
							Port: "8080",
						},
						Target: HostConfig{
							Host: "127.0.0.1",
							Port: "loc",
						},
					},
				},
			},
			expectedConfig: nil,
			expectedError:  "port in servers.[0].target.port is not valid port number",
		},
		{
			Name: "set timeout on listener",
			Config: &Config{
				[]ServerConfig{
					{
						Name: "proxy-1",
						Listener: HostConfig{
							Host:    "127.0.0.1",
							Port:    "8080",
							Timeout: 10,
						},
						Target: HostConfig{
							Host: "127.0.0.1",
							Port: "80",
						},
					},
				},
			},
			expectedConfig: &Config{
				[]ServerConfig{
					{
						Name: "proxy-1",
						Listener: HostConfig{
							Host:    "127.0.0.1",
							Port:    "8080",
							Timeout: 10,
							TLSConfig: TLSConfig{
								Role: Role{
									Server: true,
								},
							},
						},
						Target: HostConfig{
							Host:    "127.0.0.1",
							Port:    "80",
							Timeout: 300,
						},
					},
				},
			},
		},
		{
			Name: "set timeout on listener and target",
			Config: &Config{
				[]ServerConfig{
					{
						Name: "proxy-1",
						Listener: HostConfig{
							Host:    "127.0.0.1",
							Port:    "8080",
							Timeout: 10,
						},
						Target: HostConfig{
							Host:    "127.0.0.1",
							Port:    "80",
							Timeout: 200,
						},
					},
				},
			},
			expectedConfig: &Config{
				[]ServerConfig{
					{
						Name: "proxy-1",
						Listener: HostConfig{
							Host:    "127.0.0.1",
							Port:    "8080",
							Timeout: 10,
							TLSConfig: TLSConfig{
								Role: Role{
									Server: true,
								},
							},
						},
						Target: HostConfig{
							Host:    "127.0.0.1",
							Port:    "80",
							Timeout: 200,
						},
					},
				},
			},
		},
		{
			Name: "tlsConfig mode is mutual but no cert provided",
			Config: &Config{
				[]ServerConfig{
					{
						Name: "proxy-1",
						Listener: HostConfig{
							Host: "127.0.0.1",
							Port: "8080",
							TLSConfig: TLSConfig{
								Mode:   "mutual",
								Key:    "/tmp/key.pem",
								CaCert: "/tmp/ca-cert.pem",
							},
						},
						Target: HostConfig{
							Host: "127.0.0.1",
							Port: "80",
						},
					},
				},
			},
			expectedConfig: nil,
			expectedError:  "cacert, cert and key in servers.[0].listener must be set if mode is mutual",
		},
		{
			Name: "tlsConfig mode is mutual but no key provided",
			Config: &Config{
				[]ServerConfig{
					{
						Name: "proxy-1",
						Listener: HostConfig{
							Host: "127.0.0.1",
							Port: "8080",
						},
						Target: HostConfig{
							Host: "127.0.0.1",
							Port: "80",
							TLSConfig: TLSConfig{
								Mode:   "mutual",
								Cert:   "/tmp/cert.pem",
								CaCert: "/tmp/ca-cert.pem",
							},
						},
					},
				},
			},
			expectedConfig: nil,
			expectedError:  "cacert, cert and key in servers.[0].target must be set if mode is mutual",
		},
		{
			Name: "caCert, cert, and key is set, but mode is not set",
			Config: &Config{
				[]ServerConfig{
					{
						Name: "proxy-1",
						Listener: HostConfig{
							Host: "127.0.0.1",
							Port: "8080",
						},
						Target: HostConfig{
							Host: "127.0.0.1",
							Port: "80",
							TLSConfig: TLSConfig{
								CaCert: "/tmp/ca-cert.pem",
								Cert:   "/tmp/cert.pem",
								Key:    "/tmp/key.pem",
							},
						},
					},
				},
			},
			expectedConfig: nil,
			expectedError:  "ignore cacert, cert or key in servers.[0].target because tlsConfig.mode is not set",
		},
		{
			Name: "check if mode is not mutual or simple",
			Config: &Config{
				[]ServerConfig{
					{
						Name: "proxy-1",
						Listener: HostConfig{
							Host: "127.0.0.1",
							Port: "8080",
						},
						Target: HostConfig{
							Host: "127.0.0.1",
							Port: "80",
							TLSConfig: TLSConfig{
								CaCert: "/tmp/ca-cert.pem",
								Cert:   "/tmp/cert.pem",
								Key:    "/tmp/key.pem",
								Mode:   "server",
							},
						},
					},
				},
			},
			expectedConfig: nil,
			expectedError:  "not supported mode in servers.[0].target",
		},
		{
			Name: "check if mode is nil",
			Config: &Config{
				[]ServerConfig{
					{
						Name: "proxy-1",
						Listener: HostConfig{
							Host: "127.0.0.1",
							Port: "8080",
						},
						Target: HostConfig{
							Host: "127.0.0.1",
							Port: "80",
							TLSConfig: TLSConfig{
								CaCert: "/tmp/ca-cert.pem",
								Cert:   "/tmp/cert.pem",
								Key:    "/tmp/key.pem",
								Mode:   "",
							},
						},
					},
				},
			},
			expectedConfig: nil,
			expectedError:  "ignore cacert, cert or key in servers.[0].target because tlsConfig.mode is not set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			config, err := validateConfig(tt.Config)
			if err != nil {
				if !strings.Contains(err.Error(), tt.expectedError) {
					t.Fatalf("got %v, want %v", err, tt.expectedError)
				}
			}

			if !reflect.DeepEqual(config, tt.expectedConfig) {
				t.Fatalf("got %v, want %v", config, tt.expectedConfig)
			}
		})
	}
}
