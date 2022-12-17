package config

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/nothinux/octo-proxy/pkg/errors"

	"gopkg.in/yaml.v2"
)

type hostConfigType int

const (
	slistener hostConfigType = iota
	starget
	smirror
	smetrics
)

type Config struct {
	ServerConfigs []ServerConfig `yaml:"servers"`
	MetricsConfig HostConfig     `yaml:"metrics"`
}

type ServerConfig struct {
	Name     string       `yaml:"name"`
	Listener HostConfig   `yaml:"listener"`
	Targets  []HostConfig `yaml:"targets"`
	Mirror   HostConfig   `yaml:"mirror"`
}

type HostConfig struct {
	Host            string `yaml:"host"`
	Port            string `yaml:"port"`
	Timeout         string `yaml:"timeout"`
	TimeoutDuration time.Duration
	TLSConfig       `yaml:"tlsConfig"`
}

type TLSConfig struct {
	CaCert          string   `yaml:"caCert"`
	Cert            string   `yaml:"cert"`
	Key             string   `yaml:"key"`
	Mode            string   `yaml:"mode"`
	SubjectAltNames []string `yaml:"subjectAltNames"`
	SubjectAltName
	Role
}

type SubjectAltName struct {
	IPAddress []string
	Uri       []string
	DNS       []string
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
	return [...]string{"listener", "target", "mirror", "metrics"}[h]
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

func GenerateConfig(listener string, targets []string, metrics string) (*Config, error) {
	l := strings.Split(listener, ":")

	if len(l) != 2 {
		return nil, errors.New("error", "listener must be specified in format host:port")
	}

	c := &Config{
		ServerConfigs: []ServerConfig{
			{
				Name: "default",
				Listener: HostConfig{
					Host: l[0],
					Port: l[1],
				},
				Targets: []HostConfig{},
			},
		},
	}

	for _, target := range targets {
		t := strings.Split(target, ":")
		if len(t) != 2 {
			return nil, errors.New("error", "target must be specified in format host:port")
		}

		hc := HostConfig{
			Host: t[0],
			Port: t[1],
		}

		c.ServerConfigs[0].Targets = append(c.ServerConfigs[0].Targets, hc)
	}

	if len(metrics) > 0 {
		t := strings.Split(metrics, ":")
		if len(t) != 2 {
			return nil, errors.New("error", "metrics server address must be specified in format host:port")
		}

		c.MetricsConfig = HostConfig{
			Host: t[0],
			Port: t[1],
		}
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

	for i := range c.ServerConfigs {
		listener := &c.ServerConfigs[i].Listener
		mirror := &c.ServerConfigs[i].Mirror

		if err := errorCheck(i, slistener, listener); err != nil {
			return nil, err
		}

		if len(c.ServerConfigs[i].Targets) == 0 {
			return nil, errors.New("server", fmt.Sprintf("no target configurations in servers.[%d]", i))
		}

		for j := range c.ServerConfigs[i].Targets {
			if err := errorCheck(i, starget, &c.ServerConfigs[i].Targets[j]); err != nil {
				return nil, err
			}

			if err := setTimeout(&c.ServerConfigs[i].Targets[j]); err != nil {
				return nil, errors.New("server", fmt.Sprintf("failed to parse timeout servers.[%d].targets[%d]: %v", i, j, err))
			}

			setSAN(&c.ServerConfigs[i].Targets[j])
		}

		// check config for error only when configuration is not nil
		if !reflect.DeepEqual(HostConfig{}, *mirror) {
			if err := errorCheck(i, smirror, mirror); err != nil {
				return nil, err
			}

			if err := setTimeout(mirror); err != nil {
				return nil, errors.New("server", fmt.Sprintf("failed to parse timeout servers.[%d].mirror: %v", i, err))
			}

			setSAN(mirror)
		}
		// set all listener role to server
		if !reflect.DeepEqual(TLSConfig{}, c.ServerConfigs[i].Listener.TLSConfig) {
			listener.TLSConfig.Role.Server = true
		}

		if err := setTimeout(listener); err != nil {
			return nil, errors.New("server", fmt.Sprintf("failed to parse timeout servers.[%d]: %v", i, err))
		}

		setSAN(listener)
	}

	if !reflect.DeepEqual(HostConfig{}, c.MetricsConfig) {
		if err := errorCheck(0, smetrics, &c.MetricsConfig); err != nil {
			// TODO: handle error in errorcheck
			return nil, errors.New("metrics", strings.TrimLeft(err.Error(), "[server] host in servers.[0]."))
		}
	}

	return c, nil
}

func setTimeout(c *HostConfig) error {
	seconds := 300

	if c.Timeout != "" {
		var err error
		seconds, err = strconv.Atoi(c.Timeout)
		if err != nil {
			return err
		}
	}

	c.TimeoutDuration = time.Duration(seconds) * time.Second

	return nil
}

func setSAN(c *HostConfig) {
	if len(c.TLSConfig.SubjectAltNames) != 0 {
		c.TLSConfig.SubjectAltName = *parseSubjectAltNames(c.TLSConfig.SubjectAltNames)
	}
}

func errorCheck(i int, hct hostConfigType, c *HostConfig) error {
	if reflect.DeepEqual(HostConfig{}, *c) {
		return errors.New("server", fmt.Sprintf("no %s configuration in servers.[%d]", hct.String(), i))
	}

	if c.Host == "" {
		return errors.New("server", fmt.Sprintf("host in servers.[%d].%s.host not specified", i, hct.String()))
	}

	if hct != slistener {
		if !hostIsValid(c.Host) {
			return errors.New("server", fmt.Sprintf("host in servers.[%d].%s.host is not valid ip address", i, hct.String()))
		}
	} else {
		if !hostIPIsValid(c.Host) {
			return errors.New("server", fmt.Sprintf("host in servers.[%d].%s.host is not valid ip address", i, hct.String()))
		}
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
