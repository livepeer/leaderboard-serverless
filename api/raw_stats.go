package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"livepeer.org/leaderboard/common"
	"livepeer.org/leaderboard/db"
	"livepeer.org/leaderboard/models"
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resultsEncoded)
}
