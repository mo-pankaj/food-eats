package routes

import (
	"crypto/rand"
	"encoding/base64"
	"food-eats/cmd/web/model"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWsApplication_ConnectUserWebSocket(t *testing.T) {
	key := make([]byte, 16)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("error generating random key: %v", err)
	}
	secWebSocketKey := base64.StdEncoding.EncodeToString(key)

	// Create a new echo context with the required headers
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "upgrade")
	req.Header.Set("Sec-WebSocket-Key", secWebSocketKey)
	req.Header.Set("Sec-WebSocket-Version", "13")
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	// Set up the fields with necessary dependencies/mock objects for the test
	fields := struct {
		mongodb *mongo.Database
		sM      model.WebSocketManager
	}{
		// Initialize your dependencies/mock objects here
		mongodb: nil,
		sM:      nil,
	}

	ua := &WsApplication{
		Mongodb: fields.mongodb,
		SM:      fields.sM,
	}

	// Call the method being tested
	err := ua.ConnectUserWebSocket(c)

	// Check if the error matches the expectation
	wantErr := false // Expecting no error
	if (err != nil) != wantErr {
		t.Errorf("ConnectUserWebSocket() error = %v, wantErr %v", err, wantErr)
	}
}
