package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/livepeer/leaderboard-serverless/common"
	"github.com/livepeer/leaderboard-serverless/db"
	"github.com/livepeer/leaderboard-serverless/models"
	"github.com/livepeer/leaderboard-serverless/score"
	"github.com/livepeer/leaderboard-serverless/testutils"
)

func TestMain(m *testing.M) {
	testutils.TestMain(m)
}

func TestAggregatedStatsHandler(t *testing.T) {

	// insert some data into the test database
	testStats := testutils.GetTranscodingStats()
	aiTestStats := testutils.GetBestAIStats()

	//create the aggregated stats from the test data and compare
	testTranscodingStatsArray := []*models.Stats{&testStats}
	testTranscodingStatsResults := &models.AggregatedStatsResults{Stats: testTranscodingStatsArray}
	testTranscodingAggregatedStats := score.CreateAggregatedStats(testTranscodingStatsResults)

	//create the AI aggregated stats from the test data and compare
	testAIStatsArray := []*models.Stats{&aiTestStats}
	testAIStatsResults := &models.AggregatedStatsResults{Stats: testAIStatsArray, MedianRTT: 0.1}
	testAIAggregatedStats := score.CreateAggregatedStats(testAIStatsResults)

	// create an array with testStats and aiTestStats
	allStatsArray := []*models.Stats{&testStats, &aiTestStats}

	until := testutils.GetUnixTimeInFiveSecStr()
	since := testutils.GetUnixTimeMinus24HrStr()
	tests := []struct {
		name              string
		queryParams       string
		expectedStatus    int
		expectedAggrStats map[string]map[string]*models.AggregatedStats
		expectedError     string
	}{
		{
			name:              "Test with valid transcoding input and query params",
			queryParams:       "until=" + until + "&since=" + since,
			expectedStatus:    http.StatusOK,
			expectedAggrStats: testTranscodingAggregatedStats,
		},
		{
			name:              "Test with valid AI input and query params",
			queryParams:       "pipeline=" + url.QueryEscape(aiTestStats.Pipeline) + "&model=" + url.QueryEscape(aiTestStats.Model) + "&until=" + until + "&since=" + since,
			expectedStatus:    http.StatusOK,
			expectedAggrStats: testAIAggregatedStats,
		},
		{
			name:              "Test with invalid input and query params",
			queryParams:       "until=" + strconv.FormatInt(time.Now().AddDate(-1, 0, 0).Unix(), 10), // One year ago
			expectedStatus:    http.StatusOK,
			expectedAggrStats: map[string]map[string]*models.AggregatedStats{}, //expect nothing
		},
		{
			name:           "Test with invalid query params - no pipeline",
			queryParams:    "model=testModel&until=" + strconv.FormatInt(time.Now().AddDate(-1, 0, 0).Unix(), 10), // One year ago
			expectedStatus: http.StatusBadRequest,
			expectedError:  `{"error":"pipeline required"}`,
		},
		{
			name:           "Test with invalid query params - no model",
			queryParams:    "pipeline=Text%20to%20image&until=" + strconv.FormatInt(time.Now().AddDate(-1, 0, 0).Unix(), 10), // One year ago
			expectedStatus: http.StatusBadRequest,
			expectedError:  `{"error":"model required"}`,
		},
	}
	runTests(t, tests, allStatsArray)
}

func runTests(t *testing.T, tests []struct {
	name              string
	queryParams       string
	expectedStatus    int
	expectedAggrStats map[string]map[string]*models.AggregatedStats
	expectedError     string
}, allStatsArray []*models.Stats) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			common.Logger.Info("Running test: %v", tt.name)
			testutils.NewDB(t)
			for _, statsToInsert := range allStatsArray {
				if err := db.Store.InsertStats(statsToInsert); err != nil {
					t.Fatalf("Unexpected error when inserting stats: %v", err)
				}
			}

			// Create a new HTTP request with query parameters
			common.Logger.Info("Query params: %v", tt.queryParams)
			req, err := http.NewRequest("GET", "/aggregated-stats?"+tt.queryParams, bytes.NewBuffer([]byte{}))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			// Create a ResponseRecorder to record the response
			rr := httptest.NewRecorder()

			// Call the handler
			handler := http.HandlerFunc(AggregatedStatsHandler)
			handler.ServeHTTP(rr, req)

			// Check the status code
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("Handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			expectedBodyString := tt.expectedError
			if tt.expectedError == "" {
				// Convert the response body to the expected type
				var responseBody map[string]map[string]*models.AggregatedStats
				err = json.Unmarshal(rr.Body.Bytes(), &responseBody)
				if err != nil {
					t.Fatalf("Failed to unmarshal response body [%v]\n Body: %s", err, rr.Body.String())
				}
				common.Logger.Info("Response body: %v", rr.Body.String())
				// Convert the expected body to JSON string for comparison
				expectedBodyBytes, err := json.Marshal(tt.expectedAggrStats)
				if err != nil {
					t.Fatalf("Failed to marshal expected body: %v", err)
				}
				expectedBodyString = string(expectedBodyBytes)
			}

			// Compare the response body with the expected body making sure
			// to ignore carriage returns
			var responseBody string
			if rr.Body != nil {
				responseBody = string(bytes.ReplaceAll(bytes.ReplaceAll(rr.Body.Bytes(), []byte("\r"), []byte("")), []byte("\n"), []byte("")))
			}

			if responseBody != expectedBodyString {
				t.Errorf("Handler returned unexpected body: got '%v' want '%v'", rr.Body.String(), expectedBodyString)
			}
		})
	}
}
