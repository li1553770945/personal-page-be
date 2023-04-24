package chat

import (
	"context"
	"encoding/json"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/hertz-contrib/websocket"
	"github.com/patrickmn/go-cache"
	"personal-page-be/biz/internal/domain"
	U "personal-page-be/biz/internal/utils"
	"time"
)

func (s *ChatService) CreateChat(ctx context.Context, c *app.RequestContext) {
	var upgrader = websocket.HertzUpgrader{
		CheckOrigin: func(ctx *app.RequestContext) bool {
			return true
		},
	}

	var chatID string
	for {
		chatID = U.RandSeq(4)
		_, found := s.Cache.Get(chatID)
		if !found {
			break
		}
	}
	var chatEntity domain.ChatEntity
	err := s.Cache.Add(chatID, &chatEntity, cache.DefaultExpiration)
	if err != nil {
		c.JSON(200, utils.H{"code": 5001, "msg": err.Error()})
		return
	}
	chatEntity.ChatID = chatID
	chatEntity.CreaterID = U.RandSeq(10)

	err = upgrader.Upgrade(c, func(conn *websocket.Conn) {
		s.MessageHandler(conn, &chatEntity, "creater")
	})
	if err != nil {
		s.Log.Error("update to websocket failed:", err.Error())
		c.JSON(200, utils.H{"code": 5001, "msg": err.Error()})
		return
	}
	//c.JSON(200, utils.H{"code": 0, "msg": "创建成功", "data": utils.H{"chat_id": chatEntity.ChatID, "client_id": chatEntity.CreaterID}})
}
func (s *ChatService) JoinChat(ctx context.Context, c *app.RequestContext) {
	var upgrader = websocket.HertzUpgrader{
		CheckOrigin: func(ctx *app.RequestContext) bool {
			return true
		},
	}
	ChatID := c.DefaultQuery("chat_id", "")
	clientID := c.DefaultQuery("client_id", "")
	chatInterface, found := s.Cache.Get(ChatID)
	if !found {
		c.JSON(200, utils.H{"code": 4004, "msg": "未找到相关chat"})
		return
	}

	var chatEntity *domain.ChatEntity
	if entity, ok := chatInterface.(*domain.ChatEntity); ok {
		chatEntity = entity
	} else {
		c.JSON(200, utils.H{"code": 5001, "msg": "内部错误，chatInterface断言失败"})
		return
	}
	var role string
	if clientID != "" {
		if clientID == chatEntity.CreaterID {
			role = "creater"
		} else if clientID == chatEntity.JoinerID {
			role = "joiner"
		} else {
			c.JSON(200, utils.H{"code": 5001, "msg": "客户端ID错误"})
			s.Log.Error("client id not found")
			return
		}
	} else {
		chatEntity.JoinerID = U.RandSeq(10)
		role = "joiner"
	}

	err := upgrader.Upgrade(c, func(conn *websocket.Conn) {
		s.MessageHandler(conn, chatEntity, role)
	})
	if err != nil {
		c.JSON(200, utils.H{"code": 5001, "msg": err.Error()})
		return
	}

}

func (s *ChatService) MessageHandler(recvConn *websocket.Conn, chatEntity *domain.ChatEntity, role string) {

	var msgEntity domain.ChatMessageEntity
	msgEntity.Data = chatEntity.ChatID
	msgEntity.Type = "chat_id"
	msgEntity.Time = time.Now()
	jsonBytes, _ := json.Marshal(msgEntity)
	err := recvConn.WriteMessage(websocket.TextMessage, jsonBytes)
	if err != nil {
		s.Log.Error("send chat id failed:", err.Error())
		return
	}

	msgEntity.Type = "client_id"
	msgEntity.Time = time.Now()

	if role == "creater" {
		msgEntity.Data = chatEntity.CreaterID
		chatEntity.CreaterConn = recvConn
	} else {
		msgEntity.Data = chatEntity.JoinerID
		chatEntity.JoinerConn = recvConn
	}

	jsonBytes, _ = json.Marshal(msgEntity)
	err = recvConn.WriteMessage(websocket.TextMessage, jsonBytes)
	if err != nil {
		s.Log.Error("send client id failed:", err.Error())
	}
	for {
		var result map[string]interface{}

		mt, message, err := recvConn.ReadMessage()
		if err != nil {
			s.Log.Error("receive msg failed:", err.Error())
			if chatEntity.JoinerConn != nil {
				chatEntity.JoinerConn.Close()
				chatEntity.JoinerConn = nil
			}
			if chatEntity.CreaterConn != nil {
				chatEntity.CreaterConn.Close()
				chatEntity.CreaterConn = nil
			}
			return
		}
		if err = json.Unmarshal([]byte(message), &result); err != nil {
			s.Log.Error("convert received message to json failed,message:", message)
			continue
		}

		if mt == websocket.TextMessage {
			var sendMsgEntity domain.ChatMessageEntity
			sendMsgEntity.Data = result["data"].(string)
			sendMsgEntity.Type = "client_message"
			sendMsgEntity.Time = time.Now()
			jsonBytes, err := json.Marshal(sendMsgEntity)
			if err != nil {
				s.Log.Error("convert sendMsgEntity to json failed", err.Error())
			}

			if role == "creater" {

				if chatEntity.JoinerConn != nil { // 如果发送方已经上线
					err = chatEntity.JoinerConn.WriteMessage(mt, jsonBytes)
					if err != nil {
						s.Log.Error("send msg to creater failed", err.Error())
					}
				} else {
					sendMsgEntity.Data = ""
					sendMsgEntity.Type = "warning"
					sendMsgEntity.Time = time.Now()
					sendMsgEntity.ErrorMsg = "对方未上线，发送失败"
					jsonBytes, _ := json.Marshal(sendMsgEntity)
					recvConn.WriteMessage(websocket.TextMessage, jsonBytes)
				}

			} else { // 如果是接收方
				if chatEntity.CreaterConn != nil {
					err = chatEntity.CreaterConn.WriteMessage(mt, jsonBytes)
					if err != nil {
						s.Log.Error("send msg to joiner failed", err.Error())
					}
				}
			}
		}

	}
}
func (s *ChatService) GetMessageList(ctx context.Context, c *app.RequestContext) {
	chatID := c.DefaultQuery("chat_id", "")
	id := c.DefaultQuery("client_id", "")
	chatInterface, found := s.Cache.Get(chatID)
	if !found {
		c.JSON(200, utils.H{"code": 4004, "msg": "chat不存在或已结束"})
		return
	}
	var chatEntity *domain.ChatEntity
	if entity, ok := chatInterface.(*domain.ChatEntity); ok {
		chatEntity = entity
	} else {
		c.JSON(200, utils.H{"code": 5001, "msg": "内部错误，chatInterface断言失败"})
		return
	}
	var messages []*domain.ChatMessageEntity
	if id == chatEntity.JoinerID {
		chatEntity.CreaterMessageLock.Lock()
		messages = chatEntity.MessageSendByCreater

		chatEntity.CreaterMessageLock.Unlock()
		c.JSON(200, utils.H{"code": 0, "data": messages})
	} else if id == chatEntity.CreaterID {
		chatEntity.JoinerMessageLock.Lock()
		messages = chatEntity.MessageSendByJoiner
		chatEntity.JoinerMessageLock.Unlock()
		c.JSON(200, utils.H{"code": 0, "data": messages})
	} else {
		c.JSON(200, utils.H{"code": 4004, "msg": "无效的ID"})
	}
}
func (s *ChatService) CloseChat(ctx context.Context, c *app.RequestContext) {
	chatID := string(c.FormValue("chat_id"))
	id := string(c.FormValue("id"))
	chatInterface, found := s.Cache.Get(chatID)
	if !found {
		c.JSON(200, utils.H{"code": 4004, "msg": "chat不存在或已结束"})
		return
	}
	var chatEntity *domain.ChatEntity
	if entity, ok := chatInterface.(*domain.ChatEntity); ok {
		chatEntity = entity
	} else {
		c.JSON(200, utils.H{"code": 5001, "msg": "内部错误，chatInterface断言失败"})
		return
	}

	if id != chatEntity.CreaterID {
		c.JSON(200, utils.H{"code": 4003, "msg": "无权进行操作"})
		return
	}
	s.Cache.Delete(chatID)
	c.JSON(200, utils.H{"code": 0, "msg": "操作成功"})
}
