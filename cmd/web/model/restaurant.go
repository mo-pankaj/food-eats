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

// Item Individual food item
type Item struct {
	Name        string  `json:"name" bson:"name"`
	Description string  `json:"description" bson:"description"`
	Price       float32 `json:"price" bson:"price"`
	ItemType    string  `json:"item_type" bson:"itemType"`
}

// Menu of a food delivery app
type Menu struct {
	Items []Item `bson:"items" json:"items"`
}

// Restaurant hold information for Restaurant
type Restaurant struct {
	Id          primitive.ObjectID `json:"id" bson:"_id,omitempty"` // omitempty so a mongo-driver can generate unique id
	Name        string             `json:"name" bson:"name"`
	EmailId     string             `json:"email_id" bson:"emailId"`
	PhoneNumber string             `json:"phone_number" bson:"phoneNumber"`
	Website     string             `json:"website" bson:"website,omitempty"`
	CreatedAt   time.Time          `json:"created_At" bson:"createdAt"`
	UpdatedAt   time.Time          `bson:"updated_at" bson:"UpdatedAt"`

	// Default address fields
	Address       string   `json:"address" bson:"address"`
	Location      Location `json:"location" bson:"location"`
	AverageRating float64  `json:"average_rating" bson:"averageRating"`
	AverageTime   float64  `json:"-" bson:"averageTime"`

	Status   string `json:"status" bson:"status"`
	Cuisines string `json:"cuisines" bson:"cuisines"`
	MealType string `json:"meal_type" bson:"mealType"`

	Menu Menu `json:"menu" bson:"menu"`
}

// RestaurantSearchResponse response used to serve to users on search
// to view a menu each restaurant api is called
type RestaurantSearchResponse struct {
	Id                string  `json:"id" bson:"_id"`
	Name              string  `json:"name" bson:"name"`
	Status            string  `json:"status" bson:"status"`
	Cuisines          string  `json:"cuisines" bson:"cuisines"`
	MealType          string  `json:"meal_type" bson:"mealType"`
	Distance          float64 `json:"distance" bson:"distance"`
	AverageRating     float64 `json:"average_rating" bson:"averageRating"`
	AverageTime       float64 `json:"-" bson:"averageTime"`
	AverageTimeToUser float64 `json:"average_time_to_user" bson:"-"`
}

type totalResult struct {
	totalCount int64 `bson:"totalCount"`
}

type SearchRestaurantQuery struct {
	Name        string
	CuisineType string
	MealType    string
	MinRating   float64
	Latitude    float64
	Longitude   float64
	Radius      float64
	SortBy      string
	Limit       int
	Skip        int
}

// RestaurantRepository will be the restaurant repository, a database needs to implement this contract
type RestaurantRepository interface {
	CreateRestaurant(ctx context.Context, restaurant Restaurant) (Restaurant, error)
	UpdateRestaurant(ctx context.Context, restaurant Restaurant) error
	GetRestaurant(ctx context.Context, id primitive.ObjectID) (Restaurant, error)
	DeleteRestaurant(ctx context.Context, id primitive.ObjectID) error
	SearchRestaurant(ctx context.Context, query SearchRestaurantQuery) ([]RestaurantSearchResponse, int64, error)
	UpdateAverageRating(ctx context.Context, id primitive.ObjectID, rating float64) error
	UpdateAverageDeliveryTime(ctx context.Context, id primitive.ObjectID, delivery float64) error
}

// RestaurantMongo type with embedded mongo.Database
type RestaurantMongo struct {
	// todo can make this private using type using dependency injection
	DB *mongo.Database
}

// RestaurantMongoRepo create new mongo DB
func RestaurantMongoRepo(db *mongo.Database) *RestaurantMongo {
	return &RestaurantMongo{db}
}

func (u RestaurantMongo) CreateRestaurant(ctx context.Context, restaurant Restaurant) (Restaurant, error) {
	insertedId, err := u.DB.Collection("Restaurant").InsertOne(ctx, restaurant) // ignoring the parameter as we don't require it
	if err != nil {
		return Restaurant{}, err
	}
	if insertedId == nil {
		return Restaurant{}, errors.Join(errors2.ServerError, errors.New("empty inserted id"))
	}
	id, _ := (insertedId.InsertedID).(primitive.ObjectID)
	restaurant.Id = id
	return restaurant, nil
}

func (u RestaurantMongo) UpdateRestaurant(ctx context.Context, restaurant Restaurant) error {
	// todo instead of whole object set, we can use individual fields set
	updateResult, err := u.DB.Collection("Restaurant").UpdateOne(ctx, bson.M{"_id": restaurant.Id}, bson.M{"$set": restaurant})
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

func (u RestaurantMongo) GetRestaurant(ctx context.Context, id primitive.ObjectID) (Restaurant, error) {
	var restaurant Restaurant
	err := u.DB.Collection("Restaurant").FindOne(ctx, bson.M{"_id": id}).Decode(&restaurant)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return restaurant, errors.Join(errors2.ClientError, errors.New("no such entry"))
		}
		return restaurant, errors.Join(errors2.ServerError, err)
	}
	return restaurant, nil
}

// DeleteRestaurant if we want to delete a restaurant, for now we are marking it as deleted, to delete as a whole we need to delete all records for that person
func (u RestaurantMongo) DeleteRestaurant(ctx context.Context, id primitive.ObjectID) error {
	updateResult, err := u.DB.Collection("Restaurant").UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"status": "DELETED", "updatedAt": time.Now()}})
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

// SearchRestaurant search restaurants
// FIXME: found issues with month Count document, unable to query using count document, raise an issue with mongodb reference https://github.com/mongodb/laravel-mongodb/issues/1819
func (u RestaurantMongo) SearchRestaurant(ctx context.Context, query SearchRestaurantQuery) ([]RestaurantSearchResponse, int64, error) {
	var pipeline []bson.M

	matchStage := bson.M{"$match": bson.M{}}

	// match query
	if query.Name != "" {
		matchStage["$match"].(bson.M)["name"] = query.Name
	}
	if query.CuisineType != "" {
		matchStage["$match"].(bson.M)["cuisineType"] = query.CuisineType
	}
	if query.MealType != "" {
		matchStage["$match"].(bson.M)["mealType"] = query.MealType
	}
	if query.MinRating > 0 {
		matchStage["$match"].(bson.M)["rating"] = bson.M{"$gte": query.MinRating}
	}

	pipeline = append(pipeline, matchStage)

	countPipeline := make([]bson.M, len(pipeline))
	copy(countPipeline, pipeline)
	countPipeline = append(countPipeline, bson.M{"$count": "totalCount"})

	countCursor, err := u.DB.Collection("Restaurant").Aggregate(ctx, countPipeline)
	if err != nil {
		return nil, 0, err
	}
	defer countCursor.Close(ctx)

	var totalCount int64
	if countCursor.Next(ctx) {
		var tR totalResult
		if err := countCursor.Decode(&tR); err != nil {
			return nil, 0, err
		}
		totalCount = tR.totalCount
	}

	// using projection to limit data size
	projectionStage := bson.M{"$project": bson.M{
		"_id":         1,
		"name":        1,
		"cuisines":    1,
		"status":      1,
		"mealType":    1,
		"distance":    1,
		"averageTime": 1,
	}}
	pipeline = append(pipeline, projectionStage)

	// near to user query
	if query.Latitude != 0 && query.Longitude != 0 && query.Radius > 0 {
		geoNearStage := bson.M{
			"$geoNear": bson.M{
				"near":          bson.M{"type": "Point", "coordinates": []float64{query.Longitude, query.Latitude}},
				"distanceField": "distance",
				"maxDistance":   query.Radius * 1000, // converting Km into meters
				"spherical":     true,
			},
		}
		pipeline = append([]bson.M{geoNearStage}, pipeline...)
	}
	sortQuery := bson.M{}
	if query.SortBy != "" {
		sortQuery[query.SortBy] = 1
	}
	pipeline = append(pipeline, bson.M{"$sort": sortQuery})
	pipeline = append(pipeline, bson.M{"$skip": query.Skip})
	pipeline = append(pipeline, bson.M{"$limit": query.Limit})
	cursor, err := u.DB.Collection("Restaurant").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var restaurants []RestaurantSearchResponse
	for cursor.Next(ctx) {
		var restaurant RestaurantSearchResponse
		if err := cursor.Decode(&restaurant); err != nil {
			return nil, 0, err
		}
		restaurant.AverageTimeToUser = restaurant.AverageTime * restaurant.Distance / 1000.0
		restaurants = append(restaurants, restaurant)
	}

	return restaurants, totalCount, nil
}

func (u RestaurantMongo) UpdateAverageRating(ctx context.Context, id primitive.ObjectID, rating float64) error {
	updateResult, err := u.DB.Collection("Restaurant").UpdateByID(ctx, id, bson.M{"$set": bson.M{"averageRating": rating}})
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

func (u RestaurantMongo) UpdateAverageDeliveryTime(ctx context.Context, id primitive.ObjectID, deliveryTime float64) error {
	updateResult, err := u.DB.Collection("Restaurant").UpdateByID(ctx, id, bson.M{"$set": bson.M{"averageDeliveryTime": deliveryTime}})
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
