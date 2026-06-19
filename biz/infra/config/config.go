package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type DatabaseConfig struct {
	Driver   string `yaml:"driver"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	Address  string `yaml:"address"`
	Port     int    `yaml:"port"`
	UseTLS   bool   `yaml:"use-tls"`
	SSLMode  string `yaml:"sslmode"`
}

type HttpConfig struct {
	Address   string `yaml:"address"`
	SecretKey string `yaml:"secret_key"`
	JWTKey    string `yaml:"jwt_key"`
}

type CosConfig struct {
	Ak       string `yaml:"ak"`
	Sk       string `yaml:"sk"`
	Endpoint string `yaml:"endpoint"`
}

type AIChatConfig struct {
	Model  string `yaml:"model"`
	APIKey string `yaml:"api-key"`
	URL    string `yaml:"url"`
}

type NotifyConfig struct {
	ServerChanKey string `yaml:"server-chan-key"`
}

type Config struct {
	DatabaseConfig DatabaseConfig `yaml:"database"`
	HttpConfig     HttpConfig     `yaml:"http"`
	CosConfig      CosConfig      `yaml:"cos"`
	AIChatConfig   AIChatConfig   `yaml:"aichat"`
	NotifyConfig   NotifyConfig   `yaml:"notify"`
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

func (c *Config) EffectiveJWTKey() string {
	if v := os.Getenv("PERSONAL_PAGE_JWT_KEY"); v != "" {
		return v
	}
	if c.HttpConfig.JWTKey != "" {
		return c.HttpConfig.JWTKey
	}
	if c.HttpConfig.SecretKey != "" {
		return c.HttpConfig.SecretKey
	}
	return "secret"
}

func (c *Config) EffectiveNotifyKey() string {
	if v := os.Getenv("SERVER_CHAN_KEY"); v != "" {
		return v
	}
	if c.NotifyConfig.ServerChanKey != "" {
		return c.NotifyConfig.ServerChanKey
	}
	return c.HttpConfig.SecretKey
}

func (c *Config) EffectiveCOSAk() string {
	if v := os.Getenv("TENCENT_COS_SECRET_ID"); v != "" {
		return v
	}
	return c.CosConfig.Ak
}

func (c *Config) EffectiveCOSSk() string {
	if v := os.Getenv("TENCENT_COS_SECRET_KEY"); v != "" {
		return v
	}
	return c.CosConfig.Sk
}

func (c *Config) EffectiveCOSEndpoint() string {
	if v := os.Getenv("TENCENT_COS_ENDPOINT"); v != "" {
		return v
	}
	return c.CosConfig.Endpoint
}

func (c *Config) EffectiveAIChatAPIKey() string {
	if v := os.Getenv("DIFY_API_KEY"); v != "" {
		return v
	}
	if v := os.Getenv("AICHAT_API_KEY"); v != "" {
		return v
	}
	return c.AIChatConfig.APIKey
}

func (c *Config) EffectiveAIChatURL() string {
	if v := os.Getenv("DIFY_API_URL"); v != "" {
		return v
	}
	if c.AIChatConfig.URL != "" {
		return c.AIChatConfig.URL
	}
	return "https://api.dify.ai/v1/chat-messages"
}
