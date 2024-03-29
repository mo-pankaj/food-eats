package routes

import (
	custom_errors "food-eats/cmd/web/custom-errors"
	"food-eats/cmd/web/handlers"
	"food-eats/cmd/web/model"
	"github.com/labstack/echo/v4"
	"github.com/patrickmn/go-cache"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
)

// RestaurantApplication contains the field dependencies for RestaurantApplication
type RestaurantApplication struct {
	MongoDb *mongo.Database
	Cache   *cache.Cache
}

// CreateRestaurant route for registering a restaurant
func (ua *RestaurantApplication) CreateRestaurant(c echo.Context) error {
	req := new(handlers.CreateRestaurantRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ctx := c.Request().Context()

	restaurantRepo := model.RestaurantMongoRepo(ua.MongoDb)
	var repo model.RestaurantRepository = restaurantRepo
	restaurantParams := handlers.RestaurantParam{
		Repository: repo,
	}
	record, err := req.CreateRestaurant(ctx, restaurantParams)

	if err != nil {
		return custom_errors.ParseError(ctx, err, req, c)
	}

	return c.JSON(http.StatusCreated, record)
}

// UpdateRestaurant route for updating a restaurant
func (ua *RestaurantApplication) UpdateRestaurant(c echo.Context) error {
	req := new(handlers.UpdateRestaurantRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ctx := c.Request().Context()

	restaurantRepo := model.RestaurantMongoRepo(ua.MongoDb)
	var repo model.RestaurantRepository = restaurantRepo
	restaurantParams := handlers.RestaurantParam{
		Repository: repo,
		Cache:      ua.Cache,
	}

	err := req.UpdateRestaurant(ctx, restaurantParams)

	if err != nil {
		return custom_errors.ParseError(ctx, err, req, c)
	}

	return c.JSON(http.StatusOK, nil)
}

// GetRestaurant getting a restaurant
func (ua *RestaurantApplication) GetRestaurant(c echo.Context) error {
	req := new(handlers.GetRestaurantRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ctx := c.Request().Context()

	restaurantRepo := model.RestaurantMongoRepo(ua.MongoDb)
	repo := model.RestaurantRepository(restaurantRepo)
	restaurantParams := handlers.RestaurantParam{
		Repository: repo,
		Cache:      ua.Cache,
	}

	restaurant, err := req.GetRestaurant(ctx, restaurantParams)

	if err != nil {
		return custom_errors.ParseError(ctx, err, req, c)
	}

	return c.JSON(http.StatusOK, restaurant)
}

// DeleteRestaurant marking a restaurant deleted
func (ua *RestaurantApplication) DeleteRestaurant(c echo.Context) error {
	req := new(handlers.GetRestaurantRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ctx := c.Request().Context()

	restaurantRepo := model.RestaurantMongoRepo(ua.MongoDb)
	var repo model.RestaurantRepository = restaurantRepo
	restaurantParams := handlers.RestaurantParam{
		Repository: repo,
		Cache:      ua.Cache,
	}

	restaurant, err := req.DeleteRestaurant(ctx, restaurantParams)

	if err != nil {
		return custom_errors.ParseError(ctx, err, req, c)
	}

	return c.JSON(http.StatusOK, restaurant)
}

// SearchRestaurant search restaurant
// takes different fields, and searches base on them
func (ua *RestaurantApplication) SearchRestaurant(c echo.Context) error {

	req := new(handlers.SearchRestaurantRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ctx := c.Request().Context()

	restaurantRepo := model.RestaurantMongoRepo(ua.MongoDb)
	var repo model.RestaurantRepository = restaurantRepo
	restaurantParams := handlers.RestaurantParam{
		Repository: repo,
		Cache:      ua.Cache,
	}
	response, err := req.SearchRestaurant(ctx, restaurantParams)

	if err != nil {
		return custom_errors.ParseError(ctx, err, req, c)
	}
	return c.JSON(http.StatusOK, response)
}
