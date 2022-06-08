package config

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/nothinux/octo-proxy/pkg/errors"

	"gopkg.in/yaml.v2"
)

type hostConfigType int

const (
	slistener hostConfigType = iota
	starget
	smirror
)

type Config struct {
	ServerConfigs []ServerConfig `yaml:"servers"`
}

type ServerConfig struct {
	Name     string     `yaml:"name"`
	Listener HostConfig `yaml:"listener"`
	Target   HostConfig `yaml:"target"`
	Mirror   HostConfig `yaml:"mirror"`
}

type HostConfig struct {
	Host      string `yaml:"host"`
	Port      string `yaml:"port"`
	Timeout   int    `yaml:"timeout"`
	TLSConfig `yaml:"tlsConfig"`
}

type TLSConfig struct {
	CaCert string `yaml:"caCert"`
	Cert   string `yaml:"cert"`
	Key    string `yaml:"key"`
	Mode   string `yaml:"mode"`
	Role
}

type Role struct {
	Server bool
}

func (t TLSConfig) IsMutual() bool {
	return t.Mode == "mutual"
}

func (t TLSConfig) IsSimple() bool {
	return t.Mode == "simple"
}

func (h hostConfigType) String() string {
	return [...]string{"listener", "target", "mirror"}[h]
}

func New(configPath string) (*Config, error) {
	c, err := openConfig(configPath)
	if err != nil {
		return nil, err
	}

	return validateConfig(c)
}

func openConfig(configPath string) (*Config, error) {
	f, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return readConfig(f)
}

func readConfig(r io.Reader) (*Config, error) {
	config := &Config{}

	d := yaml.NewDecoder(r)
	if err := d.Decode(config); err != nil {
		return nil, err
	}

	return config, nil
}

func GenerateConfig(listener, target string) (*Config, error) {
	l := strings.Split(listener, ":")
	t := strings.Split(target, ":")

	if len(l) != 2 || len(t) != 2 {
		return nil, errors.New("error", "listener or target must be specified in format host:port")
	}

	c := &Config{
		ServerConfigs: []ServerConfig{
			{
				Name: "default",
				Listener: HostConfig{
					Host: l[0],
					Port: l[1],
				},
				Target: HostConfig{
					Host: t[0],
					Port: t[1],
				},
			},
		},
	}

	return validateConfig(c)
}

func validateConfig(c *Config) (*Config, error) {
	if c == nil {
		return nil, errors.New("config", "error no configuration found")
	}

	if len(c.ServerConfigs) == 0 {
		return nil, errors.New("servers", "error no server configuration found")
	}

	for i := 0; i < len(c.ServerConfigs); i++ {
		listener := &c.ServerConfigs[i].Listener
		target := &c.ServerConfigs[i].Target
		mirror := &c.ServerConfigs[i].Mirror

		if err := errorCheck(i, slistener, listener); err != nil {
			return nil, err
		}

		if err := errorCheck(i, starget, target); err != nil {
			return nil, err
		}

		// check config for error only when configuration is not nil
		if (HostConfig{}) != *mirror {
			if err := errorCheck(i, smirror, mirror); err != nil {
				return nil, err
			}

			setTimeout(mirror)
		}
		// set all listener role to server
		if (TLSConfig{}) != c.ServerConfigs[i].Listener.TLSConfig {
			listener.TLSConfig.Role.Server = true
		}

		setTimeout(listener)
		setTimeout(target)
	}

	return c, nil
}

func setTimeout(c *HostConfig) {
	if c.Timeout == 0 {
		c.Timeout = 300
	}
}

func errorCheck(i int, hct hostConfigType, c *HostConfig) error {
	if (HostConfig{}) == *c {
		return errors.New("server", fmt.Sprintf("no %s configuration in servers.[%d]", hct.String(), i))
	}

	if c.Host == "" {
		return errors.New("server", fmt.Sprintf("host in servers.[%d].%s.host not specified", i, hct.String()))
	}

	if !hostIPIsValid(c.Host) {
		return errors.New("server", fmt.Sprintf("host in servers.[%d].%s.host is not valid ip address", i, hct.String()))
	}

	if c.Port == "" {
		return errors.New("server", fmt.Sprintf("port in servers.[%d].%s.port not specified", i, hct.String()))
	}

	if !portIsValid(c.Port) {
		return errors.New("server", fmt.Sprintf("port in servers.[%d].%s.port is not valid port number", i, hct.String()))
	}

	// inform user when cacert, cert and key provided but
	if c.TLSConfig.CaCert != "" || c.TLSConfig.Cert != "" || c.TLSConfig.Key != "" {
		if c.TLSConfig.Mode == "" {
			return errors.New("server.tlsConfig", fmt.Sprintf("ignore cacert, cert or key in servers.[%d].%s because tlsConfig.mode is not set", i, hct.String()))
		}

		if !c.TLSConfig.IsMutual() && !c.TLSConfig.IsSimple() {
			return errors.New("server.tlsConfig", fmt.Sprintf("not supported mode in servers.[%d].%s", i, hct.String()))
		}
	}

	if c.TLSConfig.IsMutual() {
		// if mode is mutual cert, key and ca cert must be set
		if c.TLSConfig.CaCert == "" || c.TLSConfig.Cert == "" || c.TLSConfig.Key == "" {
			return errors.New("server.tlsConfig", fmt.Sprintf("cacert, cert and key in servers.[%d].%s must be set if mode is mutual", i, hct.String()))
		}
	}

	return nil
}
