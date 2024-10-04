package handler

import (
	"encoding/json"
	"net/http"

	"github.com/livepeer/leaderboard-serverless/common"
	"github.com/livepeer/leaderboard-serverless/db"
	"github.com/livepeer/leaderboard-serverless/middleware"
	"github.com/livepeer/leaderboard-serverless/models"
)

// PipelinesHandler handles a request for Pipeline/Model Reference Data
func PipelinesHandler(w http.ResponseWriter, r *http.Request) {
	if err := db.CacheDB(); err != nil {
		common.HandleInternalError(w, err)
		return
	}

	middleware.AddStandardHttpHeaders(w)

	query, err := common.ParseStatsQueryParams(r)
	if err != nil {
		common.HandleBadRequest(w, err)
		return
	}

	pipelines, err := db.Store.Pipelines(query)
	if err != nil {
		common.HandleInternalError(w, err)
		return
	}
	resultsEncoded, err := json.Marshal(map[string][]*models.Pipeline{"pipelines": pipelines})
	if err != nil {
		common.HandleInternalError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resultsEncoded)
}
