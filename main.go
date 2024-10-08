package main

import (
	"net/http"

	handler "github.com/livepeer/leaderboard-serverless/api"
	"github.com/livepeer/leaderboard-serverless/common"
)

// this func is for running in local mode.  Vercel does not use this as an entrypoint
// so any logic here should only reflect what is needed for local development
func main() {

	http.HandleFunc("/api/raw_stats", handler.RawStatsHandler)
	http.HandleFunc("/api/aggregated_stats", handler.AggregatedStatsHandler)
	http.HandleFunc("/api/top_ai_score", handler.TopAiScoreHandler)
	http.HandleFunc("/api/post_stats", handler.PostStatsHandler)
	http.HandleFunc("/api/pipelines", handler.PipelinesHandler)
	http.HandleFunc("/api/regions", handler.RegionsHandler)

	common.Logger.Info("Server starting on port 8080")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		common.Logger.Fatal("Unable to start the server: %v", err)
	}
}
