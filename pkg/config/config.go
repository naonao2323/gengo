package config

type Config interface {
	GetSchema() string
	GetDbUrl() string
	GetParallel() int
	GetInclude() []string
}
