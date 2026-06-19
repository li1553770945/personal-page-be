package aichat

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"personal-page-be/biz/internal/response"
)

const (
	eventTypeMessage          = "message"
	eventTypeConversationID   = "conversationId"
	eventTypeMessageID        = "messageId"
	eventTypeWorkflowStarted  = "workflowStarted"
	eventTypeWorkflowFinished = "workflowFinished"
	eventTypeNodeStarted      = "nodeStarted"
	eventTypeNodeFinished     = "nodeFinished"
	eventTypeMessageEnd       = "messageEnd"
	eventTypeError            = "error"
)

var httpClient = &http.Client{
	Timeout: 300 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	},
}

type sendMessageReq struct {
	Message        string `json:"message"`
	ConversationID string `json:"conversation_id"`
}

type difyRequest struct {
	Query            string                 `json:"query"`
	Inputs           map[string]interface{} `json:"inputs"`
	ResponseMode     string                 `json:"response_mode"`
	AutoGenerateName bool                   `json:"auto_generate_name"`
	User             string                 `json:"user"`
	ConversationID   string                 `json:"conversation_id,omitempty"`
	Files            []interface{}          `json:"files,omitempty"`
}

type difyErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"status"`
}

type difySSEEnvelope struct {
	Event                string          `json:"event"`
	ConversationID       string          `json:"conversation_id"`
	MessageID            string          `json:"message_id"`
	CreatedAt            int64           `json:"created_at"`
	TaskID               string          `json:"task_id"`
	WorkflowRunID        string          `json:"workflow_run_id"`
	ID                   string          `json:"id"`
	Answer               string          `json:"answer"`
	FromVariableSelector []string        `json:"from_variable_selector"`
	Data                 json.RawMessage `json:"data"`
	Metadata             json.RawMessage `json:"metadata"`
	Code                 string          `json:"code,omitempty"`
	Message              string          `json:"message,omitempty"`
	Status               int             `json:"status,omitempty"`
}

type difyWorkflowData struct {
	WorkflowID string  `json:"workflow_id"`
	Status     string  `json:"status"`
	Elapsed    float64 `json:"elapsed_time"`
	TotalSteps int     `json:"total_steps"`
}

type difyNodeData struct {
	NodeID      string  `json:"node_id"`
	NodeType    string  `json:"node_type"`
	Title       string  `json:"title"`
	Index       int     `json:"index"`
	Status      string  `json:"status,omitempty"`
	ElapsedTime float64 `json:"elapsed_time,omitempty"`
}

func (s *AIChatService) SendMessage(ctx context.Context, c *app.RequestContext) {
	var req sendMessageReq
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, 2001, "参数错误: "+err.Error())
		return
	}
	if req.Message == "" {
		response.Error(c, 2001, "message cannot be empty")
		return
	}
	if s.Config.EffectiveAIChatAPIKey() == "" {
		response.Error(c, 5001, "AI chat API key is not configured")
		return
	}

	reader, writer := io.Pipe()
	c.SetContentType("text/event-stream; charset=utf-8")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.SetBodyStream(reader, -1)

	go func() {
		defer func() {
			_ = writer.Close()
		}()
		if err := s.streamDify(context.Background(), req, writer); err != nil {
			s.Logger.Error("AI chat stream failed: " + err.Error())
			_ = writeSSE(writer, eventTypeError, err.Error())
		}
	}()
}

func (s *AIChatService) streamDify(ctx context.Context, req sendMessageReq, writer io.Writer) error {
	difyReq := difyRequest{
		Query:            req.Message,
		ConversationID:   req.ConversationID,
		Inputs:           map[string]interface{}{},
		ResponseMode:     "streaming",
		AutoGenerateName: false,
		User:             "default_user_001",
	}
	reqBody, err := json.Marshal(difyReq)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, s.Config.EffectiveAIChatURL(), bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.Config.EffectiveAIChatAPIKey())

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		var errResp difyErrorResponse
		if jsonErr := json.Unmarshal(bodyBytes, &errResp); jsonErr == nil && errResp.Message != "" {
			return fmt.Errorf("dify api error: %s", errResp.Message)
		}
		return fmt.Errorf("dify api status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	scanner := bufio.NewScanner(resp.Body)
	buf := make([]byte, 0, 256*1024)
	scanner.Buffer(buf, 10*1024*1024)

	sentConversationID := req.ConversationID != ""
	sentMessageID := false
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		dataStr := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if dataStr == "" || dataStr == "[DONE]" {
			continue
		}
		var envelope difySSEEnvelope
		if err = json.Unmarshal([]byte(dataStr), &envelope); err != nil {
			s.Logger.Warnf("unmarshal stream data failed: %v", err)
			continue
		}
		if !sentConversationID && envelope.ConversationID != "" {
			if err = writeSSE(writer, eventTypeConversationID, envelope.ConversationID); err != nil {
				return err
			}
			sentConversationID = true
		}
		if !sentMessageID && envelope.MessageID != "" {
			if err = writeSSE(writer, eventTypeMessageID, envelope.MessageID); err != nil {
				return err
			}
			sentMessageID = true
		}

		if err = s.forwardDifyEnvelope(writer, envelope); err != nil {
			return err
		}
	}
	return scanner.Err()
}

func (s *AIChatService) forwardDifyEnvelope(writer io.Writer, envelope difySSEEnvelope) error {
	switch envelope.Event {
	case "message", "agent_message":
		if envelope.Answer == "" || !shouldForwardChatflowMessage(envelope) {
			return nil
		}
		return writeSSE(writer, eventTypeMessage, envelope.Answer)
	case "workflow_started":
		return writeSSEJSON(writer, eventTypeWorkflowStarted, workflowPayload(envelope, difyWorkflowData{}))
	case "workflow_finished":
		var data difyWorkflowData
		_ = json.Unmarshal(envelope.Data, &data)
		return writeSSEJSON(writer, eventTypeWorkflowFinished, workflowPayload(envelope, data))
	case "node_started":
		var data difyNodeData
		_ = json.Unmarshal(envelope.Data, &data)
		return writeSSEJSON(writer, eventTypeNodeStarted, nodePayload(envelope, data))
	case "node_finished":
		var data difyNodeData
		_ = json.Unmarshal(envelope.Data, &data)
		return writeSSEJSON(writer, eventTypeNodeFinished, nodePayload(envelope, data))
	case "message_end":
		meta := "{}"
		if len(envelope.Metadata) > 0 {
			meta = string(envelope.Metadata)
		}
		return writeSSE(writer, eventTypeMessageEnd, meta)
	case "error":
		msg := envelope.Message
		if msg == "" {
			msg = "dify stream error"
		}
		_ = writeSSE(writer, eventTypeError, msg)
		return fmt.Errorf("dify stream error: %s", msg)
	default:
		return nil
	}
}

func shouldForwardChatflowMessage(envelope difySSEEnvelope) bool {
	if len(envelope.FromVariableSelector) >= 2 {
		return envelope.FromVariableSelector[1] == "text"
	}
	return true
}

func workflowPayload(envelope difySSEEnvelope, data difyWorkflowData) map[string]interface{} {
	return map[string]interface{}{
		"event":           envelope.Event,
		"workflow_run_id": envelope.WorkflowRunID,
		"workflow_id":     data.WorkflowID,
		"task_id":         envelope.TaskID,
		"created_at":      envelope.CreatedAt,
		"status":          data.Status,
		"elapsed_time":    data.Elapsed,
		"total_steps":     data.TotalSteps,
	}
}

func nodePayload(envelope difySSEEnvelope, data difyNodeData) map[string]interface{} {
	return map[string]interface{}{
		"event":           envelope.Event,
		"workflow_run_id": envelope.WorkflowRunID,
		"task_id":         envelope.TaskID,
		"node_id":         data.NodeID,
		"node_type":       data.NodeType,
		"title":           data.Title,
		"index":           data.Index,
		"status":          data.Status,
		"elapsed_time":    data.ElapsedTime,
	}
}

func writeSSEJSON(writer io.Writer, eventType string, data interface{}) error {
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return writeSSE(writer, eventType, string(body))
}

func writeSSE(writer io.Writer, eventType, data string) error {
	payload, err := json.Marshal(map[string]string{
		"event_type": eventType,
		"data":       data,
	})
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(writer, "data: %s\n\n", payload)
	return err
}
