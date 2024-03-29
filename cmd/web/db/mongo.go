package db

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetMongoClient(ctx context.Context, uri string, db string) (*mongo.Database, error) {
	opts := options.Client().ApplyURI(uri)

	// todo apply custom pool size default is 100, apply compression and decompression for better performance
	client, _ := mongo.Connect(ctx, opts)

	err := client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	// simple option as of now to have IST time stored
	databaseOptions := &options.DatabaseOptions{
		BSONOptions: &options.BSONOptions{
			UseLocalTimeZone: true,
		},
	}
	database := client.Database(db, databaseOptions)

	return database, nil
}
