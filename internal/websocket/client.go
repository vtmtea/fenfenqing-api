package websocket

import (
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// 写入超时时间
	writeWait = 10 * time.Second

	// pong 响应超时时间
	pongWait = 60 * time.Second

	// ping 周期
	pingPeriod = (pongWait * 9) / 10

	// 最大消息大小
	maxMessageSize = 512
)

// Client WebSocket 客户端连接
type Client struct {
	Hub *Hub

	// WebSocket 连接
	Conn *websocket.Conn

	// 发送通道
	SendChan chan *Message

	// 房间 ID
	RoomID uint

	// 用户 ID
	UserID uint
}

// readPump 从 WebSocket 连接读取消息
func (c *Client) readPump() {
	defer func() {
		c.Hub.UnregisterClient(c)
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				// 非预期关闭错误
			}
			break
		}

		// 处理接收到的消息（可选）
		var msg map[string]interface{}
		if err := json.Unmarshal(message, &msg); err == nil {
			c.handleMessage(msg)
		}
	}
}

// WritePump 向 WebSocket 连接写入消息
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.SendChan:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub 关闭了通道
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			data, err := json.Marshal(message)
			if err != nil {
				continue
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(data)

			// 批量发送队列中的消息
			n := len(c.SendChan)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				data, _ := json.Marshal(<-c.SendChan)
				w.Write(data)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ReadPump 从 WebSocket 连接读取消息
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.UnregisterClient(c)
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				// 非预期关闭错误
			}
			break
		}

		// 处理接收到的消息（可选）
		var msg map[string]interface{}
		if err := json.Unmarshal(message, &msg); err == nil {
			c.handleMessage(msg)
		}
	}
}

// handleMessage 处理接收到的消息
func (c *Client) handleMessage(msg map[string]interface{}) {
	// 可以处理客户端发送的消息，例如心跳等
	msgType, ok := msg["type"].(string)
	if !ok {
		return
	}

	switch msgType {
	case "ping":
		// 响应心跳
		c.SendChan <- &Message{Type: "pong"}
	case "join_room":
		// 处理加入房间请求
		if roomID, ok := msg["roomId"].(float64); ok {
			c.RoomID = uint(roomID)
			c.Hub.RegisterClient(c)
		}
	}
}

// Close 关闭客户端连接
func (c *Client) Close() {
	close(c.SendChan)
}

// Send 发送消息
func (c *Client) Send(message *Message) {
	select {
	case c.SendChan <- message:
	default:
		// 通道已满
	}
}
