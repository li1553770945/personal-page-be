package domain

import (
	"github.com/hertz-contrib/websocket"
	"golang.org/x/sync/semaphore"
	"sync"
	"time"
)

type ChatMessageEntity struct {
	Type     string    `json:"type"`
	Data     string    `json:"data"`
	Time     time.Time `json:"time"`
	ErrorMsg string    `json:"error_msg"`
}

type ChatEntity struct {
	ChatID                 string
	CreaterID              string
	JoinerID               string
	CreaterMessageLock     sync.Mutex
	JoinerMessageLock      sync.Mutex
	CreaterMessageWriteSem *semaphore.Weighted
	CreaterMessageReadSem  *semaphore.Weighted
	JoinerMessageWriteSem  *semaphore.Weighted
	JoinerMessageReadSem   *semaphore.Weighted
	CreaterConn            *websocket.Conn
	JoinerConn             *websocket.Conn
	MessageSendByCreater   []*ChatMessageEntity
	MessageSendByJoiner    []*ChatMessageEntity
}
