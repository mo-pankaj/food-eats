package handlers_test

import (
	"context"
	"food-eats/cmd/web/db"
	"testing"

	"food-eats/cmd/web/handlers"
	"food-eats/cmd/web/model"
	"github.com/stretchr/testify/assert"
)

func TestSearchRestaurantHandler(t *testing.T) {
	ctx := context.TODO()
	mongoDatabase, err := db.GetMongoClient(ctx, "mongodb://127.0.0.1:27017", "food-eats-testing")
	assert.NoError(t, err)

	param := handlers.RestaurantParam{
		Repository: model.RestaurantMongoRepo(mongoDatabase),
	}

	// searching only necessary fields
	req := handlers.SearchRestaurantRequest{
		//Name:      "Sample Restaurant",
		MealType: "Veg",
		//Cuisine:   "Indian",
		//Rating:    4.5,
		Latitude:  "28.6358",
		Longitude: "77.91011",
		//Radius:    20.0,
		SortBy: "distance",
		Limit:  10,
		Offset: 0,
	}

	restaurants, err := req.SearchRestaurant(ctx, param)
	assert.NoError(t, err)

	// Ensure that all 7 restaurants are present in the database
	expectedCount := 4
	assert.Equal(t, expectedCount, len(restaurants.Items), "Expected %d restaurants, got %d", expectedCount, len(restaurants.Items))

	for i, restaurant := range restaurants.Items {
		t.Logf("Restaurant %d: %s", i+1, restaurant.Name)
	}
}
