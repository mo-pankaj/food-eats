package handlers

import (
	"context"
	"food-eats/cmd/web/model"
	"github.com/patrickmn/go-cache"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strconv"
	"time"
)

// RestaurantParam request param contains all dependencies
type RestaurantParam struct {
	Repository model.RestaurantRepository
	Cache      *cache.Cache
}

// CreateRestaurantRequest a type for creating a restaurant request
type CreateRestaurantRequest struct {
	Name    string `json:"name" validate:"required"`
	EmailId string `json:"email_id" validate:"required,email"`

	// applied validation on indian numbers in format of +91999999999, allowing both 9 and 10 digit numbers
	PhoneNumber string `json:"phone_number" validate:"required,e164,min=12,max=13,startswith=+91"`

	Latitude  string     `json:"latitude" validate:"required,latitude"`
	Longitude string     `json:"longitude" validate:"required,longitude"`
	Address   string     `json:"address" validate:"required"`
	Website   string     `json:"website"`
	Cuisines  string     `json:"cuisines"`
	MealType  string     `json:"meal_type"`
	Menu      model.Menu `json:"menu"`
}

// UpdateRestaurantRequest a type for updaing a restaurant request
type UpdateRestaurantRequest struct {
	Id      primitive.ObjectID `json:"id" validate:"required"` // required a mongodb objectId
	Name    string             `json:"name" validate:"required"`
	EmailId string             `json:"email_id" validate:"required,email"`

	// applied validation on indian numbers in format of +91999999999, allowing both 9 and 10 digit numbers
	PhoneNumber string `json:"phone_number" validate:"required,e164,min=12,max=13,startswith=+91"`

	Latitude  string     `json:"latitude" validate:"required,latitude"`
	Longitude string     `json:"longitude" validate:"required,longitude"`
	Address   string     `json:"address" validate:"required"`
	Website   string     `json:"website"`
	Cuisines  string     `json:"cuisines"`
	MealType  string     `json:"meal_type"`
	Menu      model.Menu `json:"menu"`
}

// GetRestaurantRequest a type for updaing a restaurant request
type GetRestaurantRequest struct {
	Id primitive.ObjectID `query:"id" validate:"required"` // required a mongodb objectId
}

type SearchRestaurantRequest struct {
	Name      string  `json:"name"`
	MealType  string  `json:"meal_type"`
	Cuisine   string  `json:"cuisine"`
	Rating    float64 `json:"rating" validate:"min=0,max=5"` // the rating should be in between 0 and 5
	Latitude  string  `json:"latitude" validate:"latitude"`
	Longitude string  `json:"longitude" validate:"longitude"`
	Radius    float64 `json:"radius"`
	SortBy    string  `json:"sort_by"`
	Limit     int     `json:"limit"`
	Offset    int     `json:"offset" validate:"min=0"`
}

// CreateRestaurant register new restaurant
func (request *CreateRestaurantRequest) CreateRestaurant(ctx context.Context, param RestaurantParam) (model.Restaurant, error) {
	currTime := time.Now()
	// Create RestaurantId
	restaurant := model.Restaurant{
		Name:        request.Name,
		EmailId:     request.EmailId,
		PhoneNumber: request.PhoneNumber,
		Location:    model.NewLocationFromLongLatStr(request.Longitude, request.Latitude),
		Address:     request.Address,
		Status:      "ACTIVE",
		Website:     request.Website,
		Cuisines:    request.Cuisines,
		MealType:    request.Cuisines,
		Menu:        request.Menu,
		CreatedAt:   currTime,
		UpdatedAt:   currTime,
	}

	createdRecord, err := param.Repository.CreateRestaurant(ctx, restaurant)
	if err != nil {
		return model.Restaurant{}, err
	}

	return createdRecord, nil
}

// UpdateRestaurant update a new restaurant
func (request *UpdateRestaurantRequest) UpdateRestaurant(ctx context.Context, param RestaurantParam) error {
	currTime := time.Now()

	restaurant, err := param.Repository.GetRestaurant(ctx, request.Id)
	if err != nil {
		return err
	}

	// Update RestaurantId
	restaurant.Name = request.Name
	restaurant.EmailId = request.EmailId
	restaurant.PhoneNumber = request.PhoneNumber
	restaurant.Location = model.NewLocationFromLongLatStr(request.Longitude, request.Latitude)
	restaurant.Address = request.Address
	restaurant.Website = request.Website
	restaurant.Cuisines = request.Cuisines
	restaurant.MealType = request.Cuisines
	restaurant.Menu = request.Menu

	restaurant.UpdatedAt = currTime

	err = param.Repository.UpdateRestaurant(ctx, restaurant)
	if err != nil {
		return err
	}

	// removing item from cache
	// todo can use mutex here to handle concurrency
	// todo move to constants
	param.Cache.Delete("restaurant_cache:" + restaurant.Id.Hex())

	return nil
}

// GetRestaurant getting a restaurant
func (request *GetRestaurantRequest) GetRestaurant(ctx context.Context, param RestaurantParam) (model.Restaurant, error) {
	if value, found := param.Cache.Get("restaurant_cache:" + request.Id.Hex()); found {
		return value.(model.Restaurant), nil
	}
	restaurant, err := param.Repository.GetRestaurant(ctx, request.Id)
	if err != nil {
		return model.Restaurant{}, err
	}

	param.Cache.Set("restaurant_cache:"+request.Id.Hex(), restaurant, 10*time.Minute)

	return restaurant, nil
}

// DeleteRestaurant deleting a restaurant
func (request *GetRestaurantRequest) DeleteRestaurant(ctx context.Context, param RestaurantParam) (model.Restaurant, error) {
	restaurant, err := param.Repository.GetRestaurant(ctx, request.Id)
	if err != nil {
		return model.Restaurant{}, err
	}
	restaurant.Status = "DELETED"
	err = param.Repository.UpdateRestaurant(ctx, restaurant)
	if err != nil {
		return model.Restaurant{}, err
	}

	param.Cache.Delete("restaurant_cache:" + restaurant.Id.Hex())

	return restaurant, nil
}

// SearchRestaurant a new restaurant search
func (request *SearchRestaurantRequest) SearchRestaurant(ctx context.Context, param RestaurantParam) (*model.PageResponse[model.RestaurantSearchResponse], error) {
	searchBuilder := model.SearchRestaurantQuery{}
	if request.Name != "" {
		searchBuilder.Name = request.Name
	}
	if request.MealType != "" {
		searchBuilder.MealType = request.MealType
	}
	if request.Cuisine != "" {
		searchBuilder.CuisineType = request.Cuisine
	}

	if request.Rating > 0 {
		searchBuilder.MinRating = request.Rating
	}
	if request.Latitude != "" {
		searchBuilder.Latitude, _ = strconv.ParseFloat(request.Latitude, 64)
	}
	if request.Longitude != "" {
		searchBuilder.Longitude, _ = strconv.ParseFloat(request.Longitude, 64)
	}
	if request.Radius > 0 {
		searchBuilder.Radius = request.Radius
	}
	// using a min and max result
	if request.Limit > 10 && request.Limit < 20 {
		searchBuilder.Limit = request.Limit
	} else {
		searchBuilder.Limit = 10
	}
	if request.Offset > 1 {
		searchBuilder.Skip = (request.Offset - 1) * request.Limit
	} else {
		searchBuilder.Skip = 0
	}
	if request.SortBy != "" {
		searchBuilder.SortBy = request.SortBy
	}

	restaurants, totalCount, err := param.Repository.SearchRestaurant(ctx, searchBuilder)
	if err != nil {
		return nil, err
	}

	pageResponse := model.NewPageResponse(restaurants, totalCount, request.Offset, request.Limit)

	return pageResponse, nil
}
