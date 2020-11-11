package handler

import (
	"encoding/json"
	"net/http"

	"github.com/livepeer/leaderboard-serverless/common"
	"github.com/livepeer/leaderboard-serverless/db"
	"github.com/livepeer/leaderboard-serverless/models"
)

// AggregatedStatsHandler handles an aggregated leaderboard stats request
func AggregatedStatsHandler(w http.ResponseWriter, r *http.Request) {
	if err := db.CacheDB(); err != nil {
		common.HandleInternalError(w, err)
		return
	}

	//Get the path parameter that was sent
	query := r.URL.Query()
	orch := query.Get("orchestrator")
	region := query.Get("region")
	since := query.Get("since")

	searchRegions := models.Regions
	if region != "" {
		searchRegions = []string{region}
	}

	results := make(map[string]map[string]*models.AggregatedStats)

	// TODO: Make this concurrent
	for _, region := range searchRegions {
		stats, err := db.Store.AggregatedStats(orch, region, since)
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
	w.Header().Set("Cache-Control", "max-age=60, public")
	w.WriteHeader(http.StatusOK)
	w.Write(resultsEncoded)
}
