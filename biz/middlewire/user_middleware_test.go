package middlewire

import (
	"strings"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/golang-jwt/jwt/v5"
)

func TestInitAuthFailsClosedForWeakKeys(t *testing.T) {
	for _, secret := range []string{"", "secret", "short"} {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatalf("InitAuth(%q) should panic", secret)
				}
			}()
			InitAuth(secret)
		}()
	}
}

func TestUserFromBearerRequiresHS256(t *testing.T) {
	secret := strings.Repeat("s", 32)
	InitAuth(secret)
	claims := jwt.MapClaims{"username": "alice", "userId": 42}

	validToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}
	ctx := app.NewContext(0)
	ctx.Request.Header.Set("Authorization", "Bearer "+validToken)
	username, userID, ok := userFromBearer(ctx)
	if !ok || username != "alice" || userID != 42 {
		t.Fatalf("valid bearer identity = %q/%d/%v", username, userID, ok)
	}

	hs512Token, err := jwt.NewWithClaims(jwt.SigningMethodHS512, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}
	ctx.Request.Header.Set("Authorization", "Bearer "+hs512Token)
	if _, _, ok = userFromBearer(ctx); ok {
		t.Fatal("HS512 token must not be accepted by an HS256-only service")
	}
}
