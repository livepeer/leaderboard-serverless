package handler

import (
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

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

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=600, stale-while-revalidate=300")

	//Get the path parameter that was sent
	query := r.URL.Query()
	orch := strings.ToLower(query.Get("orchestrator"))
	region := strings.ToUpper(query.Get("region"))
	sinceStr := query.Get("since")
	untilStr := query.Get("until")

	searchRegions := models.Regions
	if region != "" {
		searchRegions = []string{region}
	}

	var since int64
	if sinceStr == "" {
		since = time.Now().Add(-24 * time.Hour).Unix()
	} else {
		var err error
		since, err = strconv.ParseInt(sinceStr, 10, 64)
		if err != nil {
			common.HandleBadRequest(w, err)
		}
	}

	var until int64
	if untilStr == "" {
		until = time.Now().Unix()
	} else {
		var err error
		until, err = strconv.ParseInt(untilStr, 10, 64)
		if err != nil {
			common.HandleBadRequest(w, err)
		}
	}

	results := make(map[string]map[string]*models.AggregatedStats)

	// TODO: Make this concurrent
	for _, region := range searchRegions {
		stats, err := db.Store.AggregatedStats(orch, region, since, until)
		if err != nil {
			common.HandleInternalError(w, err)
			return
		}

		for _, stat := range stats {
			_, ok := results[stat.ID]
			if !ok {
				results[stat.ID] = make(map[string]*models.AggregatedStats)
			}
			aggrStats := &models.AggregatedStats{
				ID:             stat.ID,
				SuccessRate:    stat.SuccessRate,
				RoundTripScore: normalizeLatencyScore(calculateLatencyScore(stat.SegDuration, stat.RoundTripTime)),
			}
			aggrStats.TotalScore = calculateTotalScore(aggrStats)
			results[stat.ID][region] = aggrStats
		}
	}

	resultsEncoded, err := json.Marshal(results)
	if err != nil {
		common.HandleInternalError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resultsEncoded)
}

func calculateTotalScore(stats *models.AggregatedStats) float64 {
	if stats == nil {
		return 0
	}

	return stats.SuccessRate * stats.RoundTripScore
}

func calculateLatencyScore(segDuration, latency float64) float64 {
	if latency == 0 {
		return 0
	}
	return segDuration / latency
}

func normalizeLatencyScore(score float64) float64 {
	return 1 - math.Pow(math.E, -score)
}
