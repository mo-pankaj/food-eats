package routes

import (
	custom_errors "food-eats/cmd/web/custom-errors"
	"food-eats/cmd/web/handlers"
	"food-eats/cmd/web/model"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
)

// RatingApplication contains the field dependencies for RatingApplication
type RatingApplication struct {
	MongoDb *mongo.Database
}

// CreateNewRating route for registering a rating
func (ua *RatingApplication) CreateNewRating(c echo.Context) error {
	req := new(handlers.RatingRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ctx := c.Request().Context()

	restaurantRepo := model.RestaurantRepository(model.RestaurantMongoRepo(ua.MongoDb))
	riderRepo := model.RiderRepository(model.RiderMongoRepo(ua.MongoDb))
	userRepo := model.UserRepository(model.UserMongoRepo(ua.MongoDb))
	orderRepo := model.OrderRepository(model.OrderMongoRepo(ua.MongoDb))
	ratingRepo := model.RatingRepository(model.RatingMongoRepo(ua.MongoDb))

	restaurantParams := handlers.RatingParam{
		RestaurantRepo: restaurantRepo,
		RiderRepo:      riderRepo,
		UserRepo:       userRepo,
		OrderRepo:      orderRepo,
		RatingRepo:     ratingRepo,
	}
	record, err := req.CreateRating(ctx, restaurantParams)

	if err != nil {
		return custom_errors.ParseError(ctx, err, req, c)
	}

	return c.JSON(http.StatusCreated, record)
}

// GetRating route for g3tting a rating
func (ua *RatingApplication) GetRating(c echo.Context) error {
	req := new(handlers.GetRatingRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ctx := c.Request().Context()

	ratingRepo := model.RatingRepository(model.RatingMongoRepo(ua.MongoDb))

	restaurantParams := handlers.RatingParam{
		RatingRepo: ratingRepo,
	}
	record, err := req.GetRating(ctx, restaurantParams)

	if err != nil {
		return custom_errors.ParseError(ctx, err, req, c)
	}

	return c.JSON(http.StatusCreated, record)
}
