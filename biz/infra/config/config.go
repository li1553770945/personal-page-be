package config

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

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
	Address    string `yaml:"address"`
	SecretKey  string `yaml:"secret_key"`
	JWTKey     string `yaml:"jwt_key"`
	SessionKey string `yaml:"session_key"`
}

type CosConfig struct {
	Ak       string `yaml:"ak"`
	Sk       string `yaml:"sk"`
	Endpoint string `yaml:"endpoint"`
}

type AIChatConfig struct {
	Model               string `yaml:"model"`
	APIKey              string `yaml:"api-key"`
	URL                 string `yaml:"url"`
	MaxRequestBodyBytes int    `yaml:"max-request-body-bytes"`
	MaxMessageRunes     int    `yaml:"max-message-runes"`
	RequestsPerMinute   int    `yaml:"requests-per-minute"`
	MaxConcurrent       int    `yaml:"max-concurrent"`
	DailyRequestBudget  int    `yaml:"daily-request-budget"`
	TrustProxyHeaders   bool   `yaml:"trust-proxy-headers"`
}

type AIChatLimits struct {
	MaxRequestBodyBytes int
	MaxMessageRunes     int
	RequestsPerMinute   int
	MaxConcurrent       int
	DailyRequestBudget  int
}

const (
	defaultHTTPMaxRequestBodyBytes = 32 * 1024 * 1024
	defaultAIChatMaxRequestBody    = 64 * 1024
	defaultAIChatMaxMessageRunes   = 4000
	defaultAIChatRequestsPerMinute = 6
	defaultAIChatMaxConcurrent     = 1
	defaultAIChatDailyBudget       = 50
	minimumSecretBytes             = 32
)

type NotifyConfig struct {
	ServerChanKey string `yaml:"server-chan-key"`
}

type SiteConfig struct {
	SuperAdminUsername string `yaml:"super_admin_username"`
}

type Config struct {
	DatabaseConfig DatabaseConfig `yaml:"database"`
	HttpConfig     HttpConfig     `yaml:"http"`
	CosConfig      CosConfig      `yaml:"cos"`
	AIChatConfig   AIChatConfig   `yaml:"aichat"`
	NotifyConfig   NotifyConfig   `yaml:"notify"`
	SiteConfig     SiteConfig     `yaml:"site"`
}

func InitConfig(path string) *Config {
	conf := &Config{}
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	err = yaml.NewDecoder(f).Decode(conf)
	if err != nil {
		panic(err)
	}
	if err = conf.Validate(); err != nil {
		panic("invalid configuration: " + err.Error())
	}

	return conf
}

func (c *Config) EffectiveJWTKey() string {
	if v := strings.TrimSpace(os.Getenv("PERSONAL_PAGE_JWT_KEY")); v != "" {
		return v
	}
	if v := strings.TrimSpace(c.HttpConfig.JWTKey); v != "" {
		return v
	}
	if v := strings.TrimSpace(c.HttpConfig.SecretKey); v != "" {
		return v
	}
	return ""
}

func (c *Config) EffectiveSessionKey() string {
	if v := strings.TrimSpace(os.Getenv("PERSONAL_PAGE_SESSION_KEY")); v != "" {
		return v
	}
	if v := strings.TrimSpace(c.HttpConfig.SessionKey); v != "" {
		return v
	}
	return deriveKey(c.EffectiveJWTKey(), "personal-page/session-cookie/v1")
}

func (c *Config) EffectiveAIChatIdentityKey() string {
	if v := strings.TrimSpace(os.Getenv("AICHAT_IDENTITY_KEY")); v != "" {
		return v
	}
	return deriveKey(c.EffectiveJWTKey(), "personal-page/aichat-identity/v1")
}

func deriveKey(secret string, purpose string) string {
	if secret == "" {
		return ""
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(purpose))
	return hex.EncodeToString(mac.Sum(nil))
}

func (c *Config) Validate() error {
	if err := validateSecret("JWT key", c.EffectiveJWTKey()); err != nil {
		return err
	}
	if explicitSessionKey := explicitSecret("PERSONAL_PAGE_SESSION_KEY", c.HttpConfig.SessionKey); explicitSessionKey != "" {
		if err := validateSecret("session key", explicitSessionKey); err != nil {
			return err
		}
	}
	if identityKey := strings.TrimSpace(os.Getenv("AICHAT_IDENTITY_KEY")); identityKey != "" {
		if err := validateSecret("AI chat identity key", identityKey); err != nil {
			return err
		}
	}
	if _, err := c.EffectiveHTTPMaxRequestBodyBytes(); err != nil {
		return err
	}
	if _, err := c.EffectiveAIChatLimits(); err != nil {
		return err
	}
	if _, err := c.EffectiveAIChatTrustProxyHeaders(); err != nil {
		return err
	}
	return nil
}

func explicitSecret(envName string, configured string) string {
	if value := strings.TrimSpace(os.Getenv(envName)); value != "" {
		return value
	}
	return strings.TrimSpace(configured)
}

func validateSecret(name string, value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("%s is required", name)
	}
	if strings.EqualFold(value, "secret") {
		return fmt.Errorf("%s must not use the insecure literal \"secret\"", name)
	}
	if len([]byte(value)) < minimumSecretBytes {
		return fmt.Errorf("%s must contain at least %d bytes", name, minimumSecretBytes)
	}
	return nil
}

func (c *Config) EffectiveHTTPMaxRequestBodyBytes() (int, error) {
	return effectivePositiveInt("HTTP_MAX_REQUEST_BODY_BYTES", 0, defaultHTTPMaxRequestBodyBytes)
}

func (c *Config) EffectiveAIChatTrustProxyHeaders() (bool, error) {
	if raw := strings.TrimSpace(os.Getenv("AICHAT_TRUST_PROXY_HEADERS")); raw != "" {
		value, err := strconv.ParseBool(raw)
		if err != nil {
			return false, fmt.Errorf("AICHAT_TRUST_PROXY_HEADERS must be true or false")
		}
		return value, nil
	}
	return c.AIChatConfig.TrustProxyHeaders, nil
}

func (c *Config) EffectiveAIChatLimits() (AIChatLimits, error) {
	maxBody, err := effectivePositiveInt("AICHAT_MAX_REQUEST_BODY_BYTES", c.AIChatConfig.MaxRequestBodyBytes, defaultAIChatMaxRequestBody)
	if err != nil {
		return AIChatLimits{}, err
	}
	maxMessage, err := effectivePositiveInt("AICHAT_MAX_MESSAGE_RUNES", c.AIChatConfig.MaxMessageRunes, defaultAIChatMaxMessageRunes)
	if err != nil {
		return AIChatLimits{}, err
	}
	requestsPerMinute, err := effectivePositiveInt("AICHAT_REQUESTS_PER_MINUTE", c.AIChatConfig.RequestsPerMinute, defaultAIChatRequestsPerMinute)
	if err != nil {
		return AIChatLimits{}, err
	}
	maxConcurrent, err := effectivePositiveInt("AICHAT_MAX_CONCURRENT", c.AIChatConfig.MaxConcurrent, defaultAIChatMaxConcurrent)
	if err != nil {
		return AIChatLimits{}, err
	}
	dailyBudget, err := effectivePositiveInt("AICHAT_DAILY_REQUEST_BUDGET", c.AIChatConfig.DailyRequestBudget, defaultAIChatDailyBudget)
	if err != nil {
		return AIChatLimits{}, err
	}
	if maxMessage > maxBody {
		return AIChatLimits{}, fmt.Errorf("AICHAT_MAX_MESSAGE_RUNES must not exceed AICHAT_MAX_REQUEST_BODY_BYTES")
	}
	return AIChatLimits{
		MaxRequestBodyBytes: maxBody,
		MaxMessageRunes:     maxMessage,
		RequestsPerMinute:   requestsPerMinute,
		MaxConcurrent:       maxConcurrent,
		DailyRequestBudget:  dailyBudget,
	}, nil
}

func effectivePositiveInt(envName string, configured int, fallback int) (int, error) {
	if raw := strings.TrimSpace(os.Getenv(envName)); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value <= 0 {
			return 0, fmt.Errorf("%s must be a positive integer", envName)
		}
		return value, nil
	}
	if configured < 0 {
		return 0, fmt.Errorf("%s configured value must be a positive integer", envName)
	}
	if configured > 0 {
		return configured, nil
	}
	return fallback, nil
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

func (c *Config) EffectiveAIChatModel() string {
	if v := os.Getenv("DIFY_MODEL"); v != "" {
		return v
	}
	if v := os.Getenv("AICHAT_MODEL"); v != "" {
		return v
	}
	if c.AIChatConfig.Model != "" {
		return c.AIChatConfig.Model
	}
	return "unknown-model"
}

func (c *Config) EffectiveAIChatChannel() string {
	if v := os.Getenv("DIFY_CHANNEL"); v != "" {
		return v
	}
	if v := os.Getenv("AICHAT_CHANNEL"); v != "" {
		return v
	}
	endpoint := c.EffectiveAIChatURL()
	if parsed, err := url.Parse(endpoint); err == nil && parsed.Hostname() != "" {
		return strings.ToLower(parsed.Hostname())
	}
	return "dify"
}

func (c *Config) EffectiveSuperAdminUsername() string {
	if v := os.Getenv("PERSONAL_PAGE_SUPER_ADMIN_USERNAME"); v != "" {
		return v
	}
	return c.SiteConfig.SuperAdminUsername
}
