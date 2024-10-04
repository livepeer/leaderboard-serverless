package handler

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/livepeer/leaderboard-serverless/common"
	"github.com/livepeer/leaderboard-serverless/db"
	"github.com/livepeer/leaderboard-serverless/models"
	"github.com/livepeer/leaderboard-serverless/testutils"
)

func TestPipelinesHandler(t *testing.T) {
	testutils.NewDB(t)

	type testCase struct {
		name           string
		method         string
		since          string
		until          string
		expectedStatus int
		expectedBody   string
		getStats       func() models.Stats
	}

	testCases := []testCase{
		{
			name:           "Valid GET request with no pipeline",
			method:         "GET",
			since:          testutils.GetUnixTimeInFiveSecStr(),
			until:          testutils.GetUnixTimeInFiveSecStr(),
			expectedStatus: http.StatusOK,
			expectedBody:   `{"pipelines": []}`,
			getStats:       nil,
		},
		{
			name:           "Valid GET request with pipeline",
			method:         "GET",
			since:          "",
			until:          "",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   testutils.GetPipelineJson(),
			getStats:       testutils.GetAIStats,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			common.Logger.Debug("Running test case: %v", tc.name)

			// if we have data, insert it into the database
			if tc.getStats != nil {
				stats := tc.getStats()
				err := db.Store.InsertStats(&stats)
				if err != nil {
					t.Fatalf("Failed to insert stats into the database: %v", err)
				}
			}

			// Create a new HTTP request with query parameters
			req, err := http.NewRequest("GET", fmt.Sprintf("/pipelines?since=%s&until=%s", tc.since, tc.until), bytes.NewBuffer([]byte{}))
			if err != nil {
				t.Fatalf("Failed to create request to retrieve pipelines: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			// Create a ResponseRecorder to record the response
			rr := httptest.NewRecorder()

			// Call the handler
			handler := http.HandlerFunc(PipelinesHandler)
			handler.ServeHTTP(rr, req)

			// Check the status code
			if status := rr.Code; status != 200 {
				t.Errorf("Handler returned wrong status code: got %v want %v", status, 200)
			}

			//remove whitespace and carriage returns from tc.expectedBody
			tc.expectedBody = strings.ReplaceAll(tc.expectedBody, " ", "")
			tc.expectedBody = strings.ReplaceAll(tc.expectedBody, "\n", "")
			tc.expectedBody = strings.ReplaceAll(tc.expectedBody, "\t", "")

			// make sure we got matching response body
			if rr.Body.String() != tc.expectedBody {
				t.Errorf("Handler returned unexpected body: got %v want %v", rr.Body.String(), tc.expectedBody)
			}
		})
	}
}
