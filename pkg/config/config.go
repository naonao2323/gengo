package config

import (
	"errors"
)

type Config interface {
	GetSchema() string
	GetDbUrl() string
	GetParallel() int
	GetInclude() *[]string
	GetWriter() Writer
}

type config struct {
	schema   string
	dbUrl    string
	parallel int
	include  *[]string
	writer   string
}

type Writer = int

const (
	File Writer = iota
	Sdout
	Unknown
)

type Deploy = int

const (
	FrameWork Deploy = iota
	Dao
	Fixture
	Container
)

type Format int

const (
	Yaml Format = iota
	// Json
	UnDefined Format = -1
)

func NewConfig(format Format, path string) (Config, error) {
	switch format {
	case Yaml:
		yaml, err := parseYamlConfig(path)
		if err != nil {
			return nil, err
		}
		conf := config{
			schema:   yaml.GetSchema(),
			dbUrl:    yaml.GetDbUrl(),
			parallel: yaml.GetParallel(),
			include:  yaml.GetInclude(),
			writer:   yaml.GetWriter(),
		}
		return conf, nil
	default:
		return nil, errors.New("undefined error")
	}
}

func (c config) GetInclude() *[]string {
	return c.include
}

func (c config) GetSchema() string {
	return c.schema
}

func (c config) GetDbUrl() string {
	return c.dbUrl
}

func (c config) GetParallel() int {
	return c.parallel
}

func (c config) GetWriter() Writer {
	switch c.writer {
	case "file":
		return File
	case "sdout":
		return Sdout
	default:
		return Unknown
	}
}
