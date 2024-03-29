package routes

import (
	"food-eats/cmd/web/custom-errors"
	"food-eats/cmd/web/handlers"
	"food-eats/cmd/web/model"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
)

// RiderApplication contains the field dependencies for RiderApplication
type RiderApplication struct {
	MongoDb *mongo.Database
}

// CreateRider route for registering a Rider
func (ua *RiderApplication) CreateRider(c echo.Context) error {
	req := new(handlers.CreateRiderRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := c.Validate(req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	riderRepo := model.RiderMongoRepo(ua.MongoDb)
	repo := model.RiderRepository(riderRepo)
	restaurantParams := handlers.RiderParam{
		Repository: repo,
	}
	record, err := req.CreateRider(ctx, restaurantParams)

	if err != nil {
		return custom_errors.ParseError(ctx, err, req, c)
	}

	return c.JSON(http.StatusCreated, record)
}

// UpdateRider route for updating a Rider
func (ua *RiderApplication) UpdateRider(c echo.Context) error {
	req := new(handlers.UpdateRiderRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := c.Validate(req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	riderRepo := model.RiderMongoRepo(ua.MongoDb)
	repo := model.RiderRepository(riderRepo)
	restaurantParams := handlers.RiderParam{
		Repository: repo,
	}
	err := req.UpdateRider(ctx, restaurantParams)

	if err != nil {
		return custom_errors.ParseError(ctx, err, req, c)
	}

	return c.JSON(http.StatusOK, nil)
}

// GetRider getting a Rider
func (ua *RiderApplication) GetRider(c echo.Context) error {
	req := new(handlers.GetRiderRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := c.Validate(req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	riderRepo := model.RiderMongoRepo(ua.MongoDb)
	repo := model.RiderRepository(riderRepo)
	restaurantParams := handlers.RiderParam{
		Repository: repo,
	}
	rider, err := req.GetRider(ctx, restaurantParams)

	if err != nil {
		return custom_errors.ParseError(ctx, err, req, c)
	}

	return c.JSON(http.StatusOK, rider)
}

// DeleteRider marking a rider deleted
func (ua *RiderApplication) DeleteRider(c echo.Context) error {
	req := new(handlers.GetRiderRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := c.Validate(req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	riderRepo := model.RiderMongoRepo(ua.MongoDb)
	repo := model.RiderRepository(riderRepo)
	restaurantParams := handlers.RiderParam{
		Repository: repo,
	}
	rider, err := req.GetRider(ctx, restaurantParams)

	if err != nil {
		return custom_errors.ParseError(ctx, err, req, c)
	}

	return c.JSON(http.StatusOK, rider)
}
