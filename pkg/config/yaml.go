package config

import (
	"os"
	"runtime"

	"gopkg.in/yaml.v3"
)

type yamlConfig struct {
	Schema   string    `yaml:"schema"`
	DbUrl    string    `yaml:"dbUrl"`
	Parallel *int      `yaml:"parallel"`
	Include  *[]string `yaml:"include"`
	Writer   string    `yaml:"writer"`
}

func parseYamlConfig(path string) (*yamlConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config yamlConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (c yamlConfig) getInclude() *[]string {
	return c.Include
}

func (c yamlConfig) getSchema() string {
	return c.Schema
}

func (c yamlConfig) getDbUrl() string {
	return c.DbUrl
}

func (c yamlConfig) getParallel() int {
	if c.Parallel == nil {
		return runtime.NumCPU()
	}
	return *c.Parallel
}

func (c yamlConfig) getWriter() string {
	return c.Writer
}
