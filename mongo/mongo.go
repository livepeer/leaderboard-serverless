package mongo

import (
	"context"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Client for MongoDB connections
var Client *mongo.Client

// DB used for reading and writing data
var DB *mongo.Database

// Regions (collections)
var Regions = []string{"MDW", "FRA", "SIN"}

// Start a new MongoDB client connection
func Start(ctx context.Context) error {
	var err error
	Client, err = mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO")))
	if err != nil {
		return err
	}

	// Check whether the connection was succesful by pinging the MongoDB server
	err = Client.Ping(ctx, readpref.Primary())
	if err != nil {
		return err
	}

	DB = Client.Database("leaderboard")
	return nil
}
