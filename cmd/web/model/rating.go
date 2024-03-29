package model

import (
	"context"
	"errors"
	errors2 "food-eats/cmd/web/custom-errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type Rating struct {
	Id                 primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	RatingGiver        primitive.ObjectID `bson:"ratingGiver" json:"rating_giver"`
	RatingReceiver     primitive.ObjectID `bson:"ratingReceiver" json:"rating_receiver"`
	OrderID            primitive.ObjectID `bson:"orderId" json:"order_id"`
	RatingGiverType    string             `bson:"ratingGiverType" json:"rating_giver_type"`
	RatingReceiverType string             `bson:"ratingReceiverType" json:"rating_receiver_type"`
	Rating             int                `bson:"rating" json:"rating"`
	CreatedAt          time.Time          `bson:"createdAt" json:"created_at"`
	UpdatedAt          time.Time          `bson:"updatedAt" json:"updated_at"`
	Comment            string             `bson:"comment" json:"comment"`
}

type SearchRatingBuilder struct {
	RatingGiver    primitive.ObjectID
	RatingReceiver primitive.ObjectID
}

// RatingRepository will be the rating repository, a database needs to implement this contract
type RatingRepository interface {
	CreateRating(ctx context.Context, rating Rating) (Rating, error)
	GetRating(ctx context.Context, id primitive.ObjectID) (Rating, error)
	SearchRating(ctx context.Context, query SearchRatingBuilder) ([]Rating, error)
}

func RatingMongoRepo(DB *mongo.Database) RatingMongoDb {
	return RatingMongoDb{DB: DB}
}

// RatingMongoDb type with embedded mongo.Database
type RatingMongoDb struct {
	// todo can make this private using type using dependency injection
	DB *mongo.Database
}

func (u RatingMongoDb) CreateRating(ctx context.Context, rating Rating) (Rating, error) {
	insertedId, err := u.DB.Collection("Rating").InsertOne(ctx, rating)
	if err != nil {
		return Rating{}, err
	}
	if insertedId == nil {
		return Rating{}, errors.Join(errors2.ServerError, errors.New("empty inserted id"))
	}
	id, _ := (insertedId.InsertedID).(primitive.ObjectID)
	rating.Id = id
	return rating, nil
}

func (u RatingMongoDb) GetRating(ctx context.Context, id primitive.ObjectID) (Rating, error) {
	var rating Rating
	err := u.DB.Collection("Rating").FindOne(ctx, bson.M{"_id": id}).Decode(&rating)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return rating, errors.Join(errors2.ClientError, errors.New("no such entry"))
		}
		return rating, errors.Join(errors2.ServerError, err)
	}
	return rating, nil
}

func (u RatingMongoDb) SearchRating(ctx context.Context, query SearchRatingBuilder) ([]Rating, error) {
	filter := bson.M{}
	if !query.RatingGiver.IsZero() {
		filter["ratingGiver"] = query.RatingGiver
	}
	if !query.RatingReceiver.IsZero() {
		filter["ratingReceiver"] = query.RatingReceiver
	}
	cur, err := u.DB.Collection("Rating").Find(ctx, filter)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return []Rating{}, errors.Join(errors2.ClientError, errors.New("no such entry"))
		}
		return []Rating{}, errors.Join(errors2.ServerError, err)
	}

	ratings := []Rating{}
	err = cur.All(ctx, &ratings)
	if err != nil {
		return []Rating{}, err
	}
	return ratings, nil
}
