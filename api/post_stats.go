package handler

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"livepeer.org/leaderboard/common"
	"livepeer.org/leaderboard/db"
	"livepeer.org/leaderboard/models"
)

// PostStatsHandler function Using AWS Lambda Proxy Request
func PostStatsHandler(w http.ResponseWriter, r *http.Request) {
	if err := db.CacheDB(); err != nil {
		common.HandleInternalError(w, err)
		return
	}

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

	if err := db.Store.InsertStats(&stats); err != nil {
		common.HandleInternalError(w, err)
	}

	//Return inserts Object ID and  200 StatusCode response with AWS Lambda Proxy Response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func isValidRegion(region string) bool {
	for _, reg := range models.Regions {
		if reg == region {
			return true
		}
	}
	return false
}
