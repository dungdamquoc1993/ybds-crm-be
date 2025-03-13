package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"
)

// Client represents a WebSocket client connection
type Client struct {
	ID           string
	Conn         *websocket.Conn
	Hub          *Hub
	Send         chan []byte
	UserID       string
	Roles        []string
	Topics       map[string]bool
	LastActivity time.Time
	mu           sync.Mutex
}

// Message represents a WebSocket message
type Message struct {
	Type    string          `json:"type"`
	Topic   string          `json:"topic,omitempty"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// NewClient creates a new WebSocket client
func NewClient(conn *websocket.Conn, hub *Hub, userID string, roles []string) *Client {
	return &Client{
		ID:           generateID(),
		Conn:         conn,
		Hub:          hub,
		Send:         make(chan []byte, sendBufferSize),
		UserID:       userID,
		Roles:        roles,
		Topics:       make(map[string]bool),
		LastActivity: time.Now(),
		mu:           sync.Mutex{},
	}
}

// ReadPump pumps messages from the WebSocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		c.updateActivity()
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		// Update last activity time
		c.updateActivity()

		// Process the message
		c.processMessage(message)
	}
}

// WritePump pumps messages from the hub to the WebSocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current websocket message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.Send)
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

// Subscribe adds a topic to the client's subscriptions
func (c *Client) Subscribe(topic string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Topics[topic] = true
}

// Unsubscribe removes a topic from the client's subscriptions
func (c *Client) Unsubscribe(topic string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.Topics, topic)
}

// IsSubscribed checks if the client is subscribed to a topic
func (c *Client) IsSubscribed(topic string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Topics[topic]
}

// HasRole checks if the client has a specific role
func (c *Client) HasRole(role string) bool {
	for _, r := range c.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasAnyRole checks if the client has any of the specified roles
func (c *Client) HasAnyRole(roles ...string) bool {
	for _, role := range roles {
		if c.HasRole(role) {
			return true
		}
	}
	return false
}

// updateActivity updates the client's last activity timestamp
func (c *Client) updateActivity() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.LastActivity = time.Now()
}

// processMessage processes incoming WebSocket messages
func (c *Client) processMessage(data []byte) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("error unmarshaling message: %v", err)
		return
	}

	switch msg.Type {
	case "subscribe":
		if msg.Topic != "" {
			// Check if the client is authorized to subscribe to this topic
			if c.Hub.CanSubscribe(c, msg.Topic) {
				c.Subscribe(msg.Topic)
				// Send confirmation
				c.sendSubscriptionConfirmation(msg.Topic, true)
			} else {
				// Send unauthorized message
				c.sendSubscriptionConfirmation(msg.Topic, false)
			}
		}
	case "unsubscribe":
		if msg.Topic != "" {
			c.Unsubscribe(msg.Topic)
			// Send confirmation
			c.sendUnsubscriptionConfirmation(msg.Topic)
		}
	default:
		// Forward the message to the hub for processing
		c.Hub.Broadcast <- &BroadcastMessage{
			Client:  c,
			Message: data,
		}
	}
}

// sendSubscriptionConfirmation sends a subscription confirmation message
func (c *Client) sendSubscriptionConfirmation(topic string, success bool) {
	response := struct {
		Type    string `json:"type"`
		Topic   string `json:"topic"`
		Success bool   `json:"success"`
	}{
		Type:    "subscription_status",
		Topic:   topic,
		Success: success,
	}

	data, err := json.Marshal(response)
	if err != nil {
		log.Printf("error marshaling subscription confirmation: %v", err)
		return
	}

	c.Send <- data
}

// sendUnsubscriptionConfirmation sends an unsubscription confirmation message
func (c *Client) sendUnsubscriptionConfirmation(topic string) {
	response := struct {
		Type  string `json:"type"`
		Topic string `json:"topic"`
	}{
		Type:  "unsubscribed",
		Topic: topic,
	}

	data, err := json.Marshal(response)
	if err != nil {
		log.Printf("error marshaling unsubscription confirmation: %v", err)
		return
	}

	c.Send <- data
}
