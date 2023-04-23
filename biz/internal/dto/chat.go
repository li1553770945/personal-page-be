package dto

type SendMessageReq struct {
	ClientID string `json:"client_id"`
	Message  string `json:"message"`
	ChatID   string `json:"chat_id"`
}

type JoinChatReq struct {
	ChatID string `json:"chat_id"`
}
