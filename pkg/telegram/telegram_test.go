package telegram

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSendMessage(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the request method is POST
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Check if the Content-Type is application/json
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		// Return a successful response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok": true, "result": {"message_id": 1}}`))
	}))
	defer server.Close()

	// Save the real API URL and replace it for the test
	actualURL := "https://api.telegram.org/bot%s/sendMessage"
	telegramAPIURL = server.URL + "/%s/sendMessage"
	defer func() {
		telegramAPIURL = actualURL
	}()

	// Create a client with the test token
	client := NewClient("test_token")

	// Test sending a message
	err := client.SendMessage(123456789, "Test message")
	if err != nil {
		t.Errorf("SendMessage returned an error: %v", err)
	}
}

func TestSendMessageError(t *testing.T) {
	// Create a mock server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"ok": false, "description": "Bad Request: chat not found"}`))
	}))
	defer server.Close()

	// Save the real API URL and replace it for the test
	actualURL := "https://api.telegram.org/bot%s/sendMessage"
	telegramAPIURL = server.URL + "/%s/sendMessage"
	defer func() {
		telegramAPIURL = actualURL
	}()

	// Create a client with the test token
	client := NewClient("test_token")

	// Test sending a message
	err := client.SendMessage(123456789, "Test message")
	if err == nil {
		t.Error("SendMessage did not return an error when expected")
	}
}
