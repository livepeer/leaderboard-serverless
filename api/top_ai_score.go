package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/livepeer/leaderboard-serverless/common"
	"github.com/livepeer/leaderboard-serverless/db"
	"github.com/livepeer/leaderboard-serverless/middleware"
	"github.com/livepeer/leaderboard-serverless/models"
	"github.com/livepeer/leaderboard-serverless/score"
)

// TopAiScoreHandler handles a request for the top regional scores
func TopAiScoreHandler(w http.ResponseWriter, r *http.Request) {

	common.Logger.Debug("TopScoresHandler called")

	if err := db.CacheDB(); err != nil {
		common.HandleInternalError(w, err)
		return
	}

	middleware.AddStandardHttpHeaders(w)

	//get orchestratorId from query and build the query
	orchestratorId := r.URL.Query().Get("orchestrator")

	topStatsForOrch, err := db.Store.BestAIRegion(orchestratorId)
	if err != nil {
		common.HandleInternalError(w, err)
		return
	}

	// if no stats found, return empty response
	if topStatsForOrch == nil {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
		return
	}

	// if this is a request for AI jobs, let's get the aggregated stats
	// so we can calculate the top scores using the median RTT

	query := &models.StatsQuery{
		Since:    common.GetDefaultSince(),
		Until:    time.Now().UTC(),
		JobType:  models.AI,
		Model:    topStatsForOrch.Model,
		Pipeline: topStatsForOrch.Pipeline,
	}
	aggrStatResult, err := db.Store.AggregatedStats(query)
	if err != nil {
		common.HandleInternalError(w, err)
		return
	}
	aggregatedStats := score.CreateAggregatedStats(aggrStatResult)

	common.Logger.Debug("Aggregated stats %v", aggregatedStats)

	// now that we have the aggregated stats for this model and pipeline
	// let's find the record for this orchestrator and region
	// so that we can return the top scores
	topScore := models.Score{
		Orchestrator: topStatsForOrch.Orchestrator,
		Region:       topStatsForOrch.Region,
		Value:        aggregatedStats[orchestratorId][topStatsForOrch.Region].TotalScore,
		Model:        topStatsForOrch.Model,
		Pipeline:     topStatsForOrch.Pipeline,
	}

	resultsEncoded, err := json.Marshal(topScore)
	if err != nil {
		common.HandleInternalError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resultsEncoded)
}
