package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"livepeer.org/leaderboard/common"
	"livepeer.org/leaderboard/models"
	"livepeer.org/leaderboard/mongo"
)

// RawStatsHandler handles a request for raw leaderboard stats
// orchestrator parameter is required
func RawStatsHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if mongo.Client == nil || mongo.DB == nil {
		if err := mongo.Start(ctx); err != nil {
			common.HandleInternalError(w, err)
			return
		}
	}

	pingCtx, pingCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer pingCancel()
	err := mongo.Client.Ping(pingCtx, readpref.Primary())
	if err != nil {
		common.HandleInternalError(w, err)
		return
	}

	//Get the path parameter that was sent
	query := r.URL.Query()
	orch := query.Get("orchestrator")
	region := query.Get("region")
	sinceStr := query.Get("since")

	if orch == "" {
		common.HandleBadRequest(w, err)
		return
	}

	var since int64
	if sinceStr == "" {
		since = time.Now().Add(-24 * time.Hour).Unix()
	} else {
		since, err = strconv.ParseInt(sinceStr, 10, 64)
		if err != nil {
			common.HandleBadRequest(w, err)
			return
		}
	}

	opts := options.Find()
	// Descending: latest first
	opts.SetSort(bson.D{{Key: "timestamp", Value: -1}})
	filter := bson.D{
		{
			Key:   "timestamp",
			Value: bson.D{{Key: "$gte", Value: since}},
		},
		{
			Key:   "orchestrator",
			Value: bson.D{{Key: "$eq", Value: orch}},
		},
	}

	searchRegions := mongo.Regions
	if region != "" {
		searchRegions = []string{region}
	}

	results := make(map[string][]*models.Stats)

	// TODO: Make this concurrent
	for _, region := range searchRegions {
		// Query MongoDB
		queryCtx, queryCancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer queryCancel()
		stats, err := getRegionalStats(queryCtx, region, filter, opts)
		if err != nil {
			common.HandleInternalError(w, err)
			return
		}
		results[region] = stats
	}

	resultsEncoded, err := json.Marshal(results)
	if err != nil {
		common.HandleInternalError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resultsEncoded)
}

func getRegionalStats(ctx context.Context, region string, filter bson.D, opts *options.FindOptions) ([]*models.Stats, error) {
	cursor, err := mongo.DB.Collection(region).Find(ctx, filter, opts)
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
