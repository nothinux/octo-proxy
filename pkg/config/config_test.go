package config

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

var (
	validConfig = `servers:
- name: proxy-1
  listener:
    host: 127.0.0.1
    port: 8080
  targets:
    - host: 127.0.0.1
      port: 80`

	invalidConfig = `server{}`
)

func TestNewConfig(t *testing.T) {
	t.Run("Test when configuration file is not exists", func(t *testing.T) {
		_, err := New("/tmp/conf.yaml")
		if err != nil {
			if !strings.Contains(err.Error(), "open /tmp/conf.yaml: no such file or directory") {
				t.Fatal(err)
			}
		}
	})

	t.Run("Test valid configuration file", func(t *testing.T) {
		_, err := New("../testdata/test-config.yaml")
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestTlsConfigMode(t *testing.T) {
	tests := []struct {
		Name               string
		Config             TLSConfig
		expectedMutualMode bool
		expectedSimpleMode bool
	}{
		{
			Name: "Test if mode is mutual",
			Config: TLSConfig{
				Mode: "mutual",
			},
			expectedMutualMode: true,
			expectedSimpleMode: false,
		},
		{
			Name: "Test if mode is simple",
			Config: TLSConfig{
				Mode: "simple",
			},
			expectedMutualMode: false,
			expectedSimpleMode: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			if tt.Config.IsMutual() != tt.expectedMutualMode {
				t.Fatalf("got %v, want %v", tt.expectedMutualMode, tt.Config.IsMutual())
			}

			if tt.Config.IsSimple() != tt.expectedSimpleMode {
				t.Fatalf("got %v, want %v", tt.expectedSimpleMode, tt.Config.IsMutual())
			}
		})
	}
}

func TestHostConfigType(t *testing.T) {
	tests := []struct {
		Name           string
		Hct            hostConfigType
		expectedString string
	}{
		{
			Name:           "Test if hostconfigtype is listener",
			Hct:            slistener,
			expectedString: "listener",
		},
		{
			Name:           "Test if hostconfigtype is target",
			Hct:            starget,
			expectedString: "target",
		},
		{
			Name:           "Test if hostconfigtype is mirror",
			Hct:            smirror,
			expectedString: "mirror",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			if tt.Hct.String() != tt.expectedString {
				t.Fatalf("got %s, want %s", tt.Hct.String(), tt.expectedString)
			}
		})
	}
}

func TestGenerateConfig(t *testing.T) {
	tests := []struct {
		Name           string
		Listener       string
		Targets        []string
		Metrics        string
		expectedConfig *Config
		expectedError  string
	}{
		{
			Name:     "Test valid listener and target",
			Listener: "127.0.0.1:8080",
			Targets:  []string{"127.0.0.1:80"},
			Metrics:  "127.0.0.1:9123",
			expectedConfig: &Config{
				ServerConfigs: []ServerConfig{
					{
						Name: "default",
						Listener: HostConfig{
							Host: "127.0.0.1",
							Port: "8080",
							ConnectionConfig: ConnectionConfig{
								TimeoutDuration: 300 * time.Second,
							},
						},
						Targets: []HostConfig{
							{
								Host: "127.0.0.1",
								Port: "80",
								ConnectionConfig: ConnectionConfig{
									TimeoutDuration: 300 * time.Second,
								},
							},
						},
					},
				},
				MetricsConfig: HostConfig{
					Host: "127.0.0.1",
					Port: "9123",
				},
			},
		},
		{
			Name:           "Test invalid listener",
			Listener:       ":8080",
			Targets:        []string{"127.0.0.1:80"},
			expectedConfig: nil,
			expectedError:  "host in servers.[0].listener.host not specified",
		},
		{
			Name:           "Test invalid target",
			Listener:       "127.0.0.1:8080",
			Targets:        []string{":80"},
			expectedConfig: nil,
			expectedError:  "host in servers.[0].target.host not specified",
		},
		{
			Name:           "Test invalid port in listener",
			Listener:       "127.0.0.1:m",
			Targets:        []string{"127.0.0.1:80"},
			expectedConfig: nil,
			expectedError:  "port in servers.[0].listener.port is not valid",
		},
		{
			Name:           "Test no port in listener",
			Listener:       "127.0.0.1:",
			Targets:        []string{"127.0.0.1:80"},
			expectedConfig: nil,
			expectedError:  "port in servers.[0].listener.port not specified",
		},
		{
			Name:           "Test no port in target",
			Listener:       "127.0.0.1:8080",
			Targets:        []string{"127.0.0.1"},
			expectedConfig: nil,
			expectedError:  "target must be specified in format host:port",
		},
		{
			Name:           "Test multiple",
			Listener:       "127.0.0.1:8080",
			Targets:        []string{"127.0.0.1:80:8080"},
			expectedConfig: nil,
			expectedError:  "target must be specified in format host:port",
		},
		{
			Name:           "Test invalid metrics server 1",
			Listener:       "127.0.0.1:8080",
			Targets:        []string{"127.0.0.1:8080"},
			Metrics:        "localhost",
			expectedConfig: nil,
			expectedError:  "metrics server address must be specified in format host:port",
		},
		{
			Name:           "Test invalid metrics server 2",
			Listener:       "127.0.0.1:8080",
			Targets:        []string{"127.0.0.1:8080"},
			Metrics:        "localhost:80:8080",
			expectedConfig: nil,
			expectedError:  "metrics server address must be specified in format host:port",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			c, err := GenerateConfig(tt.Listener, tt.Targets, tt.Metrics)
			if err != nil {
				if !strings.Contains(err.Error(), tt.expectedError) {
					t.Fatalf("got %v, want %s", err, tt.expectedError)
				}
			}

			if !reflect.DeepEqual(c, tt.expectedConfig) {
				t.Fatalf("got %v, want %v", c, tt.expectedConfig)
			}
		})
	}
}

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
				ServerConfigs: []ServerConfig{
					{
						Name: "proxy-1",
						Listener: HostConfig{
							Host: "127.0.0.1",
							Port: "8080",
						},
						Targets: []HostConfig{
							{
								Host: "127.0.0.1",
								Port: "80",
							},
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
				ServerConfigs: []ServerConfig{
					{
						Name: "proxy-1",
						Listener: HostConfig{
							Host: "127.0.0.1",
							Port: "8080",
						},
						Targets: []HostConfig{
							{
								Host: "127.0.0.1",
								Port: "80",
							},
						},
					},
				},
			},
			expectedConfig: &Config{
				ServerConfigs: []ServerConfig{
					{
						Name: "proxy-1",
						Listener: HostConfig{
							Host: "127.0.0.1",
							Port: "8080",
							ConnectionConfig: ConnectionConfig{
								TimeoutDuration: 300 * time.Second,
							},
						},
						Targets: []HostConfig{
							{
								Host: "127.0.0.1",
								Port: "80",
								ConnectionConfig: ConnectionConfig{
									TimeoutDuration: 300 * time.Second,
								},
							},
						},
					},
				},
			},
		},
		{
			Name: "no server config",
			Config: &Config{
				ServerConfigs: []ServerConfig{},
			},
			expectedConfig: nil,
			expectedError:  "error no server configuration found",
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
				ServerConfigs: []ServerConfig{
					{
						Name: "proxy-1",
						Targets: []HostConfig{
							{
								Host: "127.0.0.1",
								Port: "80",
							},
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
				ServerConfigs: []ServerConfig{
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
			expectedError:  "no target configurations in servers.[0]",
		},
		{
			Name: "no host in listener",
			Config: &Config{
				ServerConfigs: []ServerConfig{
					{
						Name: "proxy-1",
						Listener: HostConfig{
							Port: "8080",
						},
						Targets: []HostConfig{
							{
								Host: "127.0.0.1",
								Port: "80",
							},
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
				ServerConfigs: []ServerConfig{
					{
						Name: "proxy-1",
						Listener: HostConfig{
							Host: "local.local",
							Port: "8080",
						},
						Targets: []HostConfig{
							{
								Host: "127.0.0.1",
								Port: "80",
							},
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
				ServerConfigs: []ServerConfig{
					{
						Name: "proxy-1",
						Listener: HostConfig{
							Host: "127.0.0.1",
						},
						Targets: []HostConfig{
							{
								Host: "127.0.0.1",
								Port: "80",
							},
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
				ServerConfigs: []ServerConfig{
					{
						Name: "proxy-1",
						Listener: HostConfig{
							Host: "127.0.0.1",
							Port: "8o",
						},
						Targets: []HostConfig{
							{
								Host: "127.0.0.1",
								Port: "80",
							},
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
				ServerConfigs: []ServerConfig{
					{
						Name: "proxy-1",
						Listener: HostConfig{
							Host: "127.0.0.1",
							Port: "8080",
						},
						Targets: []HostConfig{
							{
								Port: "80",
							},
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
				ServerConfigs: []ServerConfig{
					{
						Name: "proxy-1",
						Listener: HostConfig{
							Host: "local.local",
							Port: "8080",
						},
						Targets: []HostConfig{
							{
								Host: "127.0.0.1",
								Port: "80",
							},
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
				ServerConfigs: []ServerConfig{
					{
						Name: "proxy-1",
						Listener: HostConfig{
							Host: "127.0.0.1",
							Port: "8080",
						},
						Targets: []HostConfig{
							{
								Host: "127.0.0.1",
								Port: "loc",
							},
						},
					},
				},
			},
			expectedConfig: nil,
			expectedError:  "port in servers.[0].target.port is not valid port number",
		},
		{
			Name: "no host in mirror",
			Config: &Config{
				ServerConfigs: []ServerConfig{
					{
						Name: "proxy-1",
						Listener: HostConfig{
							Host: "127.0.0.1",
							Port: "8080",
						},
						Targets: []HostConfig{
							{
								Host: "127.0.0.1",
								Port: "80",
							},
						},
						Mirror: HostConfig{
							Port: "2210",
						},
					},
				},
			},
			expectedConfig: nil,
			expectedError:  "host in servers.[0].mirror.host not specified",
		},
		{
			Name: "no port in mirror",
			Config: &Config{
				ServerConfigs: []ServerConfig{
					{
						Name: "proxy-1",
						Listener: HostConfig{
							Host: "127.0.0.1",
							Port: "8080",
						},
						Targets: []HostConfig{
							{
								Host: "127.0.0.1",
								Port: "80",
							},
						},
						Mirror: HostConfig{
							Host: "192.168.1.1",
						},
					},
				},
			},
			expectedConfig: nil,
			expectedError:  "port in servers.[0].mirror.port not specified",
		},
		{
			Name: "port in mirror is invalid",
			Config: &Config{
				ServerConfigs: []ServerConfig{
					{
						Name: "proxy-1",
						Listener: HostConfig{
							Host: "127.0.0.1",
							Port: "8080",
						},
						Targets: []HostConfig{
							{
								Host: "127.0.0.1",
								Port: "80",
							},
						},
						Mirror: HostConfig{
							Host: "192.168.1.1",
							Port: "2cc",
						},
					},
				},
			},
			expectedConfig: nil,
			expectedError:  "port in servers.[0].mirror.port is not valid",
		},
		{
			Name: "set timeout on listener",
			Config: &Config{
				ServerConfigs: []ServerConfig{
					{
						Name: "proxy-1",
						Listener: HostConfig{
							Host: "127.0.0.1",
							Port: "8080",
							ConnectionConfig: ConnectionConfig{
								Timeout: "10s",
							},
						},
						Targets: []HostConfig{
							{
								Host: "127.0.0.1",
								Port: "80",
							},
						},
					},
				},
			},
			expectedConfig: &Config{
				ServerConfigs: []ServerConfig{
					{
						Name: "proxy-1",
						Listener: HostConfig{
							Host: "127.0.0.1",
							Port: "8080",
							ConnectionConfig: ConnectionConfig{
								Timeout:         "10s",
								TimeoutDuration: 10 * time.Second,
							},
						},
						Targets: []HostConfig{
							{
								Host: "127.0.0.1",
								Port: "80",
								ConnectionConfig: ConnectionConfig{
									TimeoutDuration: 300 * time.Second,
								},
							},
						},
					},
				},
			},
		},
		{
			Name: "set timeout on listener using ms",
			Config: &Config{
				ServerConfigs: []ServerConfig{
					{
						Name: "proxy-1",
						Listener: HostConfig{
							Host: "127.0.0.1",
							Port: "8080",
							ConnectionConfig: ConnectionConfig{
								Timeout: "10ms",
							},
						},
						Targets: []HostConfig{
							{
								Host: "127.0.0.1",
								Port: "80",
							},
						},
					},
				},
			},
			expectedConfig: &Config{
				ServerConfigs: []ServerConfig{
					{
						Name: "proxy-1",
						Listener: HostConfig{
							Host: "127.0.0.1",
							Port: "8080",
							ConnectionConfig: ConnectionConfig{
								Timeout:         "10ms",
								TimeoutDuration: 10 * time.Millisecond,
							},
						},
						Targets: []HostConfig{
							{
								Host: "127.0.0.1",
								Port: "80",
								ConnectionConfig: ConnectionConfig{
									TimeoutDuration: 300 * time.Second,
								},
							},
						},
					},
				},
			},
		},
		{
			Name: "set zero timeout on listener & target",
			Config: &Config{
				ServerConfigs: []ServerConfig{
					{
						Name: "proxy-1",
						Listener: HostConfig{
							Host: "127.0.0.1",
							Port: "8080",
							ConnectionConfig: ConnectionConfig{
								Timeout: "0",
							},
						},
						Targets: []HostConfig{
							{
								Host: "127.0.0.1",
								Port: "80",
								ConnectionConfig: ConnectionConfig{
									Timeout: "0",
								},
							},
						},
					},
				},
			},
			expectedConfig: &Config{
				ServerConfigs: []ServerConfig{
					{
						Name: "proxy-1",
						Listener: HostConfig{
							Host: "127.0.0.1",
							Port: "8080",
							ConnectionConfig: ConnectionConfig{
								Timeout:         "0",
								TimeoutDuration: 0,
							},
						},
						Targets: []HostConfig{
							{
								Host: "127.0.0.1",
								Port: "80",
								ConnectionConfig: ConnectionConfig{
									Timeout:         "0",
									TimeoutDuration: 0,
								},
							},
						},
					},
				},
			},
		},
		{
			Name: "set invalid timeout on listener",
			Config: &Config{
				ServerConfigs: []ServerConfig{
					{
						Name: "proxy-1",
						Listener: HostConfig{
							Host: "127.0.0.1",
							Port: "8080",
							ConnectionConfig: ConnectionConfig{
								Timeout: "foo",
							},
						},
						Targets: []HostConfig{
							{
								Host: "127.0.0.1",
								Port: "80",
							},
						},
					},
				},
			},
			expectedConfig: nil,
			expectedError:  "failed to parse timeout servers.[0]: strconv.Atoi: parsing \"foo\": invalid syntax",
		},
		{
			Name: "set invalid timeout on target",
			Config: &Config{
				ServerConfigs: []ServerConfig{
					{
						Name: "proxy-1",
						Listener: HostConfig{
							Host: "127.0.0.1",
							Port: "8080",
						},
						Targets: []HostConfig{
							{
								Host: "127.0.0.1",
								Port: "80",
								ConnectionConfig: ConnectionConfig{
									Timeout: "foo",
								},
							},
						},
					},
				},
			},
			expectedConfig: nil,
			expectedError:  "failed to parse timeout servers.[0].targets[0]: strconv.Atoi: parsing \"foo\": invalid syntax",
		},
		{
			Name: "set invalid timeout on mirror",
			Config: &Config{
				ServerConfigs: []ServerConfig{
					{
						Name: "proxy-1",
						Listener: HostConfig{
							Host: "127.0.0.1",
							Port: "8080",
						},
						Targets: []HostConfig{
							{
								Host: "127.0.0.1",
								Port: "80",
							},
						},
						Mirror: HostConfig{
							Host: "127.0.0.1",
							Port: "81",
							ConnectionConfig: ConnectionConfig{
								Timeout: "-1",
							},
						},
					},
				},
			},
			expectedConfig: nil,
			expectedError:  "failed to parse timeout servers.[0].mirror: can't use negative value for timeout",
		},
		{
			Name: "set timeout on listener, target and set mirror to default",
			Config: &Config{
				ServerConfigs: []ServerConfig{
					{
						Name: "proxy-1",
						Listener: HostConfig{
							Host: "127.0.0.1",
							Port: "8080",
						},
						Targets: []HostConfig{
							{
								Host: "127.0.0.1",
								Port: "80",
							},
						},
						Mirror: HostConfig{
							Host: "127.0.0.1",
							Port: "9999",
						},
					},
				},
			},
			expectedConfig: &Config{
				ServerConfigs: []ServerConfig{
					{
						Name: "proxy-1",
						Listener: HostConfig{
							Host: "127.0.0.1",
							Port: "8080",
							ConnectionConfig: ConnectionConfig{
								TimeoutDuration: 300 * time.Second,
							},
						},
						Targets: []HostConfig{
							{
								Host: "127.0.0.1",
								Port: "80",
								ConnectionConfig: ConnectionConfig{
									TimeoutDuration: 300 * time.Second,
								},
							},
						},
						Mirror: HostConfig{
							Host: "127.0.0.1",
							Port: "9999",
							ConnectionConfig: ConnectionConfig{
								TimeoutDuration: 300 * time.Second,
							},
						},
					},
				},
			},
		},
		{
			Name: "valid tlsConfig on listener",
			Config: &Config{
				ServerConfigs: []ServerConfig{
					{
						Name: "proxy-1",
						Listener: HostConfig{
							Host: "127.0.0.1",
							Port: "8080",
							TLSConfig: TLSConfig{
								Mode:   "mutual",
								Key:    "/tmp/key.pem",
								CaCert: "/tmp/ca-cert.pem",
								Cert:   "/tmp/cert.pem",
							},
						},
						Targets: []HostConfig{
							{
								Host: "127.0.0.1",
								Port: "80",
							},
						},
					},
				},
			},
			expectedConfig: &Config{
				ServerConfigs: []ServerConfig{
					{
						Name: "proxy-1",
						Listener: HostConfig{
							Host: "127.0.0.1",
							Port: "8080",
							ConnectionConfig: ConnectionConfig{
								TimeoutDuration: 300 * time.Second,
							},
							TLSConfig: TLSConfig{
								Role: Role{
									Server: true,
								},
								Mode:   "mutual",
								Key:    "/tmp/key.pem",
								CaCert: "/tmp/ca-cert.pem",
								Cert:   "/tmp/cert.pem",
							},
						},
						Targets: []HostConfig{
							{
								Host: "127.0.0.1",
								Port: "80",
								ConnectionConfig: ConnectionConfig{
									TimeoutDuration: 300 * time.Second,
								},
							},
						},
					},
				},
			},
		},
		{
			Name: "tlsConfig mode is mutual but no cert provided",
			Config: &Config{
				ServerConfigs: []ServerConfig{
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
						Targets: []HostConfig{
							{
								Host: "127.0.0.1",
								Port: "80",
							},
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
				ServerConfigs: []ServerConfig{
					{
						Name: "proxy-1",
						Listener: HostConfig{
							Host: "127.0.0.1",
							Port: "8080",
						},
						Targets: []HostConfig{
							{
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
			},
			expectedConfig: nil,
			expectedError:  "cacert, cert and key in servers.[0].target must be set if mode is mutual",
		},
		{
			Name: "caCert, cert, and key is set, but mode is not set",
			Config: &Config{
				ServerConfigs: []ServerConfig{
					{
						Name: "proxy-1",
						Listener: HostConfig{
							Host: "127.0.0.1",
							Port: "8080",
						},
						Targets: []HostConfig{
							{
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
			},
			expectedConfig: nil,
			expectedError:  "ignore cacert, cert or key in servers.[0].target because tlsConfig.mode is not set",
		},
		{
			Name: "check if mode is not mutual or simple",
			Config: &Config{
				ServerConfigs: []ServerConfig{
					{
						Name: "proxy-1",
						Listener: HostConfig{
							Host: "127.0.0.1",
							Port: "8080",
						},
						Targets: []HostConfig{
							{
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
			},
			expectedConfig: nil,
			expectedError:  "not supported mode in servers.[0].target",
		},
		{
			Name: "check if mode is nil",
			Config: &Config{
				ServerConfigs: []ServerConfig{
					{
						Name: "proxy-1",
						Listener: HostConfig{
							Host: "127.0.0.1",
							Port: "8080",
						},
						Targets: []HostConfig{
							{
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
			},
			expectedConfig: nil,
			expectedError:  "ignore cacert, cert or key in servers.[0].target because tlsConfig.mode is not set",
		},
		{
			Name: "check if host in metrics not specified",
			Config: &Config{
				ServerConfigs: []ServerConfig{
					{
						Name: "proxy-1",
						Listener: HostConfig{
							Host: "127.0.0.1",
							Port: "8080",
						},
						Targets: []HostConfig{
							{
								Host: "127.0.0.1",
								Port: "80",
							},
						},
					},
				},
				MetricsConfig: HostConfig{
					Port: "8080",
				},
			},
			expectedConfig: nil,
			expectedError:  "metrics.host not specified",
		},
		{
			Name: "check if port in metrics not specified",
			Config: &Config{
				ServerConfigs: []ServerConfig{
					{
						Name: "proxy-1",
						Listener: HostConfig{
							Host: "127.0.0.1",
							Port: "8080",
						},
						Targets: []HostConfig{
							{
								Host: "127.0.0.1",
								Port: "80",
							},
						},
					},
				},
				MetricsConfig: HostConfig{
					Host: "127.0.0.1",
				},
			},
			expectedConfig: nil,
			expectedError:  "metrics.port not specified",
		},
		{
			Name: "check if port in metrics is same with port that defined in listener",
			Config: &Config{
				ServerConfigs: []ServerConfig{
					{
						Name: "proxy-1",
						Listener: HostConfig{
							Host: "127.0.0.1",
							Port: "8080",
						},
						Targets: []HostConfig{
							{
								Host: "127.0.0.1",
								Port: "80",
							},
						},
					},
				},
				MetricsConfig: HostConfig{
					Host: "127.0.0.1",
					Port: "8080",
				},
			},
			expectedConfig: nil,
			expectedError:  "can't bind to port that already used by listener",
		},
		{
			Name: "check if metrics configuration is valid",
			Config: &Config{
				ServerConfigs: []ServerConfig{
					{
						Name: "proxy-1",
						Listener: HostConfig{
							Host: "127.0.0.1",
							Port: "8080",
						},
						Targets: []HostConfig{
							{
								Host: "127.0.0.1",
								Port: "80",
							},
						},
					},
				},
				MetricsConfig: HostConfig{
					Host: "127.0.0.1",
					Port: "9123",
				},
			},
			expectedConfig: &Config{
				ServerConfigs: []ServerConfig{
					{
						Name: "proxy-1",
						Listener: HostConfig{
							Host: "127.0.0.1",
							Port: "8080",
							ConnectionConfig: ConnectionConfig{
								TimeoutDuration: 300 * time.Second,
							},
						},
						Targets: []HostConfig{
							{
								Host: "127.0.0.1",
								Port: "80",
								ConnectionConfig: ConnectionConfig{
									TimeoutDuration: 300 * time.Second,
								},
							},
						},
					},
				},
				MetricsConfig: HostConfig{
					Host: "127.0.0.1",
					Port: "9123",
				},
			},
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
				t.Fatalf("\ngot %v, \nwant %v", config, tt.expectedConfig)
			}
		})
	}
}
