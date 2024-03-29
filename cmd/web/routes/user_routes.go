package routes

import (
	"food-eats/cmd/web/custom-errors"
	"food-eats/cmd/web/handlers"
	"food-eats/cmd/web/model"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
)

// UserApplication contains the field dependencies for UserApplication
type UserApplication struct {
	MongoDb *mongo.Database
	SM      *model.SocketManager
}

// CreateUser route for registering a user
func (ua *UserApplication) CreateUser(c echo.Context) error {
	req := new(handlers.CreateUserRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ctx := c.Request().Context()

	userRepo := model.UserMongoRepo(ua.MongoDb)
	repo := model.UserRepository(userRepo)
	param := handlers.UserParam{
		Repository: repo,
	}
	record, err := req.CreateUser(ctx, param)

	if err != nil {
		return custom_errors.ParseError(ctx, err, req, c)
	}

	return c.JSON(http.StatusCreated, record)
}

// UpdateUser route for updating a user
func (ua *UserApplication) UpdateUser(c echo.Context) error {
	req := new(handlers.UpdateUserRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ctx := c.Request().Context()

	userRepo := model.UserMongoRepo(ua.MongoDb)
	repo := model.UserRepository(userRepo)
	param := handlers.UserParam{
		Repository: repo,
	}
	err := req.UpdateUser(ctx, param)

	if err != nil {
		return custom_errors.ParseError(ctx, err, req, c)
	}

	return c.JSON(http.StatusOK, nil)
}

// GetUser getting a user
func (ua *UserApplication) GetUser(c echo.Context) error {
	req := new(handlers.GetUserRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ctx := c.Request().Context()

	userRepo := model.UserMongoRepo(ua.MongoDb)
	repo := model.UserRepository(userRepo)
	param := handlers.UserParam{
		Repository: repo,
	}
	user, err := req.GetUser(ctx, param)

	if err != nil {
		return custom_errors.ParseError(ctx, err, req, c)
	}

	return c.JSON(http.StatusOK, user)
}

// DeleteUser marking a user deleted
func (ua *UserApplication) DeleteUser(c echo.Context) error {
	req := new(handlers.GetUserRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ctx := c.Request().Context()

	userRepo := model.UserMongoRepo(ua.MongoDb)
	repo := model.UserRepository(userRepo)
	param := handlers.UserParam{
		Repository: repo,
	}
	err := req.DeleteUser(ctx, param)

	if err != nil {
		return custom_errors.ParseError(ctx, err, req, c)
	}

	return c.JSON(http.StatusOK, nil)
}
