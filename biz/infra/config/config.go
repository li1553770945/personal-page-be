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

type TencentConfig struct {
	APPID     int    `yaml:"app_id"`
	SecretKey string `yaml:"key"`
}

type Config struct {
	DatabaseConfig DatabaseConfig `yaml:"database"`
	HttpConfig     HttpConfig     `yaml:"http"`
	TencentConfig  TencentConfig  `yaml:"tencent"`
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
