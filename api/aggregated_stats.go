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

// AggregatedStatsHandler handles an aggregated leaderboard stats request
func AggregatedStatsHandler(w http.ResponseWriter, r *http.Request) {
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

	// Query MongoDB
	opts := options.Aggregate()
	opts.SetAllowDiskUse(true)

	pipeline := []bson.M{}

	if orch != "" {
		matcher := bson.M{
			"$match": bson.M{
				"orchestrator": orch,
				"timestamp": bson.M{
					"$gte": since,
				},
			},
		}

		pipeline = append(pipeline, matcher)
	}

	grouper := bson.M{
		"$group": bson.M{
			"_id":          "$orchestrator",
			"score":        bson.M{"$avg": "$round_trip_score"},
			"success_rate": bson.M{"$avg": "$success_rate"},
		},
	}

	pipeline = append(pipeline, grouper)

	searchRegions := mongo.Regions
	if region != "" {
		searchRegions = []string{region}
	}

	results := make(map[string]map[string]*models.AggregatedStats)

	// TODO: Make this concurrent
	for _, region := range searchRegions {
		// Query MongoDB
		queryCtx, queryCancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer queryCancel()
		stats, err := aggregateRegionalStats(queryCtx, region, pipeline, opts)
		if err != nil {
			common.HandleInternalError(w, err)
			return
		}

		for _, stat := range stats {
			_, ok := results[stat.ID]
			if !ok {
				results[stat.ID] = make(map[string]*models.AggregatedStats)
			}
			results[stat.ID][region] = stat
		}
	}

	resultsEncoded, err := json.Marshal(results)
	if err != nil {
		common.HandleInternalError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "max-age=60, public")
	w.WriteHeader(http.StatusOK)
	w.Write(resultsEncoded)
}

func aggregateRegionalStats(ctx context.Context, region string, pipeline []bson.M, opts *options.AggregateOptions) ([]*models.AggregatedStats, error) {
	cursor, err := mongo.DB.Collection(region).Aggregate(ctx, pipeline, opts)
	defer cursor.Close(ctx)
	if err != nil {
		return nil, err
	}

	var aggregatedStats []*models.AggregatedStats
	for cursor.Next(ctx) {
		var aggregatedStatsDoc models.AggregatedStats
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
