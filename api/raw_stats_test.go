package handler

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/livepeer/leaderboard-serverless/common"
	"github.com/livepeer/leaderboard-serverless/db"
	"github.com/livepeer/leaderboard-serverless/models"
	"github.com/livepeer/leaderboard-serverless/testutils"
)

func TestRawStatsHandler(t *testing.T) {

	// setup data
	testStats := testutils.GetTranscodingStats()

	aiTestFailingStats := testutils.GetAIStats()

	aiTestStatsPassingSlow := testutils.GetAIStats()
	aiTestStatsPassingSlow.SuccessRate = 1
	aiTestStatsPassingSlow.RoundTripTime = 24.1
	aiTestStatsPassingSlow.Region = "FRA"
	aiTestStatsPassingSlow.Model = "model1"
	aiTestStatsPassingSlow.Pipeline = "pipeline1"

	aiTestBestStats := testutils.GetAIStats()
	aiTestBestStats.SuccessRate = 1
	aiTestBestStats.RoundTripTime = 0.1
	aiTestBestStats.Region = "LAX"

	aiTestBestStatsSecondOrch := aiTestBestStats
	aiTestBestStatsSecondOrch.Orchestrator = "orch2"
	aiTestBestStatsSecondOrch.Model = "model2"
	aiTestBestStatsSecondOrch.Pipeline = "pipeline2"

	expectedRawStatsAITC := []*models.Stats{&aiTestFailingStats, &aiTestStatsPassingSlow, &aiTestBestStats}
	//expected results are the expectedRawStatsTC1 expect the aiTestStatsPassingSlow, which has a different model and pipeline
	expectedRawStatsResultsAITC, err := CreateRawStats([]*models.Stats{&aiTestFailingStats, &aiTestBestStats})
	if err != nil {
		t.Fatalf("Unexpected error when creating raw stats: %v", err)
	}
	expectedBodyStringAITC := string(expectedRawStatsResultsAITC)

	expectedRawStatsResultsTransTC, err := CreateRawStats([]*models.Stats{&testStats})
	if err != nil {
		t.Fatalf("Unexpected error when creating raw stats: %v", err)
	}
	expectedBodyStringTransTC := string(expectedRawStatsResultsTransTC)

	// test cases
	tests := []struct {
		name                    string
		orchToTest              *models.Stats
		expectedStatus          int
		expectedBody            string
		statsToInsertBeforeTest []*models.Stats
	}{
		{
			name:                    "Get 2 Raw AI Stats",
			orchToTest:              &aiTestBestStats,
			statsToInsertBeforeTest: expectedRawStatsAITC,
			expectedStatus:          http.StatusOK,
			expectedBody:            expectedBodyStringAITC,
		},
		{
			name:                    "Get 1 Raw Transcoding Stats",
			orchToTest:              &testStats,
			statsToInsertBeforeTest: []*models.Stats{&testStats},
			expectedStatus:          http.StatusOK,
			expectedBody:            expectedBodyStringTransTC,
		},
		{
			name:                    "No stats",
			orchToTest:              &testStats,
			statsToInsertBeforeTest: []*models.Stats{&aiTestFailingStats},
			expectedStatus:          http.StatusOK,
			expectedBody:            "{}",
		},
	}

	runRawStatsTests(t, tests)
}

func runRawStatsTests(t *testing.T, tests []struct {
	name                    string
	orchToTest              *models.Stats
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
			req, err := http.NewRequest("GET",
				fmt.Sprintf("/raw-stats?orchestrator=%s&since=%s&until=%s&model=%s&pipeline=%s",
					tt.orchToTest.Orchestrator, testutils.GetUnixTimeMinusTenSecStr(),
					testutils.GetUnixTimeInFiveSecStr(),
					url.QueryEscape(tt.orchToTest.Model), url.QueryEscape(tt.orchToTest.Pipeline),
				),
				bytes.NewBuffer([]byte{}))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			// Create a ResponseRecorder to record the response
			rr := httptest.NewRecorder()

			// Call the handler
			handler := http.HandlerFunc(RawStatsHandler)
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
