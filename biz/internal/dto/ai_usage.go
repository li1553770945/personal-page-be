package dto

type AIUsageStatsResp struct {
	Scope                 string                  `json:"scope"`
	From                  string                  `json:"from"`
	To                    string                  `json:"to"`
	Models                []string                `json:"models"`
	Channels              []string                `json:"channels,omitempty"`
	Totals                AIUsageStatsTotals      `json:"totals"`
	Days                  []AIUsageStatsDay       `json:"days"`
	ModelBreakdown        []AIUsageStatsBreakdown `json:"model_breakdown"`
	ChannelBreakdown      []AIUsageStatsBreakdown `json:"channel_breakdown,omitempty"`
	ChannelModelBreakdown []AIUsageStatsBreakdown `json:"channel_model_breakdown,omitempty"`
}

type AIUsageStatsTotals struct {
	RequestCount     int64   `json:"request_count"`
	SuccessCount     int64   `json:"success_count"`
	FailedCount      int64   `json:"failed_count"`
	PromptTokens     int64   `json:"prompt_tokens"`
	CompletionTokens int64   `json:"completion_tokens"`
	TotalTokens      int64   `json:"total_tokens"`
	TotalPrice       float64 `json:"total_price"`
	Currency         string  `json:"currency"`
}

type AIUsageStatsDay struct {
	Date             string                  `json:"date"`
	RequestCount     int64                   `json:"request_count"`
	SuccessCount     int64                   `json:"success_count"`
	FailedCount      int64                   `json:"failed_count"`
	PromptTokens     int64                   `json:"prompt_tokens"`
	CompletionTokens int64                   `json:"completion_tokens"`
	TotalTokens      int64                   `json:"total_tokens"`
	TotalPrice       float64                 `json:"total_price"`
	Currency         string                  `json:"currency"`
	Models           []AIUsageStatsBreakdown `json:"models"`
	Channels         []AIUsageStatsBreakdown `json:"channels,omitempty"`
}

type AIUsageStatsBreakdown struct {
	Date             string  `json:"date,omitempty"`
	Channel          string  `json:"channel,omitempty"`
	Model            string  `json:"model,omitempty"`
	RequestCount     int64   `json:"request_count"`
	SuccessCount     int64   `json:"success_count"`
	FailedCount      int64   `json:"failed_count"`
	PromptTokens     int64   `json:"prompt_tokens"`
	CompletionTokens int64   `json:"completion_tokens"`
	TotalTokens      int64   `json:"total_tokens"`
	TotalPrice       float64 `json:"total_price"`
	Currency         string  `json:"currency"`
}
