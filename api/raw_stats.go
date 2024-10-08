package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/livepeer/leaderboard-serverless/common"
	"github.com/livepeer/leaderboard-serverless/db"
	"github.com/livepeer/leaderboard-serverless/middleware"
	"github.com/livepeer/leaderboard-serverless/models"
)

// RawStatsHandler handles a request for raw leaderboard stats
// orchestrator parameter is required
func RawStatsHandler(w http.ResponseWriter, r *http.Request) {
	if err := db.CacheDB(); err != nil {
		common.HandleInternalError(w, err)
		return
	}

	middleware.AddStandardHttpHeaders(w)

	statsQuery, err := common.ParseStatsQueryParams(r)
	if err != nil {
		common.HandleBadRequest(w, err)
		return
	}

	if statsQuery.Orchestrator == "" {
		common.HandleBadRequest(w, errors.New("orchestrator is a required parameter"))
		return
	}

	stats, err := db.Store.RawStats(statsQuery)
	if err != nil {
		common.HandleInternalError(w, err)
		return
	}
	resultsEncoded, err := CreateRawStats(stats)
	if err != nil {
		common.HandleInternalError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resultsEncoded)
}

// CreateRawStats creates a map of raw stats by region
func CreateRawStats(stats []*models.Stats) ([]byte, error) {
	results := make(map[string][]*models.Stats)
	for _, stat := range stats {
		results[stat.Region] = append(results[stat.Region], stat)
	}
	resultsEncoded, err := json.Marshal(results)
	if err != nil {
		return nil, err
	}
	return resultsEncoded, nil
}
