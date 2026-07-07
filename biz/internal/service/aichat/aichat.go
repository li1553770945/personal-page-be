package aichat

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"personal-page-be/biz/internal/domain"
	"personal-page-be/biz/internal/dto"
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

	userID, username := s.userFromRequest(c)
	usage := &aiUsageCapture{
		RequestID:      uuid.NewString(),
		UserID:         userID,
		Username:       username,
		Channel:        s.Config.EffectiveAIChatChannel(),
		Model:          s.Config.EffectiveAIChatModel(),
		ConversationID: req.ConversationID,
		Currency:       "USD",
	}

	go func() {
		defer func() {
			_ = writer.Close()
		}()
		if err := s.streamDify(context.Background(), req, writer, usage); err != nil {
			s.Logger.Error("AI chat stream failed: " + err.Error())
			_ = writeSSE(writer, eventTypeError, err.Error())
			s.saveUsage(usage, domain.AIUsageStatusFailed, err)
			return
		}
		s.saveUsage(usage, domain.AIUsageStatusSuccess, nil)
	}()
}

func (s *AIChatService) streamDify(ctx context.Context, req sendMessageReq, writer io.Writer, usage *aiUsageCapture) error {
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
			usage.ConversationID = envelope.ConversationID
			sentConversationID = true
		}
		if !sentMessageID && envelope.MessageID != "" {
			if err = writeSSE(writer, eventTypeMessageID, envelope.MessageID); err != nil {
				return err
			}
			usage.MessageID = envelope.MessageID
			sentMessageID = true
		}
		if envelope.ConversationID != "" {
			usage.ConversationID = envelope.ConversationID
		}
		if envelope.MessageID != "" {
			usage.MessageID = envelope.MessageID
		}
		if envelope.Event == "message_end" && len(envelope.Metadata) > 0 {
			usage.applyMetadata(envelope.Metadata)
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

type aiUsageCapture struct {
	RequestID        string
	UserID           uint
	Username         string
	Channel          string
	Model            string
	ConversationID   string
	MessageID        string
	PromptTokens     int64
	CompletionTokens int64
	TotalTokens      int64
	TotalPrice       float64
	Currency         string
}

func (u *aiUsageCapture) applyMetadata(raw json.RawMessage) {
	var metadata map[string]interface{}
	if err := json.Unmarshal(raw, &metadata); err != nil {
		return
	}
	usage, _ := metadata["usage"].(map[string]interface{})
	if usage == nil {
		usage = metadata
	}

	if model := firstString(usage, metadata, "model", "model_name", "modelName"); model != "" {
		u.Model = model
	}
	if channel := firstString(usage, metadata, "channel", "provider", "model_provider", "modelProvider"); channel != "" {
		u.Channel = channel
	}
	if currency := firstString(usage, metadata, "currency"); currency != "" {
		u.Currency = strings.ToUpper(currency)
	}

	u.PromptTokens = firstInt64(usage, metadata, "prompt_tokens", "promptTokens", "input_tokens", "inputTokens")
	u.CompletionTokens = firstInt64(usage, metadata, "completion_tokens", "completionTokens", "output_tokens", "outputTokens")
	u.TotalTokens = firstInt64(usage, metadata, "total_tokens", "totalTokens")
	if u.TotalTokens == 0 {
		u.TotalTokens = u.PromptTokens + u.CompletionTokens
	}
	u.TotalPrice = firstFloat64(usage, metadata, "total_price", "totalPrice", "total_cost", "totalCost", "price", "cost")
}

func (s *AIChatService) saveUsage(usage *aiUsageCapture, status string, err error) {
	if usage == nil {
		return
	}
	if usage.Model == "" {
		usage.Model = s.Config.EffectiveAIChatModel()
	}
	if usage.Channel == "" {
		usage.Channel = s.Config.EffectiveAIChatChannel()
	}
	if usage.Currency == "" {
		usage.Currency = "USD"
	}
	errorMessage := ""
	if err != nil {
		errorMessage = err.Error()
		if len(errorMessage) > 1024 {
			errorMessage = errorMessage[:1024]
		}
	}

	if saveErr := s.Repo.SaveAIUsage(&domain.AIUsageEntity{
		UserID:           usage.UserID,
		Username:         usage.Username,
		Channel:          usage.Channel,
		Model:            usage.Model,
		ConversationID:   usage.ConversationID,
		MessageID:        usage.MessageID,
		RequestID:        usage.RequestID,
		Status:           status,
		PromptTokens:     usage.PromptTokens,
		CompletionTokens: usage.CompletionTokens,
		TotalTokens:      usage.TotalTokens,
		TotalPrice:       usage.TotalPrice,
		Currency:         usage.Currency,
		ErrorMessage:     errorMessage,
	}); saveErr != nil {
		s.Logger.Error("save ai usage failed: " + saveErr.Error())
	}
}

func firstString(primary map[string]interface{}, fallback map[string]interface{}, keys ...string) string {
	for _, source := range []map[string]interface{}{primary, fallback} {
		for _, key := range keys {
			if source == nil {
				continue
			}
			if value, ok := source[key]; ok {
				if text := strings.TrimSpace(fmt.Sprint(value)); text != "" && text != "<nil>" {
					return text
				}
			}
		}
	}
	return ""
}

func firstInt64(primary map[string]interface{}, fallback map[string]interface{}, keys ...string) int64 {
	for _, source := range []map[string]interface{}{primary, fallback} {
		for _, key := range keys {
			if source == nil {
				continue
			}
			if value, ok := source[key]; ok {
				switch v := value.(type) {
				case float64:
					return int64(v)
				case int64:
					return v
				case int:
					return int64(v)
				case json.Number:
					n, _ := v.Int64()
					return n
				case string:
					var n int64
					if _, err := fmt.Sscan(strings.TrimSpace(v), &n); err == nil {
						return n
					}
				}
			}
		}
	}
	return 0
}

func firstFloat64(primary map[string]interface{}, fallback map[string]interface{}, keys ...string) float64 {
	for _, source := range []map[string]interface{}{primary, fallback} {
		for _, key := range keys {
			if source == nil {
				continue
			}
			if value, ok := source[key]; ok {
				switch v := value.(type) {
				case float64:
					return v
				case int64:
					return float64(v)
				case int:
					return float64(v)
				case json.Number:
					n, _ := v.Float64()
					return n
				case string:
					var n float64
					if _, err := fmt.Sscan(strings.TrimSpace(v), &n); err == nil {
						return n
					}
				}
			}
		}
	}
	return 0
}

func (s *AIChatService) GetMyUsageStats(ctx context.Context, c *app.RequestContext) {
	user, ok := s.currentUser(ctx, c)
	if !ok {
		return
	}
	userID := user.ID
	resp, err := s.usageStats(c, &userID, false)
	if err != nil {
		response.Error(c, 5001, err.Error())
		return
	}
	response.OK(c, resp, "ok")
}

func (s *AIChatService) GetAdminUsageStats(ctx context.Context, c *app.RequestContext) {
	user, ok := s.currentUser(ctx, c)
	if !ok {
		return
	}
	if user.ID == 0 || !user.CanUse || !domain.IsSuperAdminRole(user.Role) {
		response.Error(c, 4003, "无权执行此操作")
		return
	}
	resp, err := s.usageStats(c, nil, true)
	if err != nil {
		response.Error(c, 5001, err.Error())
		return
	}
	response.OK(c, resp, "ok")
}

func (s *AIChatService) currentUser(ctx context.Context, c *app.RequestContext) (*domain.UserEntity, bool) {
	if username, ok := ctx.Value("username").(string); ok && username != "" {
		user, err := s.Repo.FindUser(username)
		if err != nil {
			response.Error(c, 5001, err.Error())
			return nil, false
		}
		return user, true
	}
	if userID, ok := ctx.Value("userId").(uint); ok && userID != 0 {
		user, err := s.Repo.FindUserByID(userID)
		if err != nil {
			response.Error(c, 5001, err.Error())
			return nil, false
		}
		return user, true
	}
	response.Error(c, 4003, "未登录")
	return nil, false
}

func (s *AIChatService) userFromRequest(c *app.RequestContext) (uint, string) {
	header := string(c.GetHeader("Authorization"))
	if header == "" {
		return 0, "anonymous"
	}
	tokenText := strings.TrimSpace(strings.TrimPrefix(header, "Bearer"))
	if tokenText == "" || tokenText == header {
		return 0, "anonymous"
	}

	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenText, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.Config.EffectiveJWTKey()), nil
	})
	if err != nil || token == nil || !token.Valid {
		return 0, "anonymous"
	}

	username, _ := claims["username"].(string)
	userID := uint(0)
	switch v := claims["userId"].(type) {
	case float64:
		userID = uint(v)
	case int:
		userID = uint(v)
	case int64:
		userID = uint(v)
	}
	if username == "" {
		username = "anonymous"
	}
	return userID, username
}

func (s *AIChatService) usageStats(c *app.RequestContext, userID *uint, includeChannels bool) (*dto.AIUsageStatsResp, error) {
	rng, err := parseUsageStatsRange(c)
	if err != nil {
		return nil, err
	}
	model := strings.TrimSpace(c.DefaultQuery("model", ""))
	channel := ""
	if includeChannels {
		channel = strings.TrimSpace(c.DefaultQuery("channel", ""))
	}

	optionRows, err := s.Repo.ListAIUsage(userID, &rng.Start, &rng.EndExclusive, "", "")
	if err != nil {
		return nil, err
	}
	rows, err := s.Repo.ListAIUsage(userID, &rng.Start, &rng.EndExclusive, model, channel)
	if err != nil {
		return nil, err
	}

	return buildUsageStatsResp(*rows, *optionRows, rng, includeChannels), nil
}

type usageStatsRange struct {
	Start        time.Time
	EndExclusive time.Time
	From         string
	To           string
	Dates        []string
	Location     *time.Location
}

func parseUsageStatsRange(c *app.RequestContext) (usageStatsRange, error) {
	location := usageStatsLocation()
	now := time.Now().In(location)
	defaultTo := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)
	defaultFrom := defaultTo.AddDate(0, 0, -29)

	fromText := strings.TrimSpace(c.DefaultQuery("from", ""))
	toText := strings.TrimSpace(c.DefaultQuery("to", ""))
	from := defaultFrom
	to := defaultTo
	var err error
	if fromText != "" {
		from, err = time.ParseInLocation("2006-01-02", fromText, location)
		if err != nil {
			return usageStatsRange{}, fmt.Errorf("from 日期格式应为 YYYY-MM-DD")
		}
	}
	if toText != "" {
		to, err = time.ParseInLocation("2006-01-02", toText, location)
		if err != nil {
			return usageStatsRange{}, fmt.Errorf("to 日期格式应为 YYYY-MM-DD")
		}
	}
	if to.Before(from) {
		return usageStatsRange{}, fmt.Errorf("结束日期不能早于开始日期")
	}
	if to.Sub(from).Hours()/24 > 370 {
		return usageStatsRange{}, fmt.Errorf("最多只能查询 370 天")
	}

	dates := make([]string, 0)
	for day := from; !day.After(to); day = day.AddDate(0, 0, 1) {
		dates = append(dates, day.Format("2006-01-02"))
	}

	return usageStatsRange{
		Start:        from,
		EndExclusive: to.AddDate(0, 0, 1),
		From:         from.Format("2006-01-02"),
		To:           to.Format("2006-01-02"),
		Dates:        dates,
		Location:     location,
	}, nil
}

func usageStatsLocation() *time.Location {
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return time.Local
	}
	return location
}

type usageAgg struct {
	RequestCount     int64
	SuccessCount     int64
	FailedCount      int64
	PromptTokens     int64
	CompletionTokens int64
	TotalTokens      int64
	TotalPrice       float64
	Currency         string
}

type dailyAgg struct {
	usageAgg
	ModelAggs   map[string]*usageAgg
	ChannelAggs map[string]*usageAgg
}

func buildUsageStatsResp(rows []domain.AIUsageEntity, optionRows []domain.AIUsageEntity, rng usageStatsRange, includeChannels bool) *dto.AIUsageStatsResp {
	total := &usageAgg{}
	daily := make(map[string]*dailyAgg, len(rng.Dates))
	modelAggs := make(map[string]*usageAgg)
	channelAggs := make(map[string]*usageAgg)
	channelModelAggs := make(map[string]*usageAgg)
	modelOptions := map[string]bool{}
	channelOptions := map[string]bool{}

	for _, date := range rng.Dates {
		daily[date] = &dailyAgg{
			ModelAggs:   map[string]*usageAgg{},
			ChannelAggs: map[string]*usageAgg{},
		}
	}

	for _, row := range optionRows {
		if row.Model != "" {
			modelOptions[row.Model] = true
		}
		if includeChannels && row.Channel != "" {
			channelOptions[row.Channel] = true
		}
	}

	for _, row := range rows {
		date := row.CreatedAt.In(rng.Location).Format("2006-01-02")
		if _, ok := daily[date]; !ok {
			continue
		}
		model := valueOrUnknown(row.Model)
		channel := valueOrUnknown(row.Channel)
		addUsage(total, row)
		addUsage(&daily[date].usageAgg, row)
		addUsage(aggFor(modelAggs, model), row)
		addUsage(aggFor(daily[date].ModelAggs, model), row)
		if includeChannels {
			addUsage(aggFor(channelAggs, channel), row)
			addUsage(aggFor(daily[date].ChannelAggs, channel), row)
			addUsage(aggFor(channelModelAggs, channel+"\x00"+model), row)
		}
	}

	resp := &dto.AIUsageStatsResp{
		Scope:          map[bool]string{true: "admin", false: "user"}[includeChannels],
		From:           rng.From,
		To:             rng.To,
		Models:         sortedKeys(modelOptions),
		Totals:         totalsDTO(total),
		Days:           make([]dto.AIUsageStatsDay, 0, len(rng.Dates)),
		ModelBreakdown: breakdownDTO(modelAggs, "model"),
	}
	if includeChannels {
		resp.Channels = sortedKeys(channelOptions)
		resp.ChannelBreakdown = breakdownDTO(channelAggs, "channel")
		resp.ChannelModelBreakdown = channelModelBreakdownDTO(channelModelAggs)
	}

	for _, date := range rng.Dates {
		day := daily[date]
		item := dto.AIUsageStatsDay{
			Date:             date,
			RequestCount:     day.RequestCount,
			SuccessCount:     day.SuccessCount,
			FailedCount:      day.FailedCount,
			PromptTokens:     day.PromptTokens,
			CompletionTokens: day.CompletionTokens,
			TotalTokens:      day.TotalTokens,
			TotalPrice:       day.TotalPrice,
			Currency:         normalizeCurrency(day.Currency),
			Models:           breakdownDTO(day.ModelAggs, "model"),
		}
		if includeChannels {
			item.Channels = breakdownDTO(day.ChannelAggs, "channel")
		}
		resp.Days = append(resp.Days, item)
	}
	return resp
}

func addUsage(agg *usageAgg, row domain.AIUsageEntity) {
	agg.RequestCount++
	if row.Status == domain.AIUsageStatusFailed {
		agg.FailedCount++
	} else {
		agg.SuccessCount++
	}
	agg.PromptTokens += row.PromptTokens
	agg.CompletionTokens += row.CompletionTokens
	agg.TotalTokens += row.TotalTokens
	agg.TotalPrice += row.TotalPrice
	agg.Currency = mergeCurrency(agg.Currency, row.Currency)
}

func aggFor(groups map[string]*usageAgg, key string) *usageAgg {
	if groups[key] == nil {
		groups[key] = &usageAgg{}
	}
	return groups[key]
}

func totalsDTO(agg *usageAgg) dto.AIUsageStatsTotals {
	return dto.AIUsageStatsTotals{
		RequestCount:     agg.RequestCount,
		SuccessCount:     agg.SuccessCount,
		FailedCount:      agg.FailedCount,
		PromptTokens:     agg.PromptTokens,
		CompletionTokens: agg.CompletionTokens,
		TotalTokens:      agg.TotalTokens,
		TotalPrice:       agg.TotalPrice,
		Currency:         normalizeCurrency(agg.Currency),
	}
}

func breakdownDTO(groups map[string]*usageAgg, field string) []dto.AIUsageStatsBreakdown {
	keys := make([]string, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		left := groups[keys[i]]
		right := groups[keys[j]]
		if left.TotalTokens == right.TotalTokens {
			return keys[i] < keys[j]
		}
		return left.TotalTokens > right.TotalTokens
	})

	result := make([]dto.AIUsageStatsBreakdown, 0, len(keys))
	for _, key := range keys {
		agg := groups[key]
		item := dto.AIUsageStatsBreakdown{
			RequestCount:     agg.RequestCount,
			SuccessCount:     agg.SuccessCount,
			FailedCount:      agg.FailedCount,
			PromptTokens:     agg.PromptTokens,
			CompletionTokens: agg.CompletionTokens,
			TotalTokens:      agg.TotalTokens,
			TotalPrice:       agg.TotalPrice,
			Currency:         normalizeCurrency(agg.Currency),
		}
		if field == "channel" {
			item.Channel = key
		} else {
			item.Model = key
		}
		result = append(result, item)
	}
	return result
}

func channelModelBreakdownDTO(groups map[string]*usageAgg) []dto.AIUsageStatsBreakdown {
	keys := make([]string, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		left := groups[keys[i]]
		right := groups[keys[j]]
		if left.TotalTokens == right.TotalTokens {
			return keys[i] < keys[j]
		}
		return left.TotalTokens > right.TotalTokens
	})

	result := make([]dto.AIUsageStatsBreakdown, 0, len(keys))
	for _, key := range keys {
		parts := strings.SplitN(key, "\x00", 2)
		channel := parts[0]
		model := ""
		if len(parts) > 1 {
			model = parts[1]
		}
		agg := groups[key]
		result = append(result, dto.AIUsageStatsBreakdown{
			Channel:          channel,
			Model:            model,
			RequestCount:     agg.RequestCount,
			SuccessCount:     agg.SuccessCount,
			FailedCount:      agg.FailedCount,
			PromptTokens:     agg.PromptTokens,
			CompletionTokens: agg.CompletionTokens,
			TotalTokens:      agg.TotalTokens,
			TotalPrice:       agg.TotalPrice,
			Currency:         normalizeCurrency(agg.Currency),
		})
	}
	return result
}

func sortedKeys(values map[string]bool) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func valueOrUnknown(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "unknown"
	}
	return value
}

func mergeCurrency(current string, next string) string {
	next = strings.ToUpper(strings.TrimSpace(next))
	if next == "" {
		return current
	}
	current = strings.ToUpper(strings.TrimSpace(current))
	if current == "" {
		return next
	}
	if current != next {
		return "MIXED"
	}
	return current
}

func normalizeCurrency(currency string) string {
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if currency == "" {
		return "USD"
	}
	return currency
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
