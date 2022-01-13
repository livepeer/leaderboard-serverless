package mongo

import (
	"context"
	"os"
	"time"

	"github.com/livepeer/leaderboard-serverless/models"
	"go.mongodb.org/mongo-driver/bson"
	mongodb "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type DB struct {
	*mongodb.Database
}

// Start a new MongoDB client connection
func Start() (*DB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var err error
	client, err := mongodb.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO")))
	if err != nil {
		return nil, err
	}

	// Check whether the connection was succesful by pinging the MongoDB server
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, err
	}

	store := client.Database("leaderboard")
	return &DB{store}, nil
}

func (db *DB) InsertStats(stats *models.Stats) error {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	_, err := db.Collection(stats.Region).InsertOne(ctx, stats)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) AggregatedStats(orch, region string, since, until int64) ([]*models.Stats, error) {
	// Query MongoDB
	opts := options.Aggregate()
	opts.SetAllowDiskUse(true)

	pipeline := []bson.M{}

	and := []bson.M{
		{"timestamp": bson.M{
			"$gte": since,
			"$lte": until,
		},
		},
	}

	if orch != "" {
		and = append(and, bson.M{"orchestrator": orch})
	}

	pipeline = append(pipeline, bson.M{
		"$match": bson.M{
			"$and": and,
		},
	})

	grouper := bson.M{
		"$group": bson.M{
			"_id":             "$orchestrator",
			"success_rate":    bson.M{"$avg": "$success_rate"},
			"seg_duration":    bson.M{"$avg": "$seg_duration"},
			"upload_time":     bson.M{"$avg": "$upload_time"},
			"download_time":   bson.M{"$avg": "$download_time"},
			"transcode_time":  bson.M{"$avg": "$transcode_time"},
			"round_trip_time": bson.M{"$avg": "$round_trip_time"},
		},
	}

	pipeline = append(pipeline, grouper)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := db.Collection(region).Aggregate(ctx, pipeline, opts)
	defer cursor.Close(ctx)
	if err != nil {
		return nil, err
	}

	var aggregatedStats []*models.Stats
	for cursor.Next(ctx) {
		var aggregatedStatsDoc models.Stats
		if err := cursor.Decode(&aggregatedStatsDoc); err != nil {
			return nil, err
		}
		aggregatedStats = append(aggregatedStats, &aggregatedStatsDoc)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return aggregatedStats, nil

}

func (db *DB) RawStats(orch, region string, since, until int64) ([]*models.Stats, error) {
	opts := options.Find()
	// Descending: latest first
	opts.SetSort(bson.D{{Key: "timestamp", Value: -1}})
	filter := bson.D{
		{
			Key: "timestamp",
			Value: bson.D{
				{Key: "$gte", Value: since},
				{Key: "$lte", Value: until},
			},
		},
		{
			Key:   "orchestrator",
			Value: bson.D{{Key: "$eq", Value: orch}},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := db.Collection(region).Find(ctx, filter, opts)
	defer cursor.Close(ctx)
	if err != nil {
		return nil, err
	}

	var stats []*models.Stats
	for cursor.Next(ctx) {
		var statsDoc models.Stats
		if err := cursor.Decode(&statsDoc); err != nil {
			return nil, err
		}
		stats = append(stats, &statsDoc)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return stats, nil
}
