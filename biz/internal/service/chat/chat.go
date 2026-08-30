package chat

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/hertz-contrib/websocket"
	"personal-page-be/biz/internal/domain"
	"personal-page-be/biz/internal/response"
)

const (
	roomLifetime       = time.Hour
	chatReadLimit      = 64 * 1024
	maxRoomClients     = 2
	clientAuthDeadline = 15 * time.Second
)

type wireMessage struct {
	Event string `json:"event"`
	Type  string `json:"type,omitempty"`
	Data  string `json:"data,omitempty"`
}

func (s *ChatService) CreateRoom(ctx context.Context, c *app.RequestContext) {
	username, _ := ctx.Value("username").(string)
	user, err := s.Repo.FindUser(username)
	if err != nil {
		response.Error(c, 5001, err.Error())
		return
	}
	if user.ID == 0 || !user.CanUse || !domain.IsAdminRole(domain.NormalizeRole(user.Role)) {
		response.Error(c, 4003, "只有管理员可以创建临时聊天室")
		return
	}

	room := &roomState{clients: map[string]*roomClient{}, expiresAt: time.Now().Add(roomLifetime)}
	for {
		room.id, err = randomText(6)
		if err != nil {
			response.Error(c, 5001, "生成房间失败")
			return
		}
		s.roomsMu.Lock()
		if _, exists := s.rooms[room.id]; !exists {
			s.rooms[room.id] = room
			s.roomsMu.Unlock()
			break
		}
		s.roomsMu.Unlock()
	}
	client, err := room.addClient()
	if err != nil {
		response.Error(c, 5001, err.Error())
		return
	}
	response.OK(c, map[string]string{"roomId": room.id, "clientId": client.id, "clientToken": client.token}, "创建成功")
}

func (s *ChatService) JoinRoom(ctx context.Context, c *app.RequestContext) {
	roomID := strings.TrimSpace(c.DefaultQuery("roomId", ""))
	if roomID == "" {
		roomID = strings.TrimSpace(c.DefaultQuery("room_id", ""))
	}
	room := s.findRoom(roomID)
	if room == nil {
		response.Error(c, 4004, "房间不存在或已过期")
		return
	}
	client, err := room.addClient()
	if err != nil {
		response.Error(c, 4009, err.Error())
		return
	}
	response.OK(c, map[string]string{"roomId": room.id, "clientId": client.id, "clientToken": client.token}, "加入成功")
}

func (s *ChatService) Connect(ctx context.Context, c *app.RequestContext) {
	roomID := strings.TrimSpace(c.DefaultQuery("roomId", ""))
	if roomID == "" || s.findRoom(roomID) == nil {
		c.JSON(404, map[string]interface{}{"code": 4004, "message": "房间不存在或已过期"})
		return
	}
	upgrader := websocket.HertzUpgrader{CheckOrigin: allowedChatOrigin}
	if err := upgrader.Upgrade(c, func(conn *websocket.Conn) {
		s.handleConnection(roomID, conn)
	}); err != nil {
		s.Log.WithError(err).Warn("upgrade chat websocket failed")
	}
}

func (s *ChatService) handleConnection(roomID string, conn *websocket.Conn) {
	conn.SetReadLimit(chatReadLimit)
	_ = conn.SetReadDeadline(time.Now().Add(clientAuthDeadline))
	var auth wireMessage
	if err := conn.ReadJSON(&auth); err != nil || auth.Event != "im-auth-req" || auth.Data == "" {
		_ = conn.WriteJSON(wireMessage{Event: "im-close", Data: "认证失败"})
		_ = conn.Close()
		return
	}
	room := s.findRoom(roomID)
	if room == nil {
		_ = conn.WriteJSON(wireMessage{Event: "im-close", Data: "房间不存在或已过期"})
		_ = conn.Close()
		return
	}
	client := room.authenticate(auth.Data, conn)
	if client == nil {
		_ = conn.WriteJSON(wireMessage{Event: "im-close", Data: "客户端凭据无效"})
		_ = conn.Close()
		return
	}
	defer room.disconnect(client.id, conn)
	_ = conn.SetReadDeadline(time.Time{})
	if err := client.write(wireMessage{Event: "im-auth-resp", Data: "ok"}); err != nil {
		return
	}

	for {
		var message wireMessage
		if err := conn.ReadJSON(&message); err != nil {
			return
		}
		room.touch()
		switch message.Event {
		case "im-ping":
			if err := client.write(wireMessage{Event: "im-pong", Data: "im-pong"}); err != nil {
				return
			}
		case "im-message":
			if message.Type != "text" && message.Type != "file" {
				_ = client.write(wireMessage{Event: "im-error", Data: "不支持的消息类型"})
				continue
			}
			if strings.TrimSpace(message.Data) == "" {
				continue
			}
			if err := room.broadcast(client.id, wireMessage{Event: "im-message", Type: message.Type, Data: message.Data}); err != nil {
				_ = client.write(wireMessage{Event: "im-error", Data: "对方暂未连接，消息未送达"})
			}
		default:
			_ = client.write(wireMessage{Event: "im-error", Data: "未知事件"})
		}
	}
}

func (s *ChatService) findRoom(roomID string) *roomState {
	if roomID == "" {
		return nil
	}
	s.roomsMu.RLock()
	room := s.rooms[roomID]
	s.roomsMu.RUnlock()
	if room == nil {
		return nil
	}
	room.mu.Lock()
	expired := time.Now().After(room.expiresAt)
	room.mu.Unlock()
	if expired {
		s.roomsMu.Lock()
		if s.rooms[roomID] == room {
			delete(s.rooms, roomID)
		}
		s.roomsMu.Unlock()
		room.closeAll()
		return nil
	}
	return room
}

func (s *ChatService) cleanupExpiredRooms() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		var expired []*roomState
		s.roomsMu.Lock()
		for id, room := range s.rooms {
			room.mu.Lock()
			isExpired := now.After(room.expiresAt)
			room.mu.Unlock()
			if isExpired {
				delete(s.rooms, id)
				expired = append(expired, room)
			}
		}
		s.roomsMu.Unlock()
		for _, room := range expired {
			room.closeAll()
		}
	}
}

func (r *roomState) addClient() (*roomClient, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.clients) >= maxRoomClients {
		return nil, fmt.Errorf("房间人数已满")
	}
	id, err := randomText(18)
	if err != nil {
		return nil, err
	}
	token, err := randomText(32)
	if err != nil {
		return nil, err
	}
	client := &roomClient{id: id, token: token}
	r.clients[id] = client
	r.expiresAt = time.Now().Add(roomLifetime)
	return client, nil
}

func (r *roomState) authenticate(token string, conn *websocket.Conn) *roomClient {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, client := range r.clients {
		if client.token == token {
			if client.conn != nil && client.conn != conn {
				_ = client.conn.Close()
			}
			client.conn = conn
			r.expiresAt = time.Now().Add(roomLifetime)
			return client
		}
	}
	return nil
}

func (r *roomState) disconnect(clientID string, conn *websocket.Conn) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if client := r.clients[clientID]; client != nil && client.conn == conn {
		client.conn = nil
	}
}

func (r *roomState) broadcast(senderID string, message wireMessage) error {
	r.mu.Lock()
	clients := make([]*roomClient, 0, len(r.clients))
	for id, client := range r.clients {
		if id != senderID && client.conn != nil {
			clients = append(clients, client)
		}
	}
	r.mu.Unlock()
	if len(clients) == 0 {
		return fmt.Errorf("peer offline")
	}
	for _, client := range clients {
		if err := client.write(message); err != nil {
			return err
		}
	}
	return nil
}

func (r *roomState) touch() {
	r.mu.Lock()
	r.expiresAt = time.Now().Add(roomLifetime)
	r.mu.Unlock()
}

func (r *roomState) closeAll() {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, client := range r.clients {
		if client.conn != nil {
			_ = client.conn.WriteJSON(wireMessage{Event: "im-close", Data: "房间已过期"})
			_ = client.conn.Close()
			client.conn = nil
		}
	}
}

func (c *roomClient) write(message wireMessage) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	if c.conn == nil {
		return fmt.Errorf("client offline")
	}
	return c.conn.WriteJSON(message)
}

func allowedChatOrigin(c *app.RequestContext) bool {
	origin := strings.TrimSpace(string(c.GetHeader("Origin")))
	if origin == "" {
		return true
	}
	return origin == "https://peacesheep.xyz" || origin == "https://www.peacesheep.xyz" || strings.HasPrefix(origin, "http://localhost:")
}

func randomText(byteLength int) (string, error) {
	data := make([]byte, byteLength)
	if _, err := rand.Read(data); err != nil {
		return "", err
	}
	return strings.TrimRight(base64.RawURLEncoding.EncodeToString(data), "="), nil
}
