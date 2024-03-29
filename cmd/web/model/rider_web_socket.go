package model

import (
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log/slog"
	"sync"
)

// WebSocketClient represents a WebSocket client connection.
type WebSocketClient struct {
	ws *websocket.Conn
}

// NewWebSocketClient creates a new WebSocketClient instance.
func NewWebSocketClient(ws *websocket.Conn) *WebSocketClient {
	return &WebSocketClient{
		ws: ws,
	}
}

// WebSocketManager defines the interface for managing WebSocket connections.
type WebSocketManager interface {
	RegisterRider(riderID primitive.ObjectID, conn *WebSocketClient)
	RegisterUser(userID primitive.ObjectID, conn *WebSocketClient)
	UnregisterRider(riderID primitive.ObjectID)
	UnregisterUser(userID primitive.ObjectID)
	BroadcastToRiders(message string, riderIDs []primitive.ObjectID)
	BroadcastToUsers(message string, userIDs []primitive.ObjectID)
}

// SocketManager is responsible for managing WebSocket connections.
type SocketManager struct {
	riderClients map[string]*WebSocketClient
	userClients  map[string]*WebSocketClient
	lock         sync.Mutex
	OrderRepo    OrderRepository
}

// NewWebSocketManager creates a new SocketManager instance.
func NewWebSocketManager(database *mongo.Database) *SocketManager {
	return &SocketManager{
		riderClients: make(map[string]*WebSocketClient),
		userClients:  make(map[string]*WebSocketClient),
		OrderRepo:    OrderRepository(OrderMongoRepo(database)),
	}
}

// RegisterRider registers a WebSocket connection for a rider.
func (wm *SocketManager) RegisterRider(riderID primitive.ObjectID, conn *WebSocketClient) {
	wm.lock.Lock()
	defer wm.lock.Unlock()
	wm.riderClients[riderID.Hex()] = conn
}

// RegisterUser registers a WebSocket connection for a user.
func (wm *SocketManager) RegisterUser(userID primitive.ObjectID, conn *WebSocketClient) {
	wm.lock.Lock()
	defer wm.lock.Unlock()
	wm.userClients[userID.Hex()] = conn
}

// UnregisterRider unregisters a WebSocket connection for a rider.
func (wm *SocketManager) UnregisterRider(riderID primitive.ObjectID) {
	wm.lock.Lock()
	defer wm.lock.Unlock()
	delete(wm.riderClients, riderID.Hex())
}

// UnregisterUser unregisters a WebSocket connection for a user.
func (wm *SocketManager) UnregisterUser(userID primitive.ObjectID) {
	wm.lock.Lock()
	defer wm.lock.Unlock()
	delete(wm.userClients, userID.Hex())
}

// BroadcastToRiders sends a message to WebSocket connections of riders.
func (wm *SocketManager) BroadcastToRiders(message string, riderIDs []primitive.ObjectID) {
	wm.lock.Lock()
	defer wm.lock.Unlock()
	for _, riderID := range riderIDs {
		conn := wm.riderClients[riderID.Hex()]
		if conn == nil {
			continue
		}
		if err := conn.ws.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
			slog.Error("error sending message to rider connection", "err", err)
		}
	}
}

// BroadcastToUsers sends a message to WebSocket connections of users.
func (wm *SocketManager) BroadcastToUsers(message string, userIDs []primitive.ObjectID) {
	wm.lock.Lock()
	defer wm.lock.Unlock()
	for _, userID := range userIDs {
		conn := wm.userClients[userID.Hex()]
		if conn == nil {
			continue
		}
		if err := conn.ws.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
			slog.Error("error sending message to user connection")
		}
	}
}
