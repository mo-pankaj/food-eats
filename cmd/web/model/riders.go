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

type Rider struct {
	Id          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name        string             `json:"name" bson:"name"`
	EmailId     string             `json:"email_id" bson:"emailId"`
	PhoneNumber string             `json:"phone_number" bson:"phoneNumber"`
	CreatedAt   time.Time          `json:"created_at" bson:"createdAt"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updatedAt"`

	// default address fields for verification
	Address       string   `json:"address" bson:"address"`
	Location      Location `json:"location" bson:"location"`
	AverageRating float64  `json:"averageRating" bson:"averageRating"`

	Status string `json:"status" bson:"status"`
}

type SearchRiderQuery struct {
	Latitude  float64
	Longitude float64
	Limit     int
}

// RiderRepository will be the rider repository, a database needs to implement this contract
type RiderRepository interface {
	CreateRider(ctx context.Context, rider Rider) (Rider, error)
	UpdateRider(ctx context.Context, rider Rider) error
	GetRider(ctx context.Context, id primitive.ObjectID) (Rider, error)
	DeleteRider(ctx context.Context, id primitive.ObjectID) error
	SearchRider(ctx context.Context, query SearchRiderQuery) ([]RiderSearchResponse, error)
	UpdateAverageRating(ctx context.Context, id primitive.ObjectID, rating float64) error
}

func RiderMongoRepo(DB *mongo.Database) RiderMongoDb {
	return RiderMongoDb{DB: DB}
}

// RiderMongoDb type with embedded mongo.Database
type RiderMongoDb struct {
	// todo can make this private using type using dependency injection
	DB *mongo.Database
}

func (u RiderMongoDb) CreateRider(ctx context.Context, rider Rider) (Rider, error) {
	insertedId, err := u.DB.Collection("Rider").InsertOne(ctx, rider) // ignoring the parameter as we don't require it
	if err != nil {
		return Rider{}, err
	}
	if insertedId == nil {
		return Rider{}, errors.Join(errors2.ServerError, errors.New("empty inserted id"))
	}
	id, _ := (insertedId.InsertedID).(primitive.ObjectID)
	rider.Id = id
	return rider, nil
}

func (u RiderMongoDb) UpdateRider(ctx context.Context, rider Rider) error {
	// todo instead of whole object set, we can use individual fields set
	updateResult, err := u.DB.Collection("Rider").UpdateOne(ctx, bson.M{"_id": rider.Id}, bson.M{"$set": rider})
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

func (u RiderMongoDb) GetRider(ctx context.Context, id primitive.ObjectID) (Rider, error) {
	var rider Rider
	err := u.DB.Collection("Rider").FindOne(ctx, bson.M{"_id": id}).Decode(&rider)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return rider, errors.Join(errors2.ClientError, errors.New("no such entry"))
		}
		return rider, errors.Join(errors2.ServerError, err)
	}
	return rider, nil
}

// DeleteRider if we want to delete a rider, for now we are marking it as deleted, to delete as a whole we need to delete all records for that person
func (u RiderMongoDb) DeleteRider(ctx context.Context, id primitive.ObjectID) error {
	updateResult, err := u.DB.Collection("Rider").UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"status": "DELETED", "updatedAt": time.Now()}})
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

type RiderSearchResponse struct {
	Id       primitive.ObjectID `json:"id" bson:"_id"`
	Name     string             `json:"name" bson:"name"`
	Distance float64            `json:"distance" bson:"distance"`
}

// SearchRider after accepting order search rider
func (u RiderMongoDb) SearchRider(ctx context.Context, query SearchRiderQuery) ([]RiderSearchResponse, error) {
	var pipeline []bson.M

	matchStage := bson.M{"$match": bson.M{"status": "ACTIVE"}}

	// using projection to limit data size
	projectionStage := bson.M{"$project": bson.M{
		"_id":      1,
		"name":     1,
		"distance": 1,
	}}
	pipeline = append(pipeline, matchStage, projectionStage)

	// near to user query
	if query.Latitude != 0 && query.Longitude != 0 {
		geoNearStage := bson.M{
			"$geoNear": bson.M{
				"near":          bson.M{"type": "Point", "coordinates": []float64{query.Longitude, query.Latitude}},
				"distanceField": "distance",
				"maxDistance":   10 * 1000, // default searching in 10km radius only
				"spherical":     true,
			},
		}
		pipeline = append([]bson.M{geoNearStage}, pipeline...)
	}
	// sorting by reverse rating
	sortQuery := bson.M{"distance": 1, "rating": -1}
	pipeline = append(pipeline, bson.M{"$sort": sortQuery})
	pipeline = append(pipeline, bson.M{"$limit": query.Limit})

	cursor, err := u.DB.Collection("Rider").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var riders []RiderSearchResponse
	for cursor.Next(ctx) {
		var restaurant RiderSearchResponse
		if err := cursor.Decode(&restaurant); err != nil {
			return nil, err
		}
		riders = append(riders, restaurant)
	}

	slog.Info("pipeline", "p", pipeline)

	return riders, nil
}

func (u RiderMongoDb) UpdateAverageRating(ctx context.Context, id primitive.ObjectID, rating float64) error {

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
