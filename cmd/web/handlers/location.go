package handlers

import (
	"context"
	"food-eats/cmd/web/model"
	"math"
)

func updateDeliveryTime(param OrderParam, order model.Order) error {
	ctx := context.TODO()
	query := model.SearchOrderQuery{
		OrderId:      order.Id,
		Status:       "DELIVERED",
		RestaurantId: order.RestaurantId,
	}

	searchOrders, _, err := param.OrderRepo.SearchOrder(ctx, query)
	if err != nil {
		return err
	}

	// assuming order address is not changing
	lat, long := order.PickupLatitude, order.PickupLongitude
	totalDistance := float64(0)
	totalTime := float64(0)

	for _, searchOrder := range searchOrders {
		totalTime += searchOrder.DeliveryTime
		deliveryLat, deliveryLong := searchOrder.PickupLatitude, searchOrder.PickupLongitude
		totalDistance += distanceBetweenPoints(lat, long, deliveryLat, deliveryLong)
	}

	averageTime := totalDistance / (totalTime + 1)
	err = param.RestaurantRepo.UpdateAverageDeliveryTime(ctx, order.RestaurantId, averageTime)
	if err != nil {
		return err
	}

	return nil
}

func distanceBetweenPoints(lat1 float64, lng1 float64, lat2 float64, lng2 float64) float64 {
	radlat1 := math.Pi * lat1 / 180
	radlat2 := math.Pi * lat2 / 180

	theta := lng1 - lng2
	radtheta := math.Pi * theta / 180

	dist := math.Sin(radlat1)*math.Sin(radlat2) + math.Cos(radlat1)*math.Cos(radlat2)*math.Cos(radtheta)
	if dist > 1 {
		dist = 1
	}

	dist = math.Acos(dist)
	dist = dist * 180 / math.Pi
	dist = dist * 60 * 1.1515

	dist = dist * 1.609344

	return dist
}
