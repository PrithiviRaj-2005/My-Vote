package websocket

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512
)

// upgrader configures WebSocket handshake and allows all cross-origin requests for polling clients
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for public polling access
	},
}

// Client represents a single active WebSocket connection
type Client struct {
	hub       *Hub
	conn      *websocket.Conn
	shareCode string
	send      chan []byte
}

// Hub maintains the set of active clients grouped by poll shareCode
type Hub struct {
	// Map of shareCode -> map of active clients
	clients map[string]map[*Client]bool

	// Mutex to protect concurrent access to clients map
	mu sync.RWMutex

	// Register channel
	register chan *Client

	// Unregister channel
	unregister chan *Client

	// Broadcast channel for poll messages
	broadcast chan *BroadcastMessage
}

// BroadcastMessage encapsulates data to be sent to all clients of a given poll
type BroadcastMessage struct {
	ShareCode string
	Data      []byte
}

// NewHub initializes a new WebSocket Hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *BroadcastMessage, 256),
	}
}

// Run starts the central event loop for the Hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if _, ok := h.clients[client.shareCode]; !ok {
				h.clients[client.shareCode] = make(map[*Client]bool)
			}
			h.clients[client.shareCode][client] = true
			totalConnected := len(h.clients[client.shareCode])
			h.mu.Unlock()
			log.Printf("[WebSocket] Client connected to poll [%s]. Total active for this poll: %d", client.shareCode, totalConnected)

		case client := <-h.unregister:
			h.mu.Lock()
			if clientsForPoll, ok := h.clients[client.shareCode]; ok {
				if _, ok := clientsForPoll[client]; ok {
					delete(clientsForPoll, client)
					close(client.send)
					if len(clientsForPoll) == 0 {
						delete(h.clients, client.shareCode)
					}
				}
			}
			h.mu.Unlock()
			log.Printf("[WebSocket] Client disconnected from poll [%s]", client.shareCode)

		case message := <-h.broadcast:
			h.mu.RLock()
			clientsForPoll, ok := h.clients[message.ShareCode]
			if ok {
				for client := range clientsForPoll {
					select {
					case client.send <- message.Data:
					default:
						// Send buffer full, close and mark for cleanup
						close(client.send)
						delete(clientsForPoll, client)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast sends a message to all connected clients viewing a specific poll
func (h *Hub) Broadcast(shareCode string, data []byte) {
	h.broadcast <- &BroadcastMessage{
		ShareCode: shareCode,
		Data:      data,
	}
}

// readPump pumps messages from the websocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		_ = c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		// Read messages (audience may send heartbeats or client pings)
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[WebSocket] Read error: %v", err)
			}
			break
		}
	}
}

// writePump pumps messages from the hub to the websocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			_, _ = w.Write(message)

			// Add queued chat messages to the current websocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				_, _ = w.Write([]byte{'\n'})
				_, _ = w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ServeWs handles websocket requests from the peer
func ServeWs(hub *Hub, c *gin.Context) {
	shareCode := c.Param("shareCode")
	if shareCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Share code is required"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[WebSocket] Failed to upgrade connection: %v", err)
		return
	}

	client := &Client{
		hub:       hub,
		conn:      conn,
		shareCode: shareCode,
		send:      make(chan []byte, 256),
	}

	client.hub.register <- client

	// Start concurrent read and write pumps
	go client.writePump()
	go client.readPump()
}
