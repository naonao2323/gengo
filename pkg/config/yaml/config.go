package yaml

import (
	"os"
	"runtime"

	"gopkg.in/yaml.v3"

	ConfigParser "github.com/naonao2323/testgen/pkg/config"
)

type configImpl struct {
	Schema   string    `yaml:"schema"`
	DbUrl    string    `yaml:"dbUrl"`
	Parallel *int      `yaml:"parallel"`
	Include  *[]string `yaml:"include"`
}

type Deploy = int

const (
	FrameWork Deploy = iota
	Dao
	Fixture
	Container
)

func NewConfig(path string) (ConfigParser.Config, error) {
	config, err := parseConfig(path)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func parseConfig(path string) (*configImpl, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config configImpl
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (c configImpl) GetInclude() *[]string {
	return c.Include
}

func (c configImpl) GetSchema() string {
	return c.Schema
}

func (c configImpl) GetDbUrl() string {
	return c.DbUrl
}

func (c configImpl) GetParallel() int {
	if c.Parallel == nil {
		return runtime.NumCPU()
	}
	return *c.Parallel
}
