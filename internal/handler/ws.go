package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	gws "github.com/gorilla/websocket"
	"github.com/vtmtea/fenfenqing-api/internal/websocket"
)

var upgrader = gws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源
	},
}

// WSHandler WebSocket 处理器
type WSHandler struct {
	hub *websocket.Hub
}

// NewWSHandler 创建 WebSocket 处理器
func NewWSHandler(hub *websocket.Hub) *WSHandler {
	return &WSHandler{hub: hub}
}

// ServeWS 处理 WebSocket 连接
func (h *WSHandler) ServeWS(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	// 获取用户 ID（从 JWT token）
	userID, exists := c.Get("userID")
	if !exists {
		conn.Close()
		return
	}

	client := &websocket.Client{
		Hub:      h.hub,
		Conn:     conn,
		SendChan: make(chan *websocket.Message, 256),
		UserID:   userID.(uint),
	}

	// 注册客户端
	h.hub.RegisterClient(client)

	// 启动读写协程
	go client.WritePump()
	go client.ReadPump()
}
