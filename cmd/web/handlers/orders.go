package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"food-eats/cmd/web/custom-errors"
	"food-eats/cmd/web/model"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// CreateOrderRequest a type for creating a order request
type CreateOrderRequest struct {
	UserId       primitive.ObjectID `json:"user_id"`
	RestaurantId primitive.ObjectID `json:"restaurant_id"`
	Items        []model.Item       `json:"items"`
}

// UpdateOrderRequest a type for updaing a order request
type UpdateOrderRequest struct {
	Id      primitive.ObjectID `json:"id" validate:"required"` // required a mongodb objectId
	Name    string             `json:"name" validate:"required"`
	EmailId string             `json:"email_id" validate:"required,email"`

	// applied validation on indian numbers in the format of +91999999999, allowing both 9 and 10-digit numbers
	PhoneNumber string `json:"phone_number" validate:"required,e164,min=12,max=13,startswith=+91"`

	Latitude  string `json:"latitude" validate:"required,latitude"`
	Longitude string `json:"longitude" validate:"required,longitude"`
	Address   string `json:"address" validate:"required"`
}

// GetOrderRequest a type for updaing a order request
type GetOrderRequest struct {
	Id primitive.ObjectID `query:"id" validate:"required"` // required a mongodb objectId
}

// GetPendingRestaurantOrder get pending order for the restaurant
type GetPendingRestaurantOrder struct {
	Id      primitive.ObjectID `query:"id" validate:"required"` // required a mongodb objectId
	Limit   int                `query:"limit" validate:"min=10,max=20"`
	PageNum int                `query:"page_num" validate:"min=1"`
}

// AcceptPendingRestaurantOrder accept pending order for the restaurant
type AcceptPendingRestaurantOrder struct {
	Id           primitive.ObjectID `json:"id" validate:"required"`
	RestaurantId primitive.ObjectID `json:"restaurant_id" validate:"required"`
}

type SearchOrderRequest struct {
	Id      primitive.ObjectID `json:"id" validate:"required"`
	Type    string             `json:"type" validate:"required"`
	Limit   int                `json:"limit" `
	PageNum int                `json:"page_num"`
	Status  string             `json:"status"`
}

func (request SearchOrderRequest) SearchOrder(ctx context.Context, repo OrderParam) (*model.PageResponse[model.Order], error) {
	query := model.SearchOrderQuery{}
	if request.Type == "user" {
		query.UserId = request.Id
	} else if request.Type == "restaurant" {
		query.RestaurantId = request.Id
	} else {
		query.DriverId = request.Id
	}
	skip := (request.PageNum - 1) * request.Limit

	query.Status = request.Status
	query.Limit = request.Limit
	query.Skip = skip

	orders, total, err := repo.OrderRepo.SearchOrder(ctx, query)
	if err != nil {
		return nil, err
	}

	pageResponse := model.NewPageResponse(orders, total, request.PageNum, request.Limit)
	return pageResponse, nil
}

// GetNewDeliveryOrder get pending order for the restaurant
type GetNewDeliveryOrder struct {
	Id      primitive.ObjectID `query:"id" validate:"required"` // required a mongodb objectId
	Limit   int64              `query:"limit" validate:"min=10,max=20"`
	PageNum int64              `query:"page_num" validate:"min=1"`
}

type NewOrderBroadCast struct {
	Message string             `json:"message"`
	OrderId primitive.ObjectID `json:"order_id"`
}

// OrderParam order param
type OrderParam struct {
	OrderRepo      model.OrderRepository
	UserRepo       model.UserRepository
	RestaurantRepo model.RestaurantRepository
	RiderRepo      model.RiderRepository
	SM             *model.SocketManager

	RedisConn *redis.Conn
}

// CreateOrder register new order
func (request *CreateOrderRequest) CreateOrder(ctx context.Context, param OrderParam) (model.Order, error) {
	currTime := time.Now()

	user, err := param.UserRepo.GetUser(ctx, request.UserId)
	if err != nil {
		return model.Order{}, err
	}

	restaurant, err := param.RestaurantRepo.GetRestaurant(ctx, request.RestaurantId)
	if err != nil {
		return model.Order{}, err
	}

	if restaurant.Status != "ACTIVE" {
		return model.Order{}, errors.Join(custom_errors.ClientError, errors.New("restaurant not serving"))
	}

	finalPrice := float32(0)
	for _, item := range request.Items {
		finalPrice += item.Price
	}

	// Create Order
	order := model.Order{
		UserId:            user.Id,
		RestaurantId:      restaurant.Id,
		CreatedAt:         currTime,
		UpdatedAt:         currTime,
		Items:             request.Items,
		FinalPrice:        finalPrice,
		Status:            "CREATED",
		DeliveryLatitude:  user.Location.GetLatitude(),
		DeliveryLongitude: user.Location.GetLongitude(),
		DeliveryAddress:   user.Address,

		PickupLatitude:  user.Location.GetLatitude(),
		PickupLongitude: user.Location.GetLongitude(),
		PickupAddress:   restaurant.Address,
	}

	createdRecord, err := param.OrderRepo.CreateOrder(ctx, order)
	if err != nil {
		return model.Order{}, err
	}

	return createdRecord, nil
}

// UpdateOrder update a new order
func (request *UpdateOrderRequest) UpdateOrder(ctx context.Context, param OrderParam) error {
	// todo complete this
	//currTime := time.Now()

	//order, err := param.OrderRepo.GetOrder(ctx, request.Id)
	//if err != nil {
	//	return err
	//}
	//
	//// Update Order
	//order.Name = request.Name
	//order.EmailId = request.EmailId
	//order.PhoneNumber = request.PhoneNumber
	//order.Latitude = request.Latitude
	//order.Longitude = request.Longitude
	//order.DeliveryAddress = request.DeliveryAddress
	//order.UpdatedAt = currTime
	//
	//err = param.OrderRepo.UpdateOrder(ctx, order)
	//if err != nil {
	//	return err
	//}

	return nil
}

// GetOrder getting a order
func (request *GetOrderRequest) GetOrder(ctx context.Context, param OrderParam) (model.Order, error) {
	order, err := param.OrderRepo.GetOrder(ctx, request.Id)
	if err != nil {
		return model.Order{}, err
	}

	return order, nil
}

func (request *GetPendingRestaurantOrder) GetPendingRestaurantOrder(ctx context.Context, param OrderParam) (*model.PageResponse[model.Order], error) {

	// Calculate skip value based on page number and limit
	skip := (request.PageNum - 1) * request.Limit

	// Prepare the search query
	query := model.SearchOrderQuery{
		RestaurantId: request.Id,
		Status:       "CREATED",
		Limit:        request.Limit,
		Skip:         skip,
	}

	// Call the repository method to search for orders
	orders, total, err := param.OrderRepo.SearchOrder(ctx, query)
	if err != nil {
		return nil, err
	}

	// Create a page response object
	pageResponse := model.NewPageResponse(orders, total, request.PageNum, request.Limit)

	return pageResponse, nil
}

// AcceptOrder accepts an order by updating its status to "ACCEPTED"
// todo handle failure cases
func (oa *AcceptPendingRestaurantOrder) AcceptOrder(ctx context.Context, param OrderParam) error {
	// Retrieve the order from the repository
	order, err := param.OrderRepo.GetOrder(ctx, oa.Id)
	if err != nil {
		return err
	}

	restaurant, err := param.RestaurantRepo.GetRestaurant(ctx, oa.RestaurantId)
	if err != nil {
		return err
	}

	// redis connection
	// use INCR

	// Check if the order is already accepted
	if order.Status == "ACCEPTED" {
		return errors.New("order is already accepted")
	}

	// Update the order status to "ACCEPTED"
	order.Status = "ACCEPTED"
	order.AcceptedAt = time.Now()
	if err := param.OrderRepo.UpdateOrder(ctx, order); err != nil {
		return err
	}

	searchRider := model.SearchRiderQuery{
		Latitude:  restaurant.Location.GetLatitude(),
		Longitude: restaurant.Location.GetLongitude(),
		Limit:     10,
	}

	riders, err := param.RiderRepo.SearchRider(ctx, model.SearchRiderQuery{
		Latitude:  searchRider.Latitude,
		Longitude: searchRider.Longitude,
		Limit:     searchRider.Limit,
	})

	// ws riders
	var ridersSelected []primitive.ObjectID
	for _, rider := range riders {
		// broadcast to SM
		rId := rider.Id
		ridersSelected = append(ridersSelected, rId)
	}

	newOrder := NewOrderBroadCast{
		Message: "New order for pickup",
		OrderId: order.Id,
	}
	msg, err := json.Marshal(&newOrder)
	if err != nil {
		return err
	}
	param.SM.BroadcastToRiders(string(msg), ridersSelected)

	for _, id := range ridersSelected {
		_, err := param.RedisConn.SAdd(ctx, "rider_broadcasted:"+order.Id.Hex(), id.Hex()).Result()
		if err != nil {
			return err
		}
	}
	param.RedisConn.Expire(ctx, "rider_broadcasted:"+order.Id.Hex(), 30*time.Minute)

	if err != nil {
		return err
	}

	return nil
}
