package main

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

type Config struct {
	SSLCertificateKey   string      `yaml:"ssl_certificate_key"`
	Location            []*Location `yaml:"location"`
	Schema              string      `yaml:"schema"`
	Port                int         `yaml:"port"`
	SSLCertificate      string      `yaml:"ssl_certificate"`
	HealthCheck         bool        `yaml:"health_check"`
	HealthCheckInterval uint        `yaml:"health_check_interval"`
	MaxAllowed          uint        `yaml:"max_allowed"`
}

type Location struct {
	Pattern     string   `yaml:"pattern"`
	ProxyPass   []string `yaml:"proxy_pass"`
	BalanceMode string   `yaml:"balance_mode"`
}

func ReadConfig(filePath string) (*Config, error) {
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	config := &Config{}
	err2 := yaml.Unmarshal(file, config)
	if err2 != nil {
		return nil, err
	}
	return config, nil
}

func (l *Location) String() string {
	var res string
	res += fmt.Sprintf("\t- Pattern: %v\n", l.Pattern)
	res += fmt.Sprintf("\t  ProxyPass: \n")
	for _, proxy := range l.ProxyPass {
		res += fmt.Sprintf("\t\t%v\n", proxy)
	}
	res += fmt.Sprintf("\t  BalanceMode: %v\n", l.BalanceMode)
	return res
}

func (c *Config) String() string {
	var res string
	res += fmt.Sprintf("Schema: %v\n", c.Schema)
	res += fmt.Sprintf("Port: %v\n", c.Port)
	res += fmt.Sprintf("SSLCertificate: %v\n", c.SSLCertificate)
	res += fmt.Sprintf("SSLCertificateKey: %v\n", c.SSLCertificateKey)
	res += fmt.Sprintf("HealthCheck: %v\n", c.HealthCheck)
	res += fmt.Sprintf("HealthCheckInterval: %v\n", c.HealthCheckInterval)
	res += fmt.Sprintf("MaxAllowed: %v\n", c.MaxAllowed)
	res += fmt.Sprintf("Location: \n")
	for _, l := range c.Location {
		res += fmt.Sprintf("%v", l)
	}
	return res
}

func (c *Config) simpleValidation() error {
	if c.Schema != "http" && c.Schema != "https" {
		return fmt.Errorf("not support the schema: %v", c.Schema)
	}
	if len(c.Location) == 0 {
		return fmt.Errorf("the location cannot be null")
	}
	if c.Schema == "https" && (len(c.SSLCertificate) == 0 || len(c.SSLCertificateKey) == 0) {
		return fmt.Errorf("the https proxy requires ssl_certificate_key and ssl_certificate")
	}
	if c.HealthCheckInterval < 1 {
		return errors.New("health_check_interval must be greater than 0")
	}
	return nil
}
