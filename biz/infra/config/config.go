package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

type DatabaseConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	Address  string `yaml:"address"`
}

type Config struct {
	DatabaseConfig DatabaseConfig `yaml:"database"`
}

func InitConfig(path string) *Config {
	conf := &Config{}
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	err = yaml.NewDecoder(f).Decode(conf)
	if err != nil {
		panic(err)
	}

	return conf
}
