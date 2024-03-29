package handlers

import (
	"context"
	"food-eats/cmd/web/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// CreateUserRequest a type for creating a user request
type CreateUserRequest struct {
	Name    string `json:"name" validate:"required"`
	EmailId string `json:"email_id" validate:"required,email"`

	// applied validation on indian numbers in format of +91999999999, allowing both 9 and 10 digit numbers
	PhoneNumber string `json:"phone_number" validate:"required,e164,min=12,max=13,startswith=+91"`

	Latitude  string `json:"latitude" validate:"required,latitude"`
	Longitude string `json:"longitude" validate:"required,longitude"`
	Address   string `json:"address" validate:"required"`
}

// UpdateUserRequest a type for updating a user request
type UpdateUserRequest struct {
	Id      primitive.ObjectID `json:"id" validate:"required"` // required a mongodb objectId
	Name    string             `json:"name" validate:"required"`
	EmailId string             `json:"email_id" validate:"required,email"`

	// applied validation on indian numbers in format of +91999999999, allowing both 9 and 10 digit numbers
	PhoneNumber string `json:"phone_number" validate:"required,e164,min=12,max=13,startswith=+91"`

	Latitude  string `json:"latitude" validate:"required,latitude"`
	Longitude string `json:"longitude" validate:"required,longitude"`
	Address   string `json:"address" validate:"required"`
}

type UserParam struct {
	Repository model.UserRepository
}

// GetUserRequest a type for getting a user request
type GetUserRequest struct {
	Id primitive.ObjectID `query:"id" validate:"required"` // required a mongodb objectId
}

// CreateUser register new user
func (request *CreateUserRequest) CreateUser(ctx context.Context, param UserParam) (model.User, error) {
	currTime := time.Now()
	// Create UserId
	user := model.User{
		Name:        request.Name,
		EmailId:     request.EmailId,
		PhoneNumber: request.PhoneNumber,
		Location:    model.NewLocationFromLongLatStr(request.Longitude, request.Latitude),
		Address:     request.Address,
		Status:      "ACTIVE",
		CreatedAt:   currTime,
		UpdatedAt:   currTime,
	}

	createdRecord, err := param.Repository.CreateUser(ctx, user)
	if err != nil {
		return model.User{}, err
	}

	return createdRecord, nil
}

// UpdateUser update a new user
func (request *UpdateUserRequest) UpdateUser(ctx context.Context, param UserParam) error {
	currTime := time.Now()

	user, err := param.Repository.GetUser(ctx, request.Id)
	if err != nil {
		return err
	}

	// Update UserId
	user.Name = request.Name
	user.EmailId = request.EmailId
	user.PhoneNumber = request.PhoneNumber
	user.Location = model.NewLocationFromLongLatStr(request.Longitude, request.Latitude)
	user.Address = request.Address
	user.UpdatedAt = currTime

	err = param.Repository.UpdateUser(ctx, user)
	if err != nil {
		return err
	}

	return nil
}

// GetUser getting a user
func (request *GetUserRequest) GetUser(ctx context.Context, param UserParam) (model.User, error) {
	user, err := param.Repository.GetUser(ctx, request.Id)
	if err != nil {
		return model.User{}, err
	}

	return user, nil
}

// DeleteUser deleting  a user
func (request *GetUserRequest) DeleteUser(ctx context.Context, param UserParam) error {
	err := param.Repository.DeleteUser(ctx, request.Id)
	if err != nil {
		return err
	}

	return nil
}
