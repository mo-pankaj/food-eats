package routes

import (
	"encoding/json"
	"food-eats/cmd/web/handlers"
	"food-eats/cmd/web/model"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"log/slog"
	"net/http"
)

type WsApplication struct {
	Mongodb   *mongo.Database
	SM        model.WebSocketManager
	RedisConn *redis.Conn
	OR        handlers.OrderParam
}

func NewWsApplication(mongodb *mongo.Database, sm model.WebSocketManager) *WsApplication {
	return &WsApplication{
		Mongodb: mongodb,
		SM:      sm,
	}
}

// to convert a tcp to web socket connection
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow all connections
		return true
	},
}

func (ua *WsApplication) ConnectRiderWebSocket(c echo.Context) error {
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Read the first message to get the rider_id
	_, msg, err := conn.ReadMessage()
	if err != nil {
		return err
	}

	// Parse the first message to get the rider_id
	var firstMsg handlers.RiderWebSocketReq
	if err := json.Unmarshal(msg, &firstMsg); err != nil {
		return err
	}

	// Validate the rider_id
	if firstMsg.RiderId == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Rider ID is required")
	}

	// Convert RiderId to ObjectID
	oId, err := primitive.ObjectIDFromHex(firstMsg.RiderId)
	if err != nil {
		return err
	}

	// mongo dependencies
	restaurantRepo := model.RestaurantRepository(model.RestaurantMongoRepo(ua.Mongodb))
	riderRepo := model.RiderRepository(model.RiderMongoRepo(ua.Mongodb))
	orderRepo := model.OrderRepository(model.OrderMongoRepo(ua.Mongodb))

	orderParam := handlers.OrderParam{
		OrderRepo:      model.OrderRepository(orderRepo),
		RestaurantRepo: restaurantRepo,
		RiderRepo:      riderRepo,
	}
	ua.OR = orderParam

	// Register the rider with WebSocket client
	ua.SM.RegisterRider(oId, model.NewWebSocketClient(conn))

	ctx := c.Request().Context()

	for {
		// Read subsequent messages from WebSocket
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}
		slog.Info("Received message", "msg", msg)
		handlers.ProcessRiderMessage(ua.SM, ua.OR, oId, msg, ctx)
	}

	return nil
}

func (ua *WsApplication) ConnectUserWebSocket(c echo.Context) error {
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	req := new(handlers.RiderWebSocketReq)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := c.Validate(req); err != nil {
		return err
	}

	oId, err := primitive.ObjectIDFromHex(req.RiderId)
	if err != nil {
		return err
	}

	ua.SM.RegisterUser(oId, model.NewWebSocketClient(conn))

	//ctx := c.Request().Context()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}
		slog.Info("Received message:", "msg", msg)
		// no need of user message consumption
		//handlers.ProcessUserMessage(ua.SM, req.RiderId, msg, ctx)
	}

	return nil
}
