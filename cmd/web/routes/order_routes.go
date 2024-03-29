package routes

import (
	"food-eats/cmd/web/custom-errors"
	"food-eats/cmd/web/db"
	"food-eats/cmd/web/handlers"
	"food-eats/cmd/web/model"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
)

// OrderApplication contains the field dependencies for OrderApplication
type OrderApplication struct {
	MongoDb *mongo.Database
	SM      *model.SocketManager
}

// CreateOrder route for registering a order
func (ua *OrderApplication) CreateOrder(c echo.Context) error {
	req := new(handlers.CreateOrderRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := c.Validate(req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	orderRepo := model.OrderMongoRepo(ua.MongoDb)
	redisConn, err := db.RedisConnFromPool()
	if err != nil {
		return err
	}
	defer redisConn.Close()

	restaurantRepo := model.RestaurantRepository(model.RestaurantMongoRepo(ua.MongoDb))
	riderRepo := model.RiderRepository(model.RiderMongoRepo(ua.MongoDb))
	userRepo := model.UserRepository(model.UserMongoRepo(ua.MongoDb))

	repo := handlers.OrderParam{
		OrderRepo:      model.OrderRepository(orderRepo),
		RestaurantRepo: restaurantRepo,
		RiderRepo:      riderRepo,
		UserRepo:       userRepo,
		RedisConn:      redisConn,
	}

	record, err := req.CreateOrder(ctx, repo)

	if err != nil {
		if err != nil {
			return custom_errors.ParseError(ctx, err, req, c)
		}
	}

	return c.JSON(http.StatusCreated, record)
}

// GetOrder getting a order
func (ua *OrderApplication) GetOrder(c echo.Context) error {
	req := new(handlers.GetOrderRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := c.Validate(req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	orderRepo := model.OrderMongoRepo(ua.MongoDb)
	restaurantRepo := model.RestaurantRepository(model.RestaurantMongoRepo(ua.MongoDb))
	riderRepo := model.RiderRepository(model.RiderMongoRepo(ua.MongoDb))
	userRepo := model.UserRepository(model.UserMongoRepo(ua.MongoDb))

	repo := handlers.OrderParam{
		OrderRepo:      model.OrderRepository(orderRepo),
		RestaurantRepo: restaurantRepo,
		RiderRepo:      riderRepo,
		UserRepo:       userRepo,
	}
	order, err := req.GetOrder(ctx, repo)

	if err != nil {
		if err != nil {
			return custom_errors.ParseError(ctx, err, req, c)
		}
	}

	return c.JSON(http.StatusOK, order)
}

// GetRestaurantPendingOrder getting a order
func (ua *OrderApplication) GetRestaurantPendingOrder(c echo.Context) error {
	req := new(handlers.GetPendingRestaurantOrder)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := c.Validate(req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	orderRepo := model.OrderMongoRepo(ua.MongoDb)
	restaurantRepo := model.RestaurantRepository(model.RestaurantMongoRepo(ua.MongoDb))
	riderRepo := model.RiderRepository(model.RiderMongoRepo(ua.MongoDb))
	userRepo := model.UserRepository(model.UserMongoRepo(ua.MongoDb))

	repo := handlers.OrderParam{
		OrderRepo:      model.OrderRepository(orderRepo),
		RestaurantRepo: restaurantRepo,
		RiderRepo:      riderRepo,
		UserRepo:       userRepo,
	}
	order, err := req.GetPendingRestaurantOrder(ctx, repo)

	if err != nil {
		if err != nil {
			return custom_errors.ParseError(ctx, err, req, c)
		}
	}

	return c.JSON(http.StatusOK, order)
}

// AcceptOrder accepts an order
func (ua *OrderApplication) AcceptOrder(c echo.Context) error {

	req := new(handlers.AcceptPendingRestaurantOrder)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := c.Validate(req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	redisConn, err := db.RedisConnFromPool()
	if err != nil {
		return err
	}
	defer db.Close(redisConn)

	orderRepo := model.OrderRepository(model.OrderMongoRepo(ua.MongoDb))
	restaurantRepo := model.RestaurantRepository(model.RestaurantMongoRepo(ua.MongoDb))
	riderRepo := model.RiderRepository(model.RiderMongoRepo(ua.MongoDb))
	userRepo := model.UserRepository(model.UserMongoRepo(ua.MongoDb))

	repo := handlers.OrderParam{
		OrderRepo:      orderRepo,
		RestaurantRepo: restaurantRepo,
		RiderRepo:      riderRepo,
		UserRepo:       userRepo,
		SM:             ua.SM,
		RedisConn:      redisConn,
	}

	err = req.AcceptOrder(ctx, repo)
	if err != nil {
		if err != nil {
			return custom_errors.ParseError(ctx, err, req, c)
		}
	}

	return c.JSON(http.StatusOK, nil)
}

func (ua *OrderApplication) SearchOrder(c echo.Context) error {
	req := new(handlers.SearchOrderRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := c.Validate(req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	redisConn, err := db.RedisConnFromPool()
	if err != nil {
		return err
	}
	defer db.Close(redisConn)
	if req.Type != "user" && req.Type != "restaurant" && req.Type != "rider" {
		return c.JSON(http.StatusBadRequest, "unsupported type")
	}

	orderRepo := model.OrderMongoRepo(ua.MongoDb)
	restaurantRepo := model.RestaurantRepository(model.RestaurantMongoRepo(ua.MongoDb))
	riderRepo := model.RiderRepository(model.RiderMongoRepo(ua.MongoDb))
	userRepo := model.UserRepository(model.UserMongoRepo(ua.MongoDb))

	repo := handlers.OrderParam{
		OrderRepo:      model.OrderRepository(orderRepo),
		RestaurantRepo: restaurantRepo,
		RiderRepo:      riderRepo,
		UserRepo:       userRepo,
	}

	response, err := req.SearchOrder(ctx, repo)
	if err != nil {
		if err != nil {
			return custom_errors.ParseError(ctx, err, req, c)
		}
	}

	return c.JSON(http.StatusOK, response)
}
