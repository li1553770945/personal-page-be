package config

import (
	"strings"
	"testing"
)

func TestConfigSecretPrecedenceAndDerivation(t *testing.T) {
	clearConfigEnvironment(t)
	legacyKey := strings.Repeat("l", 32)
	configuredJWT := strings.Repeat("j", 32)
	environmentJWT := strings.Repeat("e", 32)
	conf := &Config{HttpConfig: HttpConfig{SecretKey: legacyKey, JWTKey: configuredJWT}}

	if got := conf.EffectiveJWTKey(); got != configuredJWT {
		t.Fatalf("configured jwt key = %q, want configured key", got)
	}
	t.Setenv("PERSONAL_PAGE_JWT_KEY", environmentJWT)
	if got := conf.EffectiveJWTKey(); got != environmentJWT {
		t.Fatalf("environment jwt key = %q, want environment key", got)
	}

	sessionKey := conf.EffectiveSessionKey()
	identityKey := conf.EffectiveAIChatIdentityKey()
	if sessionKey == "" || identityKey == "" {
		t.Fatal("derived keys must not be empty")
	}
	if sessionKey == environmentJWT || identityKey == environmentJWT || sessionKey == identityKey {
		t.Fatal("derived keys must be domain-separated from the jwt key and each other")
	}
	if sessionKey != conf.EffectiveSessionKey() || identityKey != conf.EffectiveAIChatIdentityKey() {
		t.Fatal("derived keys must be stable")
	}
}

func TestConfigLegacySecretKeyPathRemainsSupported(t *testing.T) {
	clearConfigEnvironment(t)
	legacyKey := strings.Repeat("k", 32)
	conf := &Config{HttpConfig: HttpConfig{SecretKey: legacyKey}}
	if got := conf.EffectiveJWTKey(); got != legacyKey {
		t.Fatalf("legacy http.secret_key = %q, want %q", got, legacyKey)
	}
	if err := conf.Validate(); err != nil {
		t.Fatalf("legacy secret path should validate: %v", err)
	}
}

func TestConfigFailsClosedForMissingOrWeakSecrets(t *testing.T) {
	clearConfigEnvironment(t)
	for name, conf := range map[string]*Config{
		"missing":        {},
		"literal secret": {HttpConfig: HttpConfig{JWTKey: "secret"}},
		"too short":      {HttpConfig: HttpConfig{JWTKey: "short-key"}},
	} {
		t.Run(name, func(t *testing.T) {
			if err := conf.Validate(); err == nil {
				t.Fatal("expected invalid secret configuration to fail")
			}
		})
	}
}

func TestConfigAIChatLimitDefaultsAndEnvironmentOverrides(t *testing.T) {
	clearConfigEnvironment(t)
	conf := &Config{HttpConfig: HttpConfig{JWTKey: strings.Repeat("j", 32)}}
	limits, err := conf.EffectiveAIChatLimits()
	if err != nil {
		t.Fatal(err)
	}
	if limits.MaxRequestBodyBytes != 64*1024 || limits.MaxMessageRunes != 4000 || limits.RequestsPerMinute != 6 || limits.MaxConcurrent != 1 || limits.DailyRequestBudget != 50 {
		t.Fatalf("unexpected defaults: %+v", limits)
	}

	t.Setenv("AICHAT_MAX_REQUEST_BODY_BYTES", "65536")
	t.Setenv("AICHAT_MAX_MESSAGE_RUNES", "5000")
	t.Setenv("AICHAT_REQUESTS_PER_MINUTE", "3")
	t.Setenv("AICHAT_MAX_CONCURRENT", "2")
	t.Setenv("AICHAT_DAILY_REQUEST_BUDGET", "12")
	limits, err = conf.EffectiveAIChatLimits()
	if err != nil {
		t.Fatal(err)
	}
	if limits.MaxRequestBodyBytes != 65536 || limits.MaxMessageRunes != 5000 || limits.RequestsPerMinute != 3 || limits.MaxConcurrent != 2 || limits.DailyRequestBudget != 12 {
		t.Fatalf("unexpected overrides: %+v", limits)
	}

	t.Setenv("AICHAT_REQUESTS_PER_MINUTE", "0")
	if _, err = conf.EffectiveAIChatLimits(); err == nil {
		t.Fatal("zero rate limit must fail closed")
	}
}

func TestConfigDoesNotTrustProxyHeadersByDefault(t *testing.T) {
	clearConfigEnvironment(t)
	conf := &Config{HttpConfig: HttpConfig{JWTKey: strings.Repeat("j", 32)}}
	trusted, err := conf.EffectiveAIChatTrustProxyHeaders()
	if err != nil || trusted {
		t.Fatalf("default trust = %v, err = %v", trusted, err)
	}
	t.Setenv("AICHAT_TRUST_PROXY_HEADERS", "true")
	trusted, err = conf.EffectiveAIChatTrustProxyHeaders()
	if err != nil || !trusted {
		t.Fatalf("environment trust = %v, err = %v", trusted, err)
	}
	t.Setenv("AICHAT_TRUST_PROXY_HEADERS", "not-a-bool")
	if _, err = conf.EffectiveAIChatTrustProxyHeaders(); err == nil {
		t.Fatal("invalid proxy trust value must fail closed")
	}
}

func clearConfigEnvironment(t *testing.T) {
	t.Helper()
	for _, name := range []string{
		"PERSONAL_PAGE_JWT_KEY",
		"PERSONAL_PAGE_SESSION_KEY",
		"AICHAT_IDENTITY_KEY",
		"HTTP_MAX_REQUEST_BODY_BYTES",
		"AICHAT_MAX_REQUEST_BODY_BYTES",
		"AICHAT_MAX_MESSAGE_RUNES",
		"AICHAT_REQUESTS_PER_MINUTE",
		"AICHAT_MAX_CONCURRENT",
		"AICHAT_DAILY_REQUEST_BUDGET",
		"AICHAT_TRUST_PROXY_HEADERS",
	} {
		t.Setenv(name, "")
	}
}
