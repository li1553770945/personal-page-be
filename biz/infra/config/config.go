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
	Port     int    `yaml:"port"`
}

type HttpConfig struct {
	Address   string `yaml:"address"`
	SecretKey string `yaml:"secret_key"`
}

type Config struct {
	DatabaseConfig DatabaseConfig `yaml:"database"`
	HttpConfig     HttpConfig     `yaml:"http"`
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
