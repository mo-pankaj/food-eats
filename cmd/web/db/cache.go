package db

import (
	"github.com/patrickmn/go-cache"
	"time"
)

// GetRestaurantCache get restaurant cache
func GetRestaurantCache() *cache.Cache {
	restaurant := cache.New(10*time.Minute, 20*time.Minute)
	return restaurant
}
