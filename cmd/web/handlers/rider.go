package handlers

import (
	"context"
	"food-eats/cmd/web/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// RiderParam param for repository
type RiderParam struct {
	Repository model.RiderRepository
}

// RiderSearchRequest rider search request
type RiderSearchRequest struct {
	Latitude  float64
	Longitude float64
	// if order not accepted in time, increase search
	Limit int
}

// CreateRiderRequest a type for creating a rider request
type CreateRiderRequest struct {
	Name    string `json:"name" validate:"required"`
	EmailId string `json:"email_id" validate:"required,email"`

	// applied validation on indian numbers in the format of +91999999999, allowing both 9 and 10-digit numbers
	PhoneNumber string `json:"phone_number" validate:"required,e164,min=12,max=13,startswith=+91"`

	Latitude  string `json:"latitude" validate:"required,latitude"`
	Longitude string `json:"longitude" validate:"required,longitude"`
	Address   string `json:"address" validate:"required"`
}

// UpdateRiderRequest a type for updaing a rider request
type UpdateRiderRequest struct {
	Id      primitive.ObjectID `json:"id" validate:"required"` // required a mongodb objectId
	Name    string             `json:"name" validate:"required"`
	EmailId string             `json:"email_id" validate:"required,email"`

	// applied validation on indian numbers in the format of +91999999999, allowing both 9 and 10-digit numbers
	PhoneNumber string `json:"phone_number" validate:"required,e164,min=12,max=13,startswith=+91"`

	Latitude  string `json:"latitude" validate:"required,latitude"`
	Longitude string `json:"longitude" validate:"required,longitude"`
	Address   string `json:"address" validate:"required"`
	Status    string `json:"status"`
}

// GetRiderRequest a type for updaing a rider request
type GetRiderRequest struct {
	Id primitive.ObjectID `query:"id" validate:"required"` // required a mongodb objectId
}

// CreateRider register new rider
func (request *CreateRiderRequest) CreateRider(ctx context.Context, param RiderParam) (model.Rider, error) {
	currTime := time.Now()
	// Create Rider
	rider := model.Rider{
		Name:        request.Name,
		PhoneNumber: request.PhoneNumber,
		EmailId:     request.EmailId,
		Location:    model.NewLocationFromLongLatStr(request.Longitude, request.Latitude),
		Address:     request.Address,
		Status:      "ACTIVE",
		CreatedAt:   currTime,
		UpdatedAt:   currTime,
	}

	createdRecord, err := param.Repository.CreateRider(ctx, rider)
	if err != nil {
		return model.Rider{}, err
	}

	return createdRecord, nil
}

// UpdateRider update a new rider
func (request *UpdateRiderRequest) UpdateRider(ctx context.Context, param RiderParam) error {
	currTime := time.Now()

	rider, err := param.Repository.GetRider(ctx, request.Id)
	if err != nil {
		return err
	}

	// Update Rider
	rider.Name = request.Name
	rider.EmailId = request.EmailId
	rider.PhoneNumber = request.PhoneNumber
	rider.Location = model.NewLocationFromLongLatStr(request.Longitude, request.Latitude)
	rider.Address = request.Address
	rider.Status = request.Status
	rider.UpdatedAt = currTime

	err = param.Repository.UpdateRider(ctx, rider)
	if err != nil {
		return err
	}

	return nil
}

// GetRider getting a rider
func (request *GetRiderRequest) GetRider(ctx context.Context, param RiderParam) (model.Rider, error) {
	rider, err := param.Repository.GetRider(ctx, request.Id)
	if err != nil {
		return model.Rider{}, err
	}

	return rider, nil
}

// DeleteRider deleting  a rider
func (request *GetRiderRequest) DeleteRider(ctx context.Context, param RiderParam) error {
	err := param.Repository.DeleteRider(ctx, request.Id)
	if err != nil {
		return err
	}

	return nil
}

// SearchRider rider search
func (request *RiderSearchRequest) SearchRider(ctx context.Context, param RiderParam) (*model.PageResponse[model.RiderSearchResponse], error) {
	searchBuilder := model.SearchRiderQuery{}
	searchBuilder.Longitude = request.Longitude
	searchBuilder.Latitude = request.Latitude

	riders, err := param.Repository.SearchRider(ctx, searchBuilder)
	if err != nil {
		return nil, err
	}

	pageResponse := model.NewPageResponse(riders, 10, 10, 10)

	return pageResponse, nil
}
