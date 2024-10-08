package handler

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/livepeer/leaderboard-serverless/common"
	"github.com/livepeer/leaderboard-serverless/db"
	"github.com/livepeer/leaderboard-serverless/models"
	"github.com/livepeer/leaderboard-serverless/testutils"
)

func TestTopAIScoreHandler(t *testing.T) {

	// setup data
	testStats := testutils.GetTranscodingStats()

	aiTestFailingStats := testutils.GetAIStats()

	aiTestStatsPassingSlow := testutils.GetAIStats()
	aiTestStatsPassingSlow.SuccessRate = 1
	aiTestStatsPassingSlow.RoundTripTime = 24.1
	aiTestStatsPassingSlow.Region = "FRA"
	aiTestStatsPassingSlow.Model = "model1"
	aiTestStatsPassingSlow.Pipeline = "pipeline1"

	aiTestBestStats := testutils.GetBestAIStats()

	aiTestBestStatsSecondOrch := aiTestBestStats
	aiTestBestStatsSecondOrch.Orchestrator = "orch2"
	aiTestBestStatsSecondOrch.Model = "model2"
	aiTestBestStatsSecondOrch.Pipeline = "pipeline2"

	// test cases
	tests := []struct {
		name                    string
		orchToTest              string
		expectedStatus          int
		expectedBody            string
		statsToInsertBeforeTest []*models.Stats
	}{
		{
			name:                    "Top score for AI",
			orchToTest:              aiTestBestStats.Orchestrator,
			statsToInsertBeforeTest: []*models.Stats{&aiTestFailingStats, &aiTestStatsPassingSlow, &aiTestBestStats},
			expectedStatus:          http.StatusOK,
			expectedBody:            fmt.Sprintf(`{"region":"%s","orchestrator":"%s","value":0.9299999999999999,"model":"%s","pipeline":"%s"}`, aiTestBestStats.Region, aiTestBestStats.Orchestrator, aiTestBestStats.Model, aiTestBestStats.Pipeline),
		},
		{
			name:                    "Top score for AI - different region",
			orchToTest:              aiTestStatsPassingSlow.Orchestrator,
			statsToInsertBeforeTest: []*models.Stats{&aiTestFailingStats, &aiTestStatsPassingSlow},
			expectedStatus:          http.StatusOK,
			expectedBody:            fmt.Sprintf(`{"region":"%s","orchestrator":"%s","value":0.9299999999999999,"model":"%s","pipeline":"%s"}`, aiTestStatsPassingSlow.Region, aiTestStatsPassingSlow.Orchestrator, aiTestStatsPassingSlow.Model, aiTestStatsPassingSlow.Pipeline),
		},
		{
			name:                    "Top score for AI - different orchs",
			orchToTest:              aiTestBestStatsSecondOrch.Orchestrator,
			statsToInsertBeforeTest: []*models.Stats{&aiTestBestStatsSecondOrch, &aiTestStatsPassingSlow},
			expectedStatus:          http.StatusOK,
			expectedBody:            fmt.Sprintf(`{"region":"%s","orchestrator":"%s","value":0.9299999999999999,"model":"%s","pipeline":"%s"}`, aiTestBestStatsSecondOrch.Region, aiTestBestStatsSecondOrch.Orchestrator, aiTestBestStatsSecondOrch.Model, aiTestBestStatsSecondOrch.Pipeline),
		},
		{
			name:                    "No top score for AI",
			orchToTest:              testStats.Orchestrator,
			statsToInsertBeforeTest: []*models.Stats{&testStats},
			expectedStatus:          http.StatusOK,
			expectedBody:            "{}",
		},
	}

	runTopAIScoreTests(t, tests)
}

func runTopAIScoreTests(t *testing.T, tests []struct {
	name                    string
	orchToTest              string
	expectedStatus          int
	expectedBody            string
	statsToInsertBeforeTest []*models.Stats
}) {

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			common.Logger.Info("Running test: %v", tt.name)
			testutils.NewDB(t)

			// insert the stats before the test
			for _, stats := range tt.statsToInsertBeforeTest {
				if err := db.Store.InsertStats(stats); err != nil {
					t.Fatalf("Unexpected error when inserting stats: %v", err)
				}
			}

			// Create a new HTTP request with query parameters
			req, err := http.NewRequest("POST", "/best-ai-stats?orchestrator="+tt.orchToTest, bytes.NewBuffer([]byte{}))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			// Create a ResponseRecorder to record the response
			rr := httptest.NewRecorder()

			// Call the handler
			handler := http.HandlerFunc(TopAiScoreHandler)
			handler.ServeHTTP(rr, req)

			// Check the status code
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("Handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			// Compare the response body with the expected body
			if rr.Body.String() != tt.expectedBody {
				t.Errorf("Handler returned unexpected body: got %v want %v", rr.Body.String(), tt.expectedBody)
			}
		})
	}
}
