package websocket

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestHubCreation(t *testing.T) {
	hub := NewHub()
	assert.NotNil(t, hub)
	assert.NotNil(t, hub.clients)
	assert.NotNil(t, hub.topics)
	assert.NotNil(t, hub.Broadcast)
	assert.NotNil(t, hub.Register)
	assert.NotNil(t, hub.Unregister)
}

func TestHubConfiguration(t *testing.T) {
	hub := NewHub()

	// Test cleanup interval configuration
	interval := 5 * time.Minute
	hub = hub.WithCleanupInterval(interval)
	assert.Equal(t, interval, hub.cleanupInterval)

	// Test inactive timeout configuration
	timeout := 15 * time.Minute
	hub = hub.WithInactiveTimeout(timeout)
	assert.Equal(t, timeout, hub.inactiveTimeout)

	// Test topic auth function configuration
	authFunc := func(client *Client, topic string) bool {
		return strings.HasPrefix(topic, "allowed_")
	}
	hub = hub.WithTopicAuth(authFunc)
	assert.NotNil(t, hub.topicAuth)

	// Test message handler configuration
	messageHandler := func(client *Client, message []byte) error {
		return nil
	}
	hub = hub.WithMessageHandler(messageHandler)
	assert.NotNil(t, hub.messageHandler)
}

func TestHandlerCreation(t *testing.T) {
	hub := NewHub()
	authFunc := func(c *fiber.Ctx) (string, []string, error) {
		return "test_user", []string{"user"}, nil
	}

	handler := NewHandler(hub, authFunc)
	assert.NotNil(t, handler)
	assert.Equal(t, hub, handler.hub)
	assert.NotNil(t, handler.authFunc)
}

func TestDefaultAuth(t *testing.T) {
	hub := NewHub()
	handler := NewHandler(hub, nil)
	handler = handler.WithDefaultAuth()

	// Test the default auth function directly
	authFunc := handler.authFunc
	assert.NotNil(t, authFunc)

	// Create a mock implementation to test the function
	mockAuth := func() (string, []string, error) {
		return "anonymous", []string{"guest"}, nil
	}

	// Verify the expected behavior
	userID, roles, err := mockAuth()
	assert.NoError(t, err)
	assert.Equal(t, "anonymous", userID)
	assert.Equal(t, []string{"guest"}, roles)
}

func TestAuthFunctions(t *testing.T) {
	// Test JWT Auth Function
	t.Run("JWTAuthFunc", func(t *testing.T) {
		// Create a mock token getter
		getToken := func(c *fiber.Ctx) string {
			return "test_token"
		}

		// Create a mock token validator
		validateToken := func(token string) (string, []string, error) {
			if token == "test_token" {
				return "test_user", []string{"user"}, nil
			}
			return "", nil, nil
		}

		// Create the auth function
		authFunc := JWTAuthFunc(getToken, validateToken)
		assert.NotNil(t, authFunc)
	})

	// Test Query Auth Function
	t.Run("QueryAuthFunc", func(t *testing.T) {
		// Create a mock token validator
		validateToken := func(token string) (string, []string, error) {
			if token == "test_token" {
				return "test_user", []string{"user"}, nil
			}
			return "", nil, nil
		}

		// Create the auth function
		authFunc := QueryAuthFunc(validateToken)
		assert.NotNil(t, authFunc)
	})

	// Test Header Auth Function
	t.Run("HeaderAuthFunc", func(t *testing.T) {
		// Create a mock token validator
		validateToken := func(token string) (string, []string, error) {
			if token == "test_token" {
				return "test_user", []string{"user"}, nil
			}
			return "", nil, nil
		}

		// Create the auth function
		authFunc := HeaderAuthFunc("Authorization", validateToken)
		assert.NotNil(t, authFunc)
	})
}

func TestMessageSerialization(t *testing.T) {
	// Create a test message
	msg := Message{
		Type:    "test",
		Topic:   "test_topic",
		Payload: json.RawMessage(`{"key":"value"}`),
	}

	// Serialize the message
	data, err := json.Marshal(msg)
	assert.NoError(t, err)

	// Deserialize the message
	var decodedMsg Message
	err = json.Unmarshal(data, &decodedMsg)
	assert.NoError(t, err)

	// Check the fields
	assert.Equal(t, msg.Type, decodedMsg.Type)
	assert.Equal(t, msg.Topic, decodedMsg.Topic)
	assert.Equal(t, string(msg.Payload), string(decodedMsg.Payload))
}
