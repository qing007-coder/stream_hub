package im

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	UserID   string
	Conn     *websocket.Conn
	Send     chan []byte
	ClientID string
}

type Hub struct {
	Clients    map[string]*Client
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan *MessageBroadcast
	mu         sync.RWMutex
}

type MessageBroadcast struct {
	ReceiverID string
	Message    []byte
}

const (
	writeWait = 60 * time.Second
	pingWait  = 60 * time.Second
	pingEvery = (pingWait * 9) / 10
)

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[string]*Client),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan *MessageBroadcast, 100),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.Clients[client.ClientID] = client
			h.mu.Unlock()
			log.Printf("Client registered: %s (UserID: %s)", client.ClientID, client.UserID)
		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.Clients[client.ClientID]; ok {
				delete(h.Clients, client.ClientID)
				close(client.Send)
				log.Printf("Client unregistered: %s (UserID: %s)", client.ClientID, client.UserID)
			}
			h.mu.Unlock()
		case msg := <-h.Broadcast:
			h.mu.Lock()
			fmt.Println("map:", h.Clients)
			client, ok := h.Clients[msg.ReceiverID]

			if !ok {
				log.Printf("client %s outline", msg.ReceiverID)
				continue
			}

			select {
				case client.Send <- msg.Message:
					log.Printf("cannot send %s\n", client.ClientID)
				default:
					close(client.Send)
					delete(h.Clients, client.ClientID)
				}
			

			h.mu.Unlock()
		}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingEvery)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)
			fmt.Println("message:", string(message))

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) ReadPump(hub *Hub) {
	defer func() {
		hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512)

	_ = c.Conn.SetReadDeadline(time.Now().Add(pingWait))

	c.Conn.SetPongHandler(func(string) error {
		_ = c.Conn.SetReadDeadline(time.Now().Add(pingWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}

		if len(message) > 0 {
			_ = c.Conn.SetReadDeadline(time.Now().Add(pingWait))
		}
	}
}

func (h *Hub) SendMessageToUser(userID string, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	h.Broadcast <- &MessageBroadcast{
		ReceiverID: userID,
		Message:    data,
	}

	return nil
}

func (h *Hub) GetOnlineUserCount(userID string) int {
	count := 0
	h.mu.RLock()
	for _, client := range h.Clients {
		if client.UserID == userID {
			count++
		}
	}
	h.mu.RUnlock()
	return count
}
