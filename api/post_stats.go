package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/mongo/readpref"

	"livepeer.org/leaderboard/common"
	"livepeer.org/leaderboard/models"
	"livepeer.org/leaderboard/mongo"
)

// Handler function Using AWS Lambda Proxy Request
func PostStatsHandler(w http.ResponseWriter, r *http.Request) {
	var stats models.Stats

	// Unmarshal the json, return 400 if error
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		common.HandleBadRequest(w, err)
		return
	}

	if err := json.Unmarshal([]byte(body), &stats); err != nil {
		common.HandleBadRequest(w, err)
		return
	}

	if !isValidRegion(stats.Region) {
		common.HandleBadRequest(w, errors.New("invalid region"))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	if err := mongo.Client.Ping(ctx, readpref.Primary()); err != nil {
		common.HandleInternalError(w, err)
		return
	}

	_, err = mongo.DB.Collection(stats.Region).InsertOne(ctx, stats)
	if err != nil {
		common.HandleInternalError(w, err)
		return
	}

	//Return inserts Object ID and  200 StatusCode response with AWS Lambda Proxy Response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func isValidRegion(region string) bool {
	for _, reg := range mongo.Regions {
		if reg == region {
			return true
		}
	}
	return false
}
