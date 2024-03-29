package model_test

import (
	"context"
	"errors"
	"food-eats/cmd/web/model"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var notFoundErr = errors.New("user not found")

// InMemoryUserRepository implements UserRepository interface using in-memory storage
type InMemoryUserRepository struct {
	users map[primitive.ObjectID]model.User
}

func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		users: make(map[primitive.ObjectID]model.User),
	}
}

func (r *InMemoryUserRepository) CreateUser(ctx context.Context, user model.User) (model.User, error) {
	user.Id = primitive.NewObjectID()
	r.users[user.Id] = user

	return user, nil
}

func (r *InMemoryUserRepository) UpdateUser(ctx context.Context, user model.User) error {
	_, ok := r.users[user.Id]
	if !ok {
		return notFoundErr
	}

	// Update user in the in-memory map
	r.users[user.Id] = user

	return nil
}

func (r *InMemoryUserRepository) GetUser(ctx context.Context, id primitive.ObjectID) (model.User, error) {
	user, ok := r.users[id]
	if !ok {
		return model.User{}, notFoundErr
	}
	return user, nil
}

func (r *InMemoryUserRepository) DeleteUser(ctx context.Context, id primitive.ObjectID) error {
	_, ok := r.users[id]
	if !ok {
		return notFoundErr
	}

	delete(r.users, id)

	return nil
}

func (r *InMemoryUserRepository) UpdateAverageRating(ctx context.Context, id primitive.ObjectID, rating float64) error {
	user, ok := r.users[id]
	if !ok {
		return notFoundErr
	}

	// Update average rating of the user in the in-memory map
	user.AverageRating = rating
	r.users[id] = user

	return nil
}

func TestInMemoryUserRepository(t *testing.T) {
	repo := NewInMemoryUserRepository()

	user := model.User{
		Name:        "Test User",
		EmailId:     "test@example.com",
		PhoneNumber: "1234567890",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Address:     "Test Address",
		Location: model.Location{
			Type: "Point",
			Coordinates: []float64{
				78.91011,
				12.3456,
			},
		},
		AverageRating: 0.0,
		Status:        "ACTIVE",
	}
	createdUser, err := repo.CreateUser(context.Background(), user)
	assert.NoError(t, err)
	assert.NotNil(t, createdUser)
	assert.NotEmpty(t, createdUser.Id)

	// Test GetUser method
	retrievedUser, err := repo.GetUser(context.Background(), createdUser.Id)
	assert.NoError(t, err)
	assert.Equal(t, createdUser, retrievedUser)

	// Test UpdateUser method
	updatedUser := createdUser
	updatedUser.Name = "Updated Test User"
	err = repo.UpdateUser(context.Background(), updatedUser)
	assert.NoError(t, err)

	// Test DeleteUser method
	err = repo.DeleteUser(context.Background(), updatedUser.Id)
	assert.NoError(t, err)
	_, err = repo.GetUser(context.Background(), updatedUser.Id)
	assert.ErrorIs(t, err, notFoundErr)
}
