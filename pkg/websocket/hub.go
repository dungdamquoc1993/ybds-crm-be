package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"
)

// BroadcastMessage represents a message to be broadcast
type BroadcastMessage struct {
	Client  *Client
	Message []byte
	Topic   string
}

// TopicAuthFunc is a function that checks if a client can subscribe to a topic
type TopicAuthFunc func(client *Client, topic string) bool

// MessageHandlerFunc is a function that handles incoming messages
type MessageHandlerFunc func(client *Client, message []byte) error

// Hub maintains the set of active clients and broadcasts messages to them
type Hub struct {
	// Registered clients
	clients map[string]*Client

	// Topics and their subscribers
	topics map[string]map[string]*Client

	// Inbound messages from the clients
	Broadcast chan *BroadcastMessage

	// Register requests from the clients
	Register chan *Client

	// Unregister requests from clients
	Unregister chan *Client

	// Topic authorization function
	topicAuth TopicAuthFunc

	// Message handler function
	messageHandler MessageHandlerFunc

	// Mutex for concurrent access
	mu sync.RWMutex

	// Inactive client cleanup interval
	cleanupInterval time.Duration

	// Inactive timeout
	inactiveTimeout time.Duration
}

// NewHub creates a new hub
func NewHub() *Hub {
	return &Hub{
		Broadcast:       make(chan *BroadcastMessage),
		Register:        make(chan *Client),
		Unregister:      make(chan *Client),
		clients:         make(map[string]*Client),
		topics:          make(map[string]map[string]*Client),
		mu:              sync.RWMutex{},
		cleanupInterval: 10 * time.Minute,
		inactiveTimeout: 30 * time.Minute,
		topicAuth:       defaultTopicAuth,
		messageHandler:  defaultMessageHandler,
	}
}

// WithCleanupInterval sets the cleanup interval
func (h *Hub) WithCleanupInterval(interval time.Duration) *Hub {
	h.cleanupInterval = interval
	return h
}

// WithInactiveTimeout sets the inactive timeout
func (h *Hub) WithInactiveTimeout(timeout time.Duration) *Hub {
	h.inactiveTimeout = timeout
	return h
}

// WithTopicAuth sets the topic authorization function
func (h *Hub) WithTopicAuth(authFunc TopicAuthFunc) *Hub {
	h.topicAuth = authFunc
	return h
}

// WithMessageHandler sets the message handler function
func (h *Hub) WithMessageHandler(handlerFunc MessageHandlerFunc) *Hub {
	h.messageHandler = handlerFunc
	return h
}

// Run starts the hub
func (h *Hub) Run() {
	// Start the inactive client cleanup
	go h.cleanupInactiveClients()

	for {
		select {
		case client := <-h.Register:
			h.registerClient(client)
		case client := <-h.Unregister:
			h.unregisterClient(client)
		case message := <-h.Broadcast:
			h.handleBroadcast(message)
		}
	}
}

// registerClient registers a client with the hub
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[client.ID] = client
}

// unregisterClient unregisters a client from the hub
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Remove from clients map
	if _, ok := h.clients[client.ID]; ok {
		delete(h.clients, client.ID)
		close(client.Send)
	}

	// Remove from all topics
	for topic := range client.Topics {
		if topicClients, ok := h.topics[topic]; ok {
			delete(topicClients, client.ID)
			// If topic has no subscribers, remove it
			if len(topicClients) == 0 {
				delete(h.topics, topic)
			}
		}
	}
}

// handleBroadcast handles a broadcast message
func (h *Hub) handleBroadcast(bm *BroadcastMessage) {
	// If a topic is specified, only send to subscribers of that topic
	if bm.Topic != "" {
		h.broadcastToTopic(bm.Topic, bm.Message)
		return
	}

	// Otherwise, try to parse the message to determine the topic
	var msg Message
	if err := json.Unmarshal(bm.Message, &msg); err == nil && msg.Topic != "" {
		h.broadcastToTopic(msg.Topic, bm.Message)
		return
	}

	// If no topic is specified, call the message handler
	if h.messageHandler != nil && bm.Client != nil {
		if err := h.messageHandler(bm.Client, bm.Message); err != nil {
			log.Printf("error handling message: %v", err)
		}
	}
}

// broadcastToTopic broadcasts a message to all subscribers of a topic
func (h *Hub) broadcastToTopic(topic string, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if topicClients, ok := h.topics[topic]; ok {
		for _, client := range topicClients {
			select {
			case client.Send <- message:
			default:
				h.Unregister <- client
			}
		}
	}
}

// BroadcastToAll broadcasts a message to all connected clients
func (h *Hub) BroadcastToAll(message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.clients {
		select {
		case client.Send <- message:
		default:
			h.Unregister <- client
		}
	}
}

// BroadcastToUser broadcasts a message to a specific user
func (h *Hub) BroadcastToUser(userID string, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.clients {
		if client.UserID == userID {
			select {
			case client.Send <- message:
			default:
				h.Unregister <- client
			}
		}
	}
}

// BroadcastToRole broadcasts a message to clients with a specific role
func (h *Hub) BroadcastToRole(role string, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.clients {
		if client.HasRole(role) {
			select {
			case client.Send <- message:
			default:
				h.Unregister <- client
			}
		}
	}
}

// CanSubscribe checks if a client can subscribe to a topic
func (h *Hub) CanSubscribe(client *Client, topic string) bool {
	if h.topicAuth != nil {
		return h.topicAuth(client, topic)
	}
	return true
}

// cleanupInactiveClients periodically removes inactive clients
func (h *Hub) cleanupInactiveClients() {
	ticker := time.NewTicker(h.cleanupInterval)
	defer ticker.Stop()

	for {
		<-ticker.C
		h.mu.Lock()
		now := time.Now()
		for id, client := range h.clients {
			if now.Sub(client.LastActivity) > h.inactiveTimeout {
				log.Printf("removing inactive client: %s", id)
				delete(h.clients, id)
				close(client.Send)
				// Remove from all topics
				for topic := range client.Topics {
					if topicClients, ok := h.topics[topic]; ok {
						delete(topicClients, client.ID)
						// If topic has no subscribers, remove it
						if len(topicClients) == 0 {
							delete(h.topics, topic)
						}
					}
				}
			}
		}
		h.mu.Unlock()
	}
}

// defaultTopicAuth is the default topic authorization function
func defaultTopicAuth(client *Client, topic string) bool {
	// By default, allow all subscriptions
	return true
}

// defaultMessageHandler is the default message handler function
func defaultMessageHandler(client *Client, message []byte) error {
	// By default, do nothing
	return nil
}
