// Package websocket provides real-time bidirectional communication for the PMS.
package websocket

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Hub manages all active WebSocket connections.
type Hub struct {
	clients    map[string]*Client // tenantID -> client
	broadcast  chan Message
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

// NewHub creates a new WebSocket hub.
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		broadcast:  make(chan Message, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Message is a WebSocket broadcast message.
type Message struct {
	TenantID string      `json:"-"`
	Type     string      `json:"type"`
	Payload  interface{} `json:"payload"`
	Timestamp time.Time  `json:"timestamp"`
}

// Client represents a single WebSocket connection.
type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	tenantID string
	userID   string
	roles    []string
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // In production, validate origin
	},
}

// Run starts the hub event loop.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.tenantID+":"+client.userID] = client
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.tenantID+":"+client.userID]; ok {
				delete(h.clients, client.tenantID+":"+client.userID)
				close(client.send)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for id, client := range h.clients {
				if id == message.TenantID+":"+client.userID || client.tenantID == message.TenantID {
					select {
					case client.send <- h.marshal(message):
					default:
						// Client buffer full, drop message
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) marshal(msg Message) []byte {
	b, _ := json.Marshal(msg)
	return b
}

// Broadcast sends a message to all clients in a tenant.
func (h *Hub) Broadcast(tenantID string, msgType string, payload interface{}) {
	h.broadcast <- Message{
		TenantID:  tenantID,
		Type:      msgType,
		Payload:   payload,
		Timestamp: time.Now().UTC(),
	}
}

// ServeWS handles WebSocket upgrade requests.
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request, tenantID, userID string, roles []string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := &Client{
		hub:      h,
		conn:     conn,
		send:     make(chan []byte, 256),
		tenantID: tenantID,
		userID:   userID,
		roles:    roles,
	}

	h.register <- client

	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				// Log unexpected close
			}
			break
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.conn.WriteMessage(websocket.TextMessage, message)

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// RoomStatusUpdate sends a real-time room status change.
func (h *Hub) RoomStatusUpdate(tenantID, roomID, roomNumber, status string) {
	h.Broadcast(tenantID, "room_status_changed", map[string]string{
		"room_id":     roomID,
		"room_number": roomNumber,
		"status":      status,
	})
}

// ReservationUpdate sends a real-time reservation event.
func (h *Hub) ReservationUpdate(tenantID, reservationID, eventType string) {
	h.Broadcast(tenantID, "reservation_"+eventType, map[string]string{
		"reservation_id": reservationID,
		"event":          eventType,
	})
}

// HousekeepingTaskUpdate sends a real-time housekeeping event.
func (h *Hub) HousekeepingTaskUpdate(tenantID, taskID, status string) {
	h.Broadcast(tenantID, "housekeeping_task_changed", map[string]string{
		"task_id": taskID,
		"status":  status,
	})
}

// LockEvent sends a real-time smart lock event.
func (h *Hub) LockEvent(tenantID, lockID, eventType string) {
	h.Broadcast(tenantID, "lock_event", map[string]string{
		"lock_id": lockID,
		"event":   eventType,
	})
}

// NewMessage sends a real-time guest message.
func (h *Hub) NewMessage(tenantID, messageID, guestID string) {
	h.Broadcast(tenantID, "new_message", map[string]string{
		"message_id": messageID,
		"guest_id":   guestID,
	})
}

// DigitalKeyUpdate sends a real-time digital key event.
func (h *Hub) DigitalKeyUpdate(tenantID, keyID, reservationID, status string) {
	h.Broadcast(tenantID, "digital_key_changed", map[string]string{
		"key_id":         keyID,
		"reservation_id": reservationID,
		"status":         status,
	})
}
