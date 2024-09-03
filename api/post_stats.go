package handler

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/livepeer/leaderboard-serverless/common"
	"github.com/livepeer/leaderboard-serverless/db"
	"github.com/livepeer/leaderboard-serverless/middleware"
	"github.com/livepeer/leaderboard-serverless/middleware/auth"
	"github.com/livepeer/leaderboard-serverless/models"
)

// PostStatsHandler function Using AWS Lambda Proxy Request
func PostStatsHandler(w http.ResponseWriter, r *http.Request) {
	if err := db.CacheDB(); err != nil {
		common.HandleInternalError(w, err)
		return
	}

	middleware.HandlePreflightRequest(w, r)

	var stats models.Stats

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		common.HandleBadRequest(w, err)
		return
	}

	if ok := auth.IsAuthorized(
		r.Header.Get("Authorization"),
		body,
	); !ok {
		common.RespondWithError(w, errors.New("request can not be authenticated"), http.StatusForbidden)
		return
	}

	// Unmarshal the json, return 400 if error
	if err := json.Unmarshal([]byte(body), &stats); err != nil {
		common.HandleBadRequest(w, err)
		return
	}

	if !isValidRegion(stats.Region) {
		common.HandleBadRequest(w, errors.New("invalid region"))
		return
	}

	if err := db.Store.InsertStats(&stats); err != nil {
		common.HandleInternalError(w, err)
	}

	//Return inserts Object ID and  200 StatusCode response with AWS Lambda Proxy Response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func isValidRegion(region string) bool {
	for _, reg := range models.GetRegions() {
		if reg == region {
			return true
		}
	}
	return false
}
