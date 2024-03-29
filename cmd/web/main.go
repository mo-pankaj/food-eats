package main

import (
	"context"
	"flag"
	"food-eats/cmd/web/db"
	"food-eats/cmd/web/logger"
	middleware2 "food-eats/cmd/web/middelwares"
	"food-eats/cmd/web/model"
	"food-eats/cmd/web/routes"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"log/slog"
	"net/http"
	"os"
)

// CustomValidator to have more control on our validator
type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		// Optionally, you could return the error to give each route more control over the status code
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

func main() {
	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}

	// setting Long date, Long time, Long Microseconds, and Long file path for log
	opts := slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
	}
	jsonHandler := slog.NewJSONHandler(os.Stdout, &opts)
	ctxHandler := logger.ContextHandler{Handler: jsonHandler}
	l := slog.New(ctxHandler)
	slog.SetDefault(l)

	// todo run via flags
	uri := flag.String("mongo-uri", "mongodb://127.0.0.1:27017", "Mongodb uri")
	mongodb := flag.String("mongo db", "food-eats", "Mongodb database")
	redisUri := flag.String("redis-uri", "127.0.0.1:6380", "Redis uri")
	flag.Parse()

	mongoDatabase, err := db.GetMongoClient(context.TODO(), *uri, *mongodb)
	if err != nil {
		panic("unable to connect to mongo db")
	}

	// init redis
	db.InitRedisPool(*redisUri, 100, 200)

	sm := model.NewWebSocketManager(mongoDatabase)

	// adding middlewares
	e.Pre(middleware2.RequestIDMiddleware)
	e.Pre(middleware2.AddMetaData)
	e.Use(middleware.Recover())

	// using GZIP to compress the result
	e.Use(middleware.Gzip())

	initRoutes(mongoDatabase, sm, e)

	log.Println("Server starting....")
	log.Panic(e.Start(":8080"))
}

func initRoutes(mongoDatabase *mongo.Database, sm *model.SocketManager, e *echo.Echo) {
	initUserEndPoints(mongoDatabase, sm, e)
	initRestaurantEndPoints(mongoDatabase, sm, e)
	initRiderEndPoints(mongoDatabase, sm, e)
	initOrderEndPoints(mongoDatabase, sm, e)
	initRatingEndPoints(mongoDatabase, e)
	initWebSocketConnect(mongoDatabase, sm, e)
}

func initUserEndPoints(mongodb *mongo.Database, sm *model.SocketManager, e *echo.Echo) {
	userGroup := e.Group("/v1/user")
	userApplication := routes.UserApplication{MongoDb: mongodb}
	userGroup.POST("/create", userApplication.CreateUser)
	userGroup.PUT("/edit", userApplication.UpdateUser)
	userGroup.GET("/get", userApplication.GetUser)
	userGroup.DELETE("/delete", userApplication.DeleteUser)
}

func initRestaurantEndPoints(mongodb *mongo.Database, sm *model.SocketManager, e *echo.Echo) {
	restaurantGroup := e.Group("/v1/restaurant")
	restaurantApplication := routes.RestaurantApplication{
		MongoDb: mongodb,
		Cache:   db.GetRestaurantCache(),
	}
	restaurantGroup.POST("/create", restaurantApplication.CreateRestaurant)
	restaurantGroup.PUT("/edit", restaurantApplication.UpdateRestaurant)
	restaurantGroup.GET("/get", restaurantApplication.GetRestaurant)
	restaurantGroup.DELETE("/delete", restaurantApplication.DeleteRestaurant)

	restaurantGroup.POST("/search_restaurant", restaurantApplication.SearchRestaurant)
}

func initOrderEndPoints(mongodb *mongo.Database, sm *model.SocketManager, e *echo.Echo) {
	userGroup := e.Group("/v1/order")
	orderApplication := routes.OrderApplication{
		MongoDb: mongodb,
		SM:      sm,
	}
	userGroup.POST("/create", orderApplication.CreateOrder)
	userGroup.GET("/restaurant/get_pending_orders", orderApplication.GetRestaurantPendingOrder)
	userGroup.POST("/restaurant/accept_order", orderApplication.AcceptOrder)
	userGroup.POST("/search/get_orders", orderApplication.SearchOrder)
}

func initRiderEndPoints(mongodb *mongo.Database, sm *model.SocketManager, e *echo.Echo) {
	riderGroup := e.Group("/v1/rider")
	userApplication := routes.RiderApplication{MongoDb: mongodb}
	riderGroup.POST("/create", userApplication.CreateRider)
	riderGroup.PUT("/edit", userApplication.UpdateRider)
	riderGroup.GET("/get", userApplication.GetRider)
	riderGroup.DELETE("/delete", userApplication.DeleteRider)
}

func initWebSocketConnect(mongodb *mongo.Database, sm *model.SocketManager, e *echo.Echo) {
	wsGroup := e.Group("/v1/websocket")
	wsApplication := routes.NewWsApplication(mongodb, sm)
	wsGroup.GET("/rider", wsApplication.ConnectRiderWebSocket)
	wsGroup.GET("/user", wsApplication.ConnectUserWebSocket)
}

func initRatingEndPoints(mongodb *mongo.Database, e *echo.Echo) {
	ratingGroup := e.Group("/v1/rating")
	userApplication := routes.RatingApplication{MongoDb: mongodb}
	ratingGroup.POST("/create", userApplication.CreateNewRating)
	ratingGroup.GET("/get", userApplication.GetRating)
}
