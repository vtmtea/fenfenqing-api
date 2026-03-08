package websocket

import (
	"sync"
)

// Hub 维护活跃客户端的连接并广播消息
type Hub struct {
	// 注册的客户端
	clients map[*Client]bool

	// 按房间分组的客户端
	roomClients map[uint]map[*Client]bool

	// 广播通道
	broadcast chan *Message

	// 注册请求
	register chan *Client

	// 注销请求
	unregister chan *Client

	mu sync.RWMutex
}

// Message 消息结构
type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// NewHub 创建新的 Hub
func NewHub() *Hub {
	return &Hub{
		clients:     make(map[*Client]bool),
		roomClients: make(map[uint]map[*Client]bool),
		broadcast:   make(chan *Message),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
	}
}

// Run 运行 Hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.RegisterClient(client)
		case client := <-h.unregister:
			h.UnregisterClient(client)
		case message := <-h.broadcast:
			h.broadcastMessage(message)
		}
	}
}

// RegisterClient 注册客户端
func (h *Hub) RegisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client]; ok {
		return
	}

	h.clients[client] = true

	// 将客户端添加到房间
	if client.RoomID > 0 {
		if _, ok := h.roomClients[client.RoomID]; !ok {
			h.roomClients[client.RoomID] = make(map[*Client]bool)
		}
		h.roomClients[client.RoomID][client] = true
	}
}

// UnregisterClient 注销客户端
func (h *Hub) UnregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client]; !ok {
		return
	}

	delete(h.clients, client)
	client.Close()

	// 从房间中移除
	if client.RoomID > 0 {
		if room, ok := h.roomClients[client.RoomID]; ok {
			delete(room, client)
			// 如果房间没有客户端了，删除房间
			if len(room) == 0 {
				delete(h.roomClients, client.RoomID)
			}
		}
	}
}

// broadcastMessage 广播消息
func (h *Hub) broadcastMessage(message *Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		select {
		case client.SendChan <- message:
		default:
			// 发送失败，客户端可能已断开
			go h.UnregisterClient(client)
		}
	}
}

// BroadcastToRoom 向指定房间广播消息
func (h *Hub) BroadcastToRoom(roomID uint, message *Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if room, ok := h.roomClients[roomID]; ok {
		for client := range room {
			select {
			case client.SendChan <- message:
			default:
				// 发送失败，客户端可能已断开
				go h.UnregisterClient(client)
			}
		}
	}
}

// GetRoomClientCount 获取房间客户端数量
func (h *Hub) GetRoomClientCount(roomID uint) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if room, ok := h.roomClients[roomID]; ok {
		return len(room)
	}
	return 0
}
