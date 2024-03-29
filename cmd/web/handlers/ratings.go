package handlers

import (
	"context"
	"errors"
	"fmt"
	"food-eats/cmd/web/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type RatingRequest struct {
	RatingGiver    primitive.ObjectID `json:"rating_giver"`
	RatingReceiver primitive.ObjectID `json:"rating_receiver"`
	OrderId        primitive.ObjectID `json:"order_id"`

	RatingGiverType    string `json:"rating_giver_type"`
	RatingReceiverType string `json:"rating_receiver_type"`
	Rating             int    `json:"rating" validate:"min=0,max=5"`
	Comment            string
}

type GetRatingRequest struct {
	Id primitive.ObjectID `query:"id"`
}

// RatingParam request param contains all dependencies
type RatingParam struct {
	RestaurantRepo model.RestaurantRepository
	RiderRepo      model.RiderRepository
	OrderRepo      model.OrderRepository
	UserRepo       model.UserRepository
	RatingRepo     model.RatingRepository
}

// CreateRating creates rating
func (request *RatingRequest) CreateRating(ctx context.Context, param RatingParam) (model.Rating, error) {
	var reciever any
	var err error
	switch request.RatingGiverType {
	case "rider":
		_, err := param.RiderRepo.GetRider(ctx, request.RatingGiver)
		if err != nil {
			return model.Rating{}, err
		}
	case "user":
		_, err := param.UserRepo.GetUser(ctx, request.RatingGiver)
		if err != nil {
			return model.Rating{}, err
		}
	case "restaurant":
		_, err := param.RestaurantRepo.GetRestaurant(ctx, request.RatingReceiver)
		if err != nil {
			return model.Rating{}, err
		}
	default:
		return model.Rating{}, errors.New("invalid rating giver type")
	}

	switch request.RatingReceiverType {
	case "rider":
		reciever, err = param.RiderRepo.GetRider(ctx, request.RatingReceiver)
		if err != nil {
			return model.Rating{}, err
		}
	case "user":
		reciever, err = param.UserRepo.GetUser(ctx, request.RatingReceiver)
		if err != nil {
			return model.Rating{}, err
		}
	case "restaurant":
		reciever, err = param.RestaurantRepo.GetRestaurant(ctx, request.RatingReceiver)
		if err != nil {
			return model.Rating{}, err
		}
	default:
		return model.Rating{}, errors.New(fmt.Sprintf("invalid rating receiver type: %v", reciever))
	}

	rating := model.Rating{
		RatingGiver:        request.RatingGiver,
		RatingReceiver:     request.RatingReceiver,
		OrderID:            request.OrderId,
		RatingGiverType:    request.RatingGiverType,
		RatingReceiverType: request.RatingReceiverType,
		Rating:             request.Rating,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
		Comment:            request.Comment,
	}

	createdRating, err := param.RatingRepo.CreateRating(ctx, rating)
	if err != nil {
		return model.Rating{}, err
	}

	switch request.RatingReceiverType {

	case "rider":
		rider, ok := reciever.(model.Rider)
		if !ok {
			break
		}
		go request.UpdateAverageRatingsForRider(context.TODO(), param, rider.Id)

	case "user":
		user, ok := reciever.(model.User)
		if !ok {
			break
		}
		go request.UpdateAverageRatingsForUser(context.TODO(), param, user.Id)
	case "restaurant":
		restaurant, ok := reciever.(model.Restaurant)
		if !ok {
			break
		}
		go request.UpdateAverageRatingsForRestaurant(context.TODO(), param, restaurant.Id)
	default:
		return model.Rating{}, errors.New(fmt.Sprintf("invalid rating receiver type: %v", reciever))
	}

	return createdRating, nil
}

func (request *GetRatingRequest) GetRating(ctx context.Context, param RatingParam) (model.Rating, error) {
	rating, err := param.RatingRepo.GetRating(ctx, request.Id)
	if err != nil {
		return model.Rating{}, err
	}

	return rating, nil
}

// UpdateAverageRatingsForUser updates average ratings for a user
func (request *RatingRequest) UpdateAverageRatingsForUser(ctx context.Context, param RatingParam, userID primitive.ObjectID) error {
	ratings, err := param.RatingRepo.SearchRating(ctx, model.SearchRatingBuilder{RatingGiver: request.RatingGiver})
	if err != nil {
		return err
	}

	averageRating := request.calculateAverageRating(ratings)

	err = param.UserRepo.UpdateAverageRating(ctx, userID, averageRating)
	if err != nil {
		return err
	}

	return nil
}

// UpdateAverageRatingsForRider updates average ratings for a rider
func (request *RatingRequest) UpdateAverageRatingsForRider(ctx context.Context, param RatingParam, riderID primitive.ObjectID) error {
	ratings, err := param.RatingRepo.SearchRating(ctx, model.SearchRatingBuilder{RatingGiver: request.RatingGiver})
	if err != nil {
		return err
	}

	averageRating := request.calculateAverageRating(ratings)

	err = param.RiderRepo.UpdateAverageRating(ctx, riderID, averageRating)
	if err != nil {
		return err
	}

	return nil
}

// UpdateAverageRatingsForRestaurant updates average ratings for a restaurant
func (request *RatingRequest) UpdateAverageRatingsForRestaurant(ctx context.Context, param RatingParam, restaurantID primitive.ObjectID) error {
	ratings, err := param.RatingRepo.SearchRating(ctx, model.SearchRatingBuilder{RatingGiver: request.RatingGiver})
	if err != nil {
		return err
	}

	averageRating := request.calculateAverageRating(ratings)

	err = param.RestaurantRepo.UpdateAverageRating(ctx, restaurantID, averageRating)
	if err != nil {
		return err
	}

	return nil
}

// calculateAverageRating calculates the average rating from a slice of ratings
func (request *RatingRequest) calculateAverageRating(ratings []model.Rating) float64 {
	if len(ratings) == 0 {
		return 0
	}

	totalRating := 0
	for _, r := range ratings {
		totalRating += r.Rating
	}

	averageRating := float64(totalRating) / float64(len(ratings))
	return averageRating
}
