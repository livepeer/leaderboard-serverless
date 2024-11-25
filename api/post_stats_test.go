package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/livepeer/leaderboard-serverless/common"
	"github.com/livepeer/leaderboard-serverless/db"
	"github.com/livepeer/leaderboard-serverless/middleware/auth"
	"github.com/livepeer/leaderboard-serverless/models"
	"github.com/livepeer/leaderboard-serverless/testutils"
)

func TestPostStatsHandler(t *testing.T) {

	// insert some data into the test database
	testStats := testutils.GetTranscodingStats()
	aiTestStats := testutils.GetAIStats()

	tests := []struct {
		name           string
		requestBody    models.Stats
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Test with valid transcoding input and query params",
			requestBody:    testStats,
			expectedStatus: http.StatusOK,
			expectedBody:   "ok",
		},
		{
			name:           "Test with valid AI input and query params",
			requestBody:    aiTestStats,
			expectedStatus: http.StatusOK,
			expectedBody:   "ok",
		},
	}

	runPostTests(t, tests)
}

func runPostTests(t *testing.T, tests []struct {
	name           string
	requestBody    models.Stats
	expectedStatus int
	expectedBody   string
}) {
	os.Setenv("SECRET", "secret-key")
	defer os.Unsetenv("SECRET")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			common.Logger.Info("Running test: %v", tt.name)
			testutils.NewDB(t)
			statUnderTest := tt.requestBody
			// Create a request body
			body, err := json.Marshal(statUnderTest)
			if err != nil {
				t.Fatalf("Failed to marshal request body: %v", err)
			}

			authHeader := auth.EncryptHeader([]byte(body))

			// Create a new HTTP request with query parameters
			req, err := http.NewRequest("POST", "/post-stats", bytes.NewBuffer(body))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", authHeader)

			// Create a ResponseRecorder to record the response
			rr := httptest.NewRecorder()

			// Call the handler
			handler := http.HandlerFunc(PostStatsHandler)
			handler.ServeHTTP(rr, req)

			// Check the status code
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("Handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			// Compare the response body with the expected body
			if rr.Body.String() != tt.expectedBody {
				t.Errorf("Handler returned unexpected body: got %v want %v", rr.Body.String(), tt.expectedBody)
			}

			common.Logger.Info("Validating that the request stats object was stored in the database. Expected: %v", statUnderTest)
			//get the statsRetrievedFromDb object from the database
			statsRetrievedFromDb, err := db.Store.RawStats(&models.StatsQuery{
				Orchestrator: statUnderTest.Orchestrator,
				Since:        testutils.GetUnixTimeMinusTenSec(),
				Until:        testutils.GetUnixTimeInFiveSec(),

				Pipeline: statUnderTest.Pipeline,
				Model:    statUnderTest.Model,
			})
			if err != nil {
				t.Fatalf("Failed to get stats from database: %v", err)
			}

			if len(statsRetrievedFromDb) == 0 {
				t.Errorf("No stats found in the database.  We expected to find the stats object that was posted")
			} else if cmp.Equal(statsRetrievedFromDb[0], statUnderTest) {
				t.Errorf("Handler returned unexpected stats: got %v want %v", statsRetrievedFromDb[0], statUnderTest)
			}
		})
	}
}
