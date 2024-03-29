package model

import (
	"context"
	"errors"
	errors2 "food-eats/cmd/web/custom-errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log/slog"
	"time"
)

// User represent a user in database
type User struct {
	Id          primitive.ObjectID `json:"id" bson:"_id,omitempty"` // omitempty so a mongo-driver can generate unique id
	Name        string             `json:"name" bson:"name"`
	EmailId     string             `json:"email_id" bson:"emailId"`
	PhoneNumber string             `json:"phone_number" bson:"phoneNumber"`
	CreatedAt   time.Time          `json:"created_at" bson:"createdAt"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updatedAt"`

	// default address fields
	Address       string   `json:"address" bson:"address"`
	Location      Location `json:"location" bson:"location"`
	AverageRating float64  `json:"averageRating" bson:"averageRating"`

	Status string `json:"status" bson:"status"`
}

// hiding sensitive information while logging
func (c User) LogValue() slog.Value {
	var attributes []slog.Attr
	attributes = append(attributes, slog.Attr{Key: "user_id", Value: slog.AnyValue(c.Id)})
	return slog.GroupValue(attributes...)
}

// UserRepository will be the user repository, a database needs to implement this contract
type UserRepository interface {
	CreateUser(ctx context.Context, user User) (User, error)
	UpdateUser(ctx context.Context, user User) error
	GetUser(ctx context.Context, id primitive.ObjectID) (User, error)
	DeleteUser(ctx context.Context, id primitive.ObjectID) error
	UpdateAverageRating(ctx context.Context, id primitive.ObjectID, rating float64) error
}

func UserMongoRepo(DB *mongo.Database) UserMongoDb {
	return UserMongoDb{DB: DB}
}

// UserMongoDb type with embedded mongo.Database
type UserMongoDb struct {
	// todo can make this private using type using dependency injection
	DB *mongo.Database
}

func (u UserMongoDb) CreateUser(ctx context.Context, user User) (User, error) {
	insertedId, err := u.DB.Collection("User").InsertOne(ctx, user)
	if err != nil {
		return User{}, err
	}
	if insertedId == nil {
		return User{}, errors.Join(errors2.ServerError, errors.New("empty inserted id"))
	}
	id, _ := (insertedId.InsertedID).(primitive.ObjectID)
	user.Id = id
	return user, nil
}

func (u UserMongoDb) UpdateUser(ctx context.Context, user User) error {
	// todo instead of whole object set, we can use individual fields set
	updateResult, err := u.DB.Collection("User").UpdateOne(ctx, bson.M{"_id": user.Id}, bson.M{"$set": user})
	if err != nil {
		return err
	}

	if updateResult == nil {
		return errors.Join(errors2.ServerError, errors.New("no update result"))
	}
	if updateResult.MatchedCount != 1 {
		return errors.Join(errors2.ClientError, errors.New("no matching document"))
	}
	if updateResult.ModifiedCount != 1 {
		return errors.Join(errors2.ServerError, errors.New("update failed"))
	}

	return nil
}

func (u UserMongoDb) GetUser(ctx context.Context, id primitive.ObjectID) (User, error) {
	var user User
	err := u.DB.Collection("User").FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return user, errors.Join(errors2.ClientError, errors.New("no such entry"))
		}
		return user, errors.Join(errors2.ServerError, err)
	}
	return user, nil
}

// DeleteUser if we want to delete a user, for now we are marking it as deleted, to delete as a whole we need to delete all records for that person
func (u UserMongoDb) DeleteUser(ctx context.Context, id primitive.ObjectID) error {
	updateResult, err := u.DB.Collection("User").UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"status": "DELETED", "updatedAt": time.Now()}})
	if err != nil {
		return nil
	}
	if updateResult == nil {
		return errors.Join(errors2.ServerError, errors.New("no update result"))
	}
	if updateResult.MatchedCount != 1 {
		return errors.Join(errors2.ClientError, errors.New("no matching document"))
	}
	if updateResult.ModifiedCount != 1 {
		return errors.Join(errors2.ServerError, errors.New("update failed"))
	}

	return nil
}

func (u UserMongoDb) UpdateAverageRating(ctx context.Context, id primitive.ObjectID, rating float64) error {
	updateResult, err := u.DB.Collection("Rider").UpdateByID(ctx, id, bson.M{"$set": bson.M{"averageRating": rating}})
	if err != nil {
		return err
	}

	if updateResult == nil {
		return errors.Join(errors2.ServerError, errors.New("no update result"))
	}
	if updateResult.MatchedCount != 1 {
		return errors.Join(errors2.ClientError, errors.New("no matching document"))
	}
	if updateResult.ModifiedCount != 1 {
		return errors.Join(errors2.ServerError, errors.New("update failed"))
	}

	return nil
}
