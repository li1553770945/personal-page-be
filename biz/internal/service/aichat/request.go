package aichat

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"personal-page-be/biz/infra/config"
)

var errAIChatRequestTooLarge = errors.New("AI chat request body is too large")

func decodeSendMessage(body []byte, limits config.AIChatLimits) (sendMessageReq, error) {
	if len(body) > limits.MaxRequestBodyBytes {
		return sendMessageReq{}, errAIChatRequestTooLarge
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	var req sendMessageReq
	if err := decoder.Decode(&req); err != nil {
		return sendMessageReq{}, fmt.Errorf("invalid JSON request: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return sendMessageReq{}, errors.New("request body must contain exactly one JSON object")
	}

	req.Message = strings.TrimSpace(req.Message)
	req.ConversationID = strings.TrimSpace(req.ConversationID)
	if req.Message == "" {
		return sendMessageReq{}, errors.New("message cannot be empty")
	}
	if !utf8.ValidString(req.Message) {
		return sendMessageReq{}, errors.New("message must be valid UTF-8")
	}
	if utf8.RuneCountInString(req.Message) > limits.MaxMessageRunes {
		return sendMessageReq{}, fmt.Errorf("message exceeds the %d character limit", limits.MaxMessageRunes)
	}
	if len(req.ConversationID) > 191 {
		return sendMessageReq{}, errors.New("conversation_id is too long")
	}
	return req, nil
}
