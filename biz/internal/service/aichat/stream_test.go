package aichat

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"personal-page-be/biz/infra/config"
)

func TestStreamDifyUsesPseudonymousUser(t *testing.T) {
	t.Setenv("DIFY_API_URL", "")
	t.Setenv("DIFY_API_KEY", "")
	received := make(chan difyRequest, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req difyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		received <- req
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	service := &AIChatService{Config: &config.Config{AIChatConfig: config.AIChatConfig{URL: server.URL, APIKey: "test-key"}}}
	if err := service.streamDify(context.Background(), sendMessageReq{Message: "hello"}, "visitor_deadbeef", io.Discard, &aiUsageCapture{}); err != nil {
		t.Fatal(err)
	}
	if req := <-received; req.User != "visitor_deadbeef" {
		t.Fatalf("Dify user = %q", req.User)
	}
}

func TestStreamDifyCancelsUpstreamRequest(t *testing.T) {
	t.Setenv("DIFY_API_URL", "")
	t.Setenv("DIFY_API_KEY", "")
	started := make(chan struct{})
	releaseServer := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		select {
		case <-r.Context().Done():
		case <-releaseServer:
		}
	}))
	defer func() {
		close(releaseServer)
		server.Close()
	}()

	service := &AIChatService{Config: &config.Config{AIChatConfig: config.AIChatConfig{URL: server.URL, APIKey: "test-key"}}}
	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() {
		result <- service.streamDify(ctx, sendMessageReq{Message: "hello"}, "visitor_deadbeef", io.Discard, &aiUsageCapture{})
	}()

	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("upstream request did not start")
	}
	cancel()
	select {
	case err := <-result:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("stream error = %v, want context canceled", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("stream did not stop after cancellation")
	}
}
