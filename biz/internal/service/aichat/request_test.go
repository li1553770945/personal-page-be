package aichat

import (
	"errors"
	"strings"
	"testing"

	"personal-page-be/biz/infra/config"
)

func TestDecodeSendMessageBoundsAndValidatesJSON(t *testing.T) {
	limits := config.AIChatLimits{MaxRequestBodyBytes: 128, MaxMessageRunes: 4}
	req, err := decodeSendMessage([]byte(`{"message":" 你好 ","conversation_id":" conversation "}`), limits)
	if err != nil {
		t.Fatal(err)
	}
	if req.Message != "你好" || req.ConversationID != "conversation" {
		t.Fatalf("unexpected request: %+v", req)
	}

	if _, err = decodeSendMessage([]byte(`{"message":"hello"}`), limits); err == nil {
		t.Fatal("message rune limit should be enforced")
	}
	if _, err = decodeSendMessage([]byte(`{"message":"ok","unexpected":true}`), limits); err == nil {
		t.Fatal("unknown fields should be rejected")
	}
	if _, err = decodeSendMessage([]byte(strings.Repeat("x", 129)), limits); !errors.Is(err, errAIChatRequestTooLarge) {
		t.Fatalf("oversized body error = %v", err)
	}
}
