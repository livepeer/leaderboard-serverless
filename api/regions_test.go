package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/livepeer/leaderboard-serverless/common"
	"github.com/livepeer/leaderboard-serverless/db"
	"github.com/livepeer/leaderboard-serverless/models"
	"github.com/livepeer/leaderboard-serverless/testutils"
)

func TestRegionsHandler(t *testing.T) {
	testutils.NewDB(t)

	type testCase struct {
		name           string
		method         string
		since          string
		until          string
		expectedStatus int
		getRegion      func() *models.Region
	}

	testCases := []testCase{
		{
			name:           "Valid GET request with no region",
			method:         "GET",
			since:          testutils.GetUnixTimeInFiveSecStr(),
			until:          testutils.GetUnixTimeInFiveSecStr(),
			expectedStatus: http.StatusOK,
			getRegion:      nil,
		},
		{
			name:           "Valid GET request with regions",
			method:         "GET",
			since:          "",
			until:          "",
			expectedStatus: http.StatusMethodNotAllowed,
			getRegion:      testutils.GetNewRegion,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			common.Logger.Debug("Running test case: %v", tc.name)

			// if we have data, insert it into the database
			if tc.getRegion != nil {
				inserted, processed := db.Store.InsertRegions([]*models.Region{tc.getRegion()})
				if inserted != processed {
					t.Fatalf("Failed to insert regions into the database: %v", inserted)
				}
			}

			// Create a new HTTP request with query parameters
			req, err := http.NewRequest("GET", fmt.Sprintf("/regions?since=%s&until=%s", tc.since, tc.until), bytes.NewBuffer([]byte{}))
			if err != nil {
				t.Fatalf("Failed to create request to retrieve pipelines: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			// Create a ResponseRecorder to record the response
			rr := httptest.NewRecorder()

			// Call the handler
			handler := http.HandlerFunc(RegionsHandler)
			handler.ServeHTTP(rr, req)

			// Check the status code
			if status := rr.Code; status != 200 {
				t.Errorf("Handler returned wrong status code: got %v want %v", status, 200)
			}

			common.Logger.Info("Response body: %v", rr.Body.String())
			// convert response into regions array
			type RegionsResponse struct {
				Regions []models.Region `json:"regions"`
			}
			var regionsResponse RegionsResponse
			err = json.Unmarshal(rr.Body.Bytes(), &regionsResponse)
			if err != nil {
				t.Fatalf("Failed to unmarshal response body [%v]\n Body: %s", err, rr.Body.String())
			}

			// make sure there is at least one region
			if len(regionsResponse.Regions) == 0 {
				t.Errorf("No regions found in response")
			}

			//make sure we found the region we inserted
			if tc.getRegion != nil {
				found := false
				for _, region := range regionsResponse.Regions {
					if region.Name == tc.getRegion().Name {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Region not found in response body")
				}
			}
		})
	}
}
