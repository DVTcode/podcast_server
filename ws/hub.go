package ws

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	Conn *websocket.Conn
	Send chan []byte
}

type Hub struct {
	Clients map[string]map[*websocket.Conn]*Client // Key: documentID -> nhiều client
	Mutex   sync.RWMutex
}

var H = Hub{
	Clients: make(map[string]map[*websocket.Conn]*Client),
}

// Đăng ký client mới
func (h *Hub) Register(docID string, conn *websocket.Conn) {
	h.Mutex.Lock()
	defer h.Mutex.Unlock()

	if _, ok := h.Clients[docID]; !ok {
		h.Clients[docID] = make(map[*websocket.Conn]*Client)
	}

	client := &Client{
		Conn: conn,
		Send: make(chan []byte, 256),
	}

	h.Clients[docID][conn] = client

	go h.readPump(docID, conn)
	go h.writePump(docID, conn)
}

// Gửi message tới tất cả client đang theo dõi document đó
func (h *Hub) Send(docID string, message string) {
	h.Mutex.RLock()
	defer h.Mutex.RUnlock()

	if clients, ok := h.Clients[docID]; ok {
		for _, client := range clients {
			select {
			case client.Send <- []byte(message):
			default:
				// channel đầy, bỏ qua để tránh block
			}
		}
	}
}

// Dọn dẹp client khi ngắt kết nối
func (h *Hub) Unregister(docID string, conn *websocket.Conn) {
	h.Mutex.Lock()
	defer h.Mutex.Unlock()

	if clients, ok := h.Clients[docID]; ok {
		if client, ok := clients[conn]; ok {
			close(client.Send)
			delete(clients, conn)
		}
		if len(clients) == 0 {
			delete(h.Clients, docID)
		}
	}
}

func (h *Hub) readPump(docID string, conn *websocket.Conn) {
	defer h.Unregister(docID, conn)

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}

func (h *Hub) writePump(docID string, conn *websocket.Conn) {
	client := h.Clients[docID][conn]
	defer func() {
		conn.WriteMessage(websocket.CloseMessage, []byte{})
		conn.Close()
	}()

	for msg := range client.Send {
		if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			break
		}
	}
}

func (h *Hub) Broadcast(docID string, messageType int, data []byte) {
	h.Mutex.RLock()
	defer h.Mutex.RUnlock()

	if clients, ok := h.Clients[docID]; ok {
		for _, client := range clients {
			select {
			case client.Send <- data:
			default:
				// Nếu channel bị nghẽn, bỏ qua
			}
		}
	}
}

type StatusMessage struct {
	Status string `json:"status"`
}

func SendStatus(docID string, message string) {
	msg := StatusMessage{Status: message}
	data, err := json.Marshal(msg)
	if err != nil {
		log.Println("JSON marshal error:", err)
		return
	}
	H.Broadcast(docID, websocket.TextMessage, data)
}
