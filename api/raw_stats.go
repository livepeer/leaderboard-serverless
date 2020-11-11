package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/livepeer/leaderboard-serverless/common"
	"github.com/livepeer/leaderboard-serverless/db"
	"github.com/livepeer/leaderboard-serverless/models"
)

// RawStatsHandler handles a request for raw leaderboard stats
// orchestrator parameter is required
func RawStatsHandler(w http.ResponseWriter, r *http.Request) {
	if err := db.CacheDB(); err != nil {
		common.HandleInternalError(w, err)
		return
	}

	//Get the path parameter that was sent
	query := r.URL.Query()
	orch := query.Get("orchestrator")
	region := query.Get("region")
	since := query.Get("since")

	if orch == "" {
		common.HandleBadRequest(w, errors.New("orchestrator is a required parameter"))
		return
	}

	searchRegions := models.Regions
	if region != "" {
		searchRegions = []string{region}
	}

	results := make(map[string][]*models.Stats)

	// TODO: Make this concurrent
	for _, region := range searchRegions {
		stats, err := db.Store.RawStats(orch, region, since)
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

	// Set CORS headers for the preflight request
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resultsEncoded)
}
