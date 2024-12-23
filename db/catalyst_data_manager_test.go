package db_test

import (
	"os"
	"testing"

	"github.com/livepeer/leaderboard-serverless/common"
	"github.com/livepeer/leaderboard-serverless/db"
	"github.com/livepeer/leaderboard-serverless/testutils"
)

func TestMain(m *testing.M) {
	testutils.TestMain(m)
}

func TestUpdateRegions(t *testing.T) {

	os.Setenv("CATALYST_REGION_URL", "https://livepeer.github.io/livepeer-infra/catalysts.json")

	testutils.NewDB(t)

	existingRegionsInDB, err := db.Store.Regions()
	if err != nil {
		t.Errorf("Error getting orignial regions for validation: %v", err)
	}

	catalystDataManager := db.NewCatalystDataManager()
	regionsFound, err := catalystDataManager.GetCatalystRegions()
	if err != nil {
		t.Errorf("Error getting regions: %v", err)
	}
	totalInserted, totalProcessed := catalystDataManager.UpdateRegions()

	if totalProcessed == 0 {
		t.Errorf("No regions processed")
	}

	if totalProcessed != len(regionsFound) {
		t.Errorf("Regions processed do not match regions found")
	}

	// if new regions were inserted, we should have more regions in the db
	newRegionsInDB, err := db.Store.Regions()
	common.Logger.Debug("Original # of regions now in DB: %v", len(existingRegionsInDB))
	common.Logger.Debug("Total # of regions now in DB: %v", len(newRegionsInDB))
	if err != nil {
		t.Errorf("Error getting regions for validation of inserts: %v", err)
	}
	if len(newRegionsInDB) != len(existingRegionsInDB)+totalInserted {
		t.Errorf("Regions not inserted as expected. Expected %d, got %d", len(existingRegionsInDB)+totalInserted, len(newRegionsInDB))
	}

}
