package handlers

import (
	"context"
	"encoding/json"
	"food-eats/cmd/web/db"
	"food-eats/cmd/web/model"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log/slog"
	"strconv"
	"sync"
	"time"
)

type RiderWebSocketReq struct {
	RiderId string `json:"rider_id"`
}

type NewLocationInfo struct {
	Message   string `json:"msg"`
	Latitude  string `json:"latitude" validate:"latitude"`
	Longitude string `json:"longitude" validate:"longitude"`
}

type LocationSyncReq struct {
	Latitude  string `json:"latitude" validate:"latitude"`
	Longitude string `json:"longitude" validate:"longitude"`
}

type AcceptOrderId struct {
	OrderId   primitive.ObjectID `json:"order_id"`
	Latitude  string             `json:"latitude" validate:"latitude"`
	Longitude string             `json:"longitude" validate:"longitude"`
}

type TrackOrder struct {
	OrderId primitive.ObjectID `json:"order_id"`
}

type RiderWebsocketMessage struct {
	Type string      `json:"type"`
	Body interface{} `json:"body"`
}

func ProcessRiderMessage(rm model.WebSocketManager, or OrderParam, riderId primitive.ObjectID, message []byte, ctx context.Context) {
	var msg RiderWebsocketMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		// Handle error
		return
	}

	messageType := msg.Type
	body, err := json.Marshal(msg.Body)
	if err != nil {
		return
	}

	switch messageType {
	case "accept_order":
		acceptReq := AcceptOrderId{}
		err = json.Unmarshal(body, &acceptReq)
		if err != nil {
			return
		}
		handleOrderAcceptance(ctx, rm, or, riderId, acceptReq)
	case "send_location":
		syncReq := LocationSyncReq{}
		err = json.Unmarshal(body, &syncReq)
		if err != nil {
			return
		}
		handleSendingLocation(ctx, rm, riderId, syncReq)
	case "delivered":
		acceptReq := AcceptOrderId{}
		err = json.Unmarshal(body, &acceptReq)
		if err != nil {
			return
		}
		handleOrderDelivered(ctx, rm, or, acceptReq)
	default:
		return
	}
}

func ProcessUserMessage(rm *model.SocketManager, riderId primitive.ObjectID, message []byte, ctx context.Context) {
	var msg RiderWebsocketMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		// Handle error
		return
	}

	messageType := msg.Type
	acceptReq, ok := msg.Body.(TrackOrder)
	if !ok {
		return
	}

	switch messageType {
	case "track_order":
		handleTrackOrder(rm, acceptReq, ctx)
	default:
		return
	}
}

func handleOrderAcceptance(ctx context.Context, rm model.WebSocketManager, or OrderParam, riderId primitive.ObjectID, acceptReq AcceptOrderId) {
	// Extract order ID from message
	orderId := acceptReq.OrderId
	if orderId.IsZero() {
		return
	}
	order, err := or.OrderRepo.GetOrder(ctx, orderId)
	if err != nil {
		slog.ErrorContext(ctx, "error in fetching order", "error", err.Error(), "order", order)
		go rm.BroadcastToRiders("error assigning", []primitive.ObjectID{riderId})
		return
	}

	if order.Status == "RIDER_ASSIGNED" {
		slog.Info("order already assigned")
		go rm.BroadcastToRiders("order already assigned", []primitive.ObjectID{riderId})
		return
	}
	redisConn, err := db.RedisConnFromPool()
	if err != nil {
		go rm.BroadcastToRiders("error assigning", []primitive.ObjectID{riderId})
		return
	}

	defer db.Close(redisConn)

	// handle concurrency
	result, err := redisConn.Incr(ctx, "order_status:"+orderId.Hex()).Result()
	if err != nil {
		go rm.BroadcastToRiders("error assigning", []primitive.ObjectID{riderId})
		return
	}

	if result != 1 {
		slog.Info("order already assigned")
		go rm.BroadcastToRiders("order already assigned", []primitive.ObjectID{riderId})
		return
	}

	currTime := time.Now()

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func(rConn *redis.Conn, rm model.WebSocketManager, wg *sync.WaitGroup) {
		defer wg.Done()

		members, err := rConn.SMembers(ctx, "rider_broadcasted:"+orderId.Hex()).Result()
		if err != nil {
			return
		}
		var rIds []primitive.ObjectID
		for _, member := range members {
			rId, err := primitive.ObjectIDFromHex(member)
			if err != nil {
				continue
			}
			rIds = append(rIds, rId)
		}
		newOrder := NewOrderBroadCast{
			Message: "Already picked order",
			OrderId: order.Id,
		}
		msg, err := json.Marshal(newOrder)
		if err != nil {
			return
		}
		rm.BroadcastToRiders(string(msg), rIds)

	}(redisConn, rm, &wg)

	order.RiderId = riderId
	order.Status = "RIDER_ASSIGNED"
	order.DeliveryStarted = currTime
	order.UpdatedAt = currTime
	order.Latitude, _ = strconv.ParseFloat(acceptReq.Latitude, 64)
	order.Longitude, _ = strconv.ParseFloat(acceptReq.Longitude, 64)

	// sending back the acknowledgement
	rm.BroadcastToRiders("Order accepted", []primitive.ObjectID{riderId})

	err = or.OrderRepo.UpdateOrder(ctx, order)
	if err != nil {
		slog.ErrorContext(ctx, "error in updating order ", "err", err.Error(), "order", "order")
		return
	}
	handleSendingLocation(ctx, rm, riderId, LocationSyncReq{acceptReq.Latitude, acceptReq.Longitude})
	wg.Wait()
}

func handleOrderDelivered(ctx context.Context, rm model.WebSocketManager, or OrderParam, acceptReq AcceptOrderId) {
	// Extract order ID from message
	orderId := acceptReq.OrderId
	if orderId.IsZero() {
		return
	}
	order, err := or.OrderRepo.GetOrder(ctx, orderId)
	if err != nil {
		slog.ErrorContext(ctx, "error in fetching order", "error", err.Error(), "order", order)
		return
	}

	if order.Status == "DELIVERED" {
		slog.Info("order already delivered")
		return
	}
	currTime := time.Now()

	order.Status = "DELIVERED"
	order.DeliveredAt = currTime
	order.UpdatedAt = currTime
	order.DeliveryTime = order.DeliveredAt.Sub(order.AcceptedAt).Seconds()
	order.Latitude, _ = strconv.ParseFloat(acceptReq.Latitude, 64)
	order.Longitude, _ = strconv.ParseFloat(acceptReq.Longitude, 64)

	err = or.OrderRepo.UpdateOrder(ctx, order)
	if err != nil {
		slog.ErrorContext(ctx, "update order failed", "error", err.Error())
		return
	}
	handleSendingDeliveredStatus(ctx, rm, order.UserId, LocationSyncReq{acceptReq.Latitude, acceptReq.Longitude})

	// running a go routine to update delivery time
	// can do it async via some queue
	go updateDeliveryTime(or, order)
}

func handleSendingLocation(ctx context.Context, rm model.WebSocketManager, userId primitive.ObjectID, req LocationSyncReq) {
	updateMessage := NewLocationInfo{
		Message:   "in between",
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
	}
	msg, err := json.Marshal(updateMessage)
	if err != nil {
		return
	}
	rm.BroadcastToUsers(string(msg), []primitive.ObjectID{userId})
}

func handleSendingDeliveredStatus(ctx context.Context, rm model.WebSocketManager, userId primitive.ObjectID, req LocationSyncReq) {
	updateMessage := NewLocationInfo{
		Message: "rider reached location",
	}
	msg, err := json.Marshal(updateMessage)
	if err != nil {
		return
	}
	rm.BroadcastToRiders(string(msg), []primitive.ObjectID{userId})
}

func handleTrackOrder(rm *model.SocketManager, trackOrder TrackOrder, ctx context.Context) {
	// todo no need of this
	//orderId := trackOrder.OrderId
	//if orderId.IsZero() {
	//	return
	//}
	//
	//order, err := rm.OrderRepo.GetOrder(ctx, orderId)
	//if err != nil {
	//	slog.ErrorContext(ctx, "error tracking order", err.Error())
	//	return
	//}
	//
	//userId := order.UserId
	//
	//rm.BroadcastToUsers()
}
