package websocket

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"bookmark-sync-service/backend/pkg/redis"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// Hub maintains the set of active clients and broadcasts messages to the clients
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Inbound messages from the clients
	broadcast chan []byte

	// Register requests from the clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Redis client for pub/sub
	redisClient *redis.Client

	// Sync service for handling sync messages
	syncService SyncService

	// Logger
	logger *zap.Logger

	// Mutex for thread safety
	mutex sync.RWMutex
}

// Client is a middleman between the websocket connection and the hub
type Client struct {
	// The websocket connection
	conn *websocket.Conn

	// Buffered channel of outbound messages
	send chan []byte

	// User ID for this client
	userID string

	// Device ID for this client
	deviceID string

	// Hub reference
	hub *Hub

	// Logger
	logger *zap.Logger
}

// Message represents a WebSocket message
type Message struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	UserID    string      `json:"user_id,omitempty"`
	DeviceID  string      `json:"device_id,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

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

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// TODO: Implement proper origin checking
		return true
	},
}

// NewHub creates a new WebSocket hub
func NewHub(redisClient *redis.Client, logger *zap.Logger) *Hub {
	return &Hub{
		clients:     make(map[*Client]bool),
		broadcast:   make(chan []byte),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		redisClient: redisClient,
		logger:      logger,
	}
}

// NewHubWithSyncService creates a new WebSocket hub with sync service integration
func NewHubWithSyncService(redisClient *redis.Client, syncService SyncService, logger *zap.Logger) *Hub {
	return &Hub{
		clients:     make(map[*Client]bool),
		broadcast:   make(chan []byte),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		redisClient: redisClient,
		syncService: syncService,
		logger:      logger,
	}
}

// Run starts the hub
func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			h.mutex.Unlock()

			h.logger.Info("Client connected",
				zap.String("user_id", client.userID),
				zap.String("device_id", client.deviceID),
			)

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mutex.Unlock()

			h.logger.Info("Client disconnected",
				zap.String("user_id", client.userID),
				zap.String("device_id", client.deviceID),
			)

		case message := <-h.broadcast:
			h.mutex.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mutex.RUnlock()

		case <-ctx.Done():
			h.logger.Info("WebSocket hub shutting down")
			return
		}
	}
}

// BroadcastToUser sends a message to all clients of a specific user
func (h *Hub) BroadcastToUser(userID string, message *Message) {
	messageBytes, err := json.Marshal(message)
	if err != nil {
		h.logger.Error("Failed to marshal message", zap.Error(err))
		return
	}

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for client := range h.clients {
		if client.userID == userID {
			select {
			case client.send <- messageBytes:
			default:
				close(client.send)
				delete(h.clients, client)
			}
		}
	}
}

// HandleWebSocket handles WebSocket connections
func (h *Hub) HandleWebSocket(c *gin.Context) {
	// Extract user ID and device ID from query parameters or headers
	userID := c.Query("user_id")
	deviceID := c.Query("device_id")

	if userID == "" || deviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id and device_id are required"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade connection", zap.Error(err))
		return
	}

	client := &Client{
		conn:     conn,
		send:     make(chan []byte, 256),
		userID:   userID,
		deviceID: deviceID,
		hub:      h,
		logger:   h.logger.With(zap.String("user_id", userID), zap.String("device_id", deviceID)),
	}

	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in new goroutines
	go client.writePump()
	go client.readPump()
}

// readPump pumps messages from the websocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.logger.Error("WebSocket error", zap.Error(err))
			}
			break
		}

		// Process incoming message
		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			c.logger.Error("Failed to unmarshal message", zap.Error(err))
			continue
		}

		// Add metadata
		msg.UserID = c.userID
		msg.DeviceID = c.deviceID
		msg.Timestamp = time.Now()

		// Handle different message types
		c.handleMessage(&msg)
	}
}

// writePump pumps messages from the hub to the websocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// SyncService interface for WebSocket integration
type SyncService interface {
	HandleSyncMessage(ctx context.Context, msg *SyncMessage) (*SyncMessage, error)
}

// SyncMessage represents a sync message for WebSocket communication
type SyncMessage struct {
	Type      string                 `json:"type"`
	UserID    string                 `json:"user_id,omitempty"`
	DeviceID  string                 `json:"device_id,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// handleMessage processes incoming WebSocket messages
func (c *Client) handleMessage(msg *Message) {
	ctx := context.Background()

	switch msg.Type {
	case "ping":
		// Respond with pong
		response := &Message{
			Type:      "pong",
			Timestamp: time.Now(),
		}
		c.sendMessage(response)

	case "sync_request":
		// Handle sync request
		c.logger.Info("Sync request received", zap.String("type", msg.Type))

		if c.hub.syncService != nil {
			// Convert to sync message format
			var data map[string]interface{}
			if msg.Data != nil {
				if d, ok := msg.Data.(map[string]interface{}); ok {
					data = d
				} else {
					data = make(map[string]interface{})
				}
			} else {
				data = make(map[string]interface{})
			}

			syncMsg := &SyncMessage{
				Type:      "sync_request",
				UserID:    c.userID,
				DeviceID:  c.deviceID,
				Data:      data,
				Timestamp: time.Now(),
			}

			// Handle sync request through sync service
			syncResponse, err := c.hub.syncService.HandleSyncMessage(ctx, syncMsg)
			if err != nil {
				c.logger.Error("Failed to handle sync request", zap.Error(err))
				response := &Message{
					Type:      "sync_error",
					Data:      map[string]interface{}{"error": err.Error()},
					Timestamp: time.Now(),
				}
				c.sendMessage(response)
				return
			}

			// Convert sync response back to WebSocket message
			response := &Message{
				Type:      "sync_response",
				Data:      syncResponse.Data,
				Timestamp: time.Now(),
			}
			c.sendMessage(response)
		} else {
			// Fallback response when sync service is not available
			response := &Message{
				Type:      "sync_response",
				Data:      map[string]interface{}{"status": "sync_service_unavailable"},
				Timestamp: time.Now(),
			}
			c.sendMessage(response)
		}

	case "sync_event":
		// Handle sync event
		c.logger.Info("Sync event received",
			zap.String("type", msg.Type),
			zap.String("user_id", c.userID),
			zap.String("device_id", c.deviceID),
		)

		// For sync events, we typically just acknowledge receipt
		// The actual processing would be handled by the sync service
		response := &Message{
			Type:      "sync_event_ack",
			Data:      map[string]interface{}{"status": "received"},
			Timestamp: time.Now(),
		}
		c.sendMessage(response)

	default:
		c.logger.Warn("Unknown message type", zap.String("type", msg.Type))

		// Send error response for unknown message types
		response := &Message{
			Type:      "error",
			Data:      map[string]interface{}{"error": "unknown_message_type", "received_type": msg.Type},
			Timestamp: time.Now(),
		}
		c.sendMessage(response)
	}
}

// sendMessage sends a message to the client
func (c *Client) sendMessage(msg *Message) {
	messageBytes, err := json.Marshal(msg)
	if err != nil {
		c.logger.Error("Failed to marshal message", zap.Error(err))
		return
	}

	select {
	case c.send <- messageBytes:
	default:
		close(c.send)
	}
}
