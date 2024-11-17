package yaml

import (
	"os"

	"gopkg.in/yaml.v3"
)

type configImpl struct {
	Schema   string `yaml:"schema"`
	DBURL    string `yaml:"dbUrl"`
	Provider int    `yaml:"provider"`
}

type Provider int

const (
	Mysql Provider = iota
	Postgres
)

type Config interface {
	GetSchema() string
	GetDBURL() string
	GetProvider() Provider
}

func NewConfig(path string) (Config, error) {
	config, err := parseConfig(path)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func parseConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config configImpl
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func (c configImpl) GetDBURL() string {
	return c.DBURL
}

func (c configImpl) GetSchema() string {
	return c.Schema
}

func (c configImpl) GetProvider() Provider {
	switch c.Provider {
	case 0:
		return Mysql
	case 1:
		return Postgres
	default:
		return -1
	}
}
