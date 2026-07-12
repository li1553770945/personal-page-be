package aichat

import (
	"strings"
	"testing"
)

func TestBuildRequestIdentityIsStableIsolatedAndPrivate(t *testing.T) {
	secret := strings.Repeat("s", 64)
	visitorID := "018f2db4-5f72-7d3c-9b9f-1f9d7d55a222"
	visitor := buildRequestIdentity(secret, authenticatedIdentity{}, visitorID, "203.0.113.8")
	sameVisitor := buildRequestIdentity(secret, authenticatedIdentity{}, visitorID, "203.0.113.8")
	otherVisitor := buildRequestIdentity(secret, authenticatedIdentity{}, "018f2db4-5f72-7d3c-9b9f-1f9d7d55a333", "203.0.113.8")
	user := buildRequestIdentity(secret, authenticatedIdentity{UserID: 42, Username: "alice"}, visitorID, "203.0.113.8")

	if visitor.IdentityKey != sameVisitor.IdentityKey || visitor.DifyUser != sameVisitor.DifyUser {
		t.Fatal("the same visitor must receive a stable Dify identity")
	}
	if visitor.IdentityKey == otherVisitor.IdentityKey || visitor.IdentityKey == user.IdentityKey {
		t.Fatal("different visitors and authenticated users must be isolated")
	}
	for _, value := range []string{visitor.IdentityKey, visitor.IPKey, visitor.DifyUser} {
		if strings.Contains(value, visitorID) || strings.Contains(value, "203.0.113.8") || strings.Contains(value, "alice") {
			t.Fatalf("pseudonymous identity leaked raw input: %q", value)
		}
	}
	if user.UserID != 42 || !strings.HasPrefix(user.DifyUser, "user_") || !strings.HasPrefix(visitor.DifyUser, "visitor_") {
		t.Fatalf("unexpected identities: user=%+v visitor=%+v", user, visitor)
	}
}

func TestInvalidVisitorIDFallsBackToCanonicalIP(t *testing.T) {
	secret := strings.Repeat("s", 64)
	first := buildRequestIdentity(secret, authenticatedIdentity{}, "bad", "[2001:db8::1]:443")
	second := buildRequestIdentity(secret, authenticatedIdentity{}, "also bad", "2001:db8::1")
	if first.IdentityKey != second.IdentityKey {
		t.Fatal("invalid visitor IDs from the same IP should share the fallback identity")
	}
	if got := canonicalIP("192.0.2.10:1234"); got != "192.0.2.10" {
		t.Fatalf("canonical IPv4 = %q", got)
	}
}

func TestTrustedProxyPeerIsLimitedToLocalNetworks(t *testing.T) {
	for _, value := range []string{"127.0.0.1", "10.0.0.8", "172.16.1.8", "192.168.1.8", "::1", "fd00::8"} {
		if !isTrustedProxyPeer(value) {
			t.Fatalf("expected %s to be a trusted proxy peer", value)
		}
	}
	for _, value := range []string{"203.0.113.8", "2001:db8::8", "unknown", ""} {
		if isTrustedProxyPeer(value) {
			t.Fatalf("expected %s not to be a trusted proxy peer", value)
		}
	}
}
