package model

import (
	"context"
	"errors"
	errors2 "food-eats/cmd/web/custom-errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

// Order holds information for an order
type Order struct {
	Id           primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`           // omitempty so a mongo-driver can generate unique id
	UserId       primitive.ObjectID `json:"user_id" bson:"userId"`                       // index
	RiderId      primitive.ObjectID `json:"rider_id,omitempty" bson:"riderId,omitempty"` // index
	RestaurantId primitive.ObjectID `json:"restaurant_id" bson:"restaurantId"`           // index

	// Delivery Information
	DeliveryPhoneNumber string  `json:"delivery_phone_number" bson:"deliveryPhoneNumber"`
	DeliveryLatitude    float64 `json:"delivery_latitude" bson:"deliveryLatitude"`
	DeliveryLongitude   float64 `json:"delivery_longitude" bson:"deliveryLongitude"`
	DeliveryAddress     string  `json:"address" bson:"address"`

	// Default address fields
	PickupPhoneNumber string  `json:"pickup_phone_number" bson:"pickupPhoneNumber"`
	PickupLatitude    float64 `json:"pickup_latitude" bson:"pickupLatitude"`
	PickupLongitude   float64 `json:"pickup_longitude" bson:"pickupLongitude"`
	PickupAddress     string  `json:"pickup_address" bson:"pickupAddress"`

	Items      []Item  `json:"items" bson:"items"`
	FinalPrice float32 `json:"final_price" bson:"finalPrice"`

	// Live location fields
	Latitude  float64 `json:"latitude" bson:"latitude"`
	Longitude float64 `json:"longitude" bson:"longitude"`

	DeliveryTime float64 `json:"delivery_time" bson:"deliveryTime"` // in seconds

	CreatedAt       time.Time `json:"created_at" bson:"createdAt"`
	AcceptedAt      time.Time `json:"accepted_at,omitempty" bson:"acceptedAt,omitempty"`
	UpdatedAt       time.Time `json:"updatedAt" bson:"updatedAt"`
	DeliveryStarted time.Time `json:"delivery_started,omitempty" bson:"deliveryStarted,omitempty"`
	DeliveredAt     time.Time `json:"delivered_at,omitempty" bson:"deliveredAt,omitempty"`

	Status string `json:"status" bson:"status"`
}

// OrderRepository will be the order repository, a database needs to implement this contract
type OrderRepository interface {
	CreateOrder(ctx context.Context, order Order) (Order, error)
	UpdateOrder(ctx context.Context, order Order) error
	GetOrder(ctx context.Context, id primitive.ObjectID) (Order, error)
	SearchOrder(ctx context.Context, query SearchOrderQuery) ([]Order, int64, error)
}

// OrderMongo type with embedded mongo.Database
type OrderMongo struct {
	DB *mongo.Database
}

type SearchOrderQuery struct {
	OrderId      primitive.ObjectID
	RestaurantId primitive.ObjectID
	DriverId     primitive.ObjectID
	UserId       primitive.ObjectID
	Status       string
	Limit        int
	Skip         int
}

func OrderMongoRepo(DB *mongo.Database) OrderMongo {
	return OrderMongo{DB: DB}
}

func (u OrderMongo) CreateOrder(ctx context.Context, order Order) (Order, error) {
	insertedId, err := u.DB.Collection("Order").InsertOne(ctx, order) // ignoring the parameter as we don't require it
	if err != nil {
		return Order{}, err
	}
	if insertedId == nil {
		return Order{}, errors.Join(errors2.ServerError, errors.New("empty inserted id"))
	}
	id, _ := (insertedId.InsertedID).(primitive.ObjectID)
	order.Id = id
	return order, nil
}

func (u OrderMongo) UpdateOrder(ctx context.Context, order Order) error {
	updateResult, err := u.DB.Collection("Order").UpdateByID(ctx, order.Id, bson.M{"$set": order}) // ignoring the parameter as we don't require it
	if err != nil {
		return err
	}

	if updateResult == nil {
		return errors.Join(errors2.ServerError, errors.New("no update result"))
	}
	if updateResult.MatchedCount != 1 {
		return errors.Join(errors2.ClientError, errors.New("no matching document"))
	}
	if updateResult.ModifiedCount != 1 {
		return errors.Join(errors2.ServerError, errors.New("update failed"))
	}
	return nil
}

func (u OrderMongo) GetOrder(ctx context.Context, id primitive.ObjectID) (Order, error) {
	var order Order
	err := u.DB.Collection("Order").FindOne(ctx, bson.M{"_id": id}).Decode(&order)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return order, errors.Join(errors2.ClientError, errors.New("no such entry"))
		}
		return order, errors.Join(errors2.ServerError, err)
	}
	return order, nil
}

func (u OrderMongo) SearchOrder(ctx context.Context, query SearchOrderQuery) ([]Order, int64, error) {
	filter := bson.M{}
	if !query.OrderId.IsZero() {
		filter["_id"] = query.OrderId
	}
	if !query.RestaurantId.IsZero() {
		filter["restaurantId"] = query.RestaurantId
	}
	if !query.DriverId.IsZero() {
		filter["driverId"] = query.DriverId
	}
	if !query.UserId.IsZero() {
		filter["userId"] = query.UserId
	}
	if query.Status != "" {
		filter["status"] = query.Status
	}

	totalCount, err := u.DB.Collection("Order").CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	options := options.Find()
	options.SetLimit(int64(query.Limit))
	options.SetSkip(int64(query.Skip))

	cursor, err := u.DB.Collection("Order").Find(ctx, filter, options)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var orders []Order
	if err := cursor.All(ctx, &orders); err != nil {
		return nil, 0, err
	}

	return orders, totalCount, nil
}
