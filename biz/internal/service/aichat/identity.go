package aichat

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net"
	"regexp"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
)

const aiVisitorHeader = "X-AI-Visitor-ID"

var visitorIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{16,128}$`)

type authenticatedIdentity struct {
	UserID   uint
	Username string
}

type requestIdentity struct {
	UserID      uint
	Username    string
	IdentityKey string
	IPKey       string
	DifyUser    string
}

func buildRequestIdentity(secret string, auth authenticatedIdentity, visitorID string, ip string) requestIdentity {
	identity := requestIdentity{UserID: auth.UserID, Username: auth.Username}
	if auth.UserID != 0 || auth.Username != "" {
		stableUser := strings.TrimSpace(auth.Username)
		if auth.UserID != 0 {
			stableUser = uintString(auth.UserID)
		}
		identity.IdentityKey = privacyHash(secret, "user:"+stableUser)
		identity.DifyUser = "user_" + identity.IdentityKey
		return withIPIdentity(identity, secret, ip)
	}

	visitorID = strings.TrimSpace(visitorID)
	if !visitorIDPattern.MatchString(visitorID) {
		visitorID = "ip-fallback:" + canonicalIP(ip)
	}
	identity.Username = "anonymous"
	identity.IdentityKey = privacyHash(secret, "visitor:"+visitorID)
	identity.DifyUser = "visitor_" + identity.IdentityKey
	return withIPIdentity(identity, secret, ip)
}

func withIPIdentity(identity requestIdentity, secret string, ip string) requestIdentity {
	identity.IPKey = privacyHash(secret, "ip:"+canonicalIP(ip))
	return identity
}

func privacyHash(secret string, value string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(value))
	return hex.EncodeToString(mac.Sum(nil)[:16])
}

func canonicalRequestIP(c *app.RequestContext, trustProxyHeaders bool) string {
	peerIP := "unknown"
	if c.RemoteAddr() != nil {
		peerIP = canonicalIP(c.RemoteAddr().String())
	}
	if trustProxyHeaders && isTrustedProxyPeer(peerIP) {
		if trusted := canonicalIP(string(c.GetHeader("X-Real-IP"))); trusted != "unknown" {
			return trusted
		}
	}
	return peerIP
}

func isTrustedProxyPeer(value string) bool {
	parsed := net.ParseIP(value)
	return parsed != nil && (parsed.IsLoopback() || parsed.IsPrivate())
}

func canonicalIP(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "unknown"
	}
	if host, _, err := net.SplitHostPort(value); err == nil {
		value = host
	}
	value = strings.Trim(value, "[]")
	parsed := net.ParseIP(value)
	if parsed == nil {
		return "unknown"
	}
	return parsed.String()
}

func uintString(value uint) string {
	if value == 0 {
		return "0"
	}
	var buf [20]byte
	index := len(buf)
	for value > 0 {
		index--
		buf[index] = byte('0' + value%10)
		value /= 10
	}
	return string(buf[index:])
}
