package handler

import (
	"encoding/json"
	"net/http"

	"github.com/livepeer/leaderboard-serverless/common"
	"github.com/livepeer/leaderboard-serverless/db"
	"github.com/livepeer/leaderboard-serverless/middleware"
	"github.com/livepeer/leaderboard-serverless/models"
	"github.com/livepeer/leaderboard-serverless/score"
)

// AggregatedStatsHandler handles an aggregated leaderboard stats request
func AggregatedStatsHandler(w http.ResponseWriter, r *http.Request) {
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

	// since we need to get the median RTT for all Orchs
	// we will check if a specific orch was requested
	// and filter them out after we get the aggregated stats
	orchestrator := statsQuery.Orchestrator
	if orchestrator != "" {
		statsQuery.Orchestrator = ""
	}

	aggrStatResult, err := db.Store.AggregatedStats(statsQuery)
	if err != nil {
		common.HandleInternalError(w, err)
		return
	}

	results := score.CreateAggregatedStats(aggrStatResult)

	// if a specific orchestrator was requested, filter out the rest
	if orchestrator != "" {
		orchStats, ok := results[orchestrator]
		if !ok {
			results = make(map[string]map[string]*models.AggregatedStats)
		} else {
			results = map[string]map[string]*models.AggregatedStats{
				orchestrator: orchStats,
			}
		}
	}

	resultsEncoded, err := json.Marshal(results)
	if err != nil {
		common.HandleInternalError(w, err)
		return
	}

	common.Logger.Trace("Returning aggregated stats: %s", resultsEncoded)

	w.WriteHeader(http.StatusOK)
	w.Write(resultsEncoded)
}
