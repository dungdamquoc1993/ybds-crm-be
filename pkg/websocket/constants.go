package websocket

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512 * 1024 // 512KB

	// Buffer size for client send channel
	sendBufferSize = 256
)

var (
	newline = []byte{'\n'}
)

// generateID generates a random ID for clients
func generateID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return time.Now().String()
	}
	return hex.EncodeToString(b)
}
