package yaml

import (
	"os"

	"gopkg.in/yaml.v3"
)

type configImpl struct {
	Schema string `yaml:"schema"`
	DBURL  string `yaml:"dbUrl"`
}

type Config interface {
	GetSchema() string
	GetDBURL() string
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

func (c configImpl) GetSchema() string {
	return c.Schema
}

func (c configImpl) GetDBURL() string {
	return c.DBURL
}
