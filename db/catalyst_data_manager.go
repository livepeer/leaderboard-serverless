package db

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/livepeer/leaderboard-serverless/common"
	"github.com/livepeer/leaderboard-serverless/models"
)

type CatalystDataManager struct {
	isEnabled       bool
	catalystJSONURL string
}

// NewCatalystDataManager creates a new CatalystDataManager that will
// retrieve regions from the configured Catalyst JSON endpoint (CATALYST_REGION_URL)
// when the Regions() method is called (this is done within the postgres db implementation)
// and the cache is expired (see REGIONS_CACHE_TIMEOUT)
func NewCatalystDataManager() *CatalystDataManager {
	catalystJSONURL := os.Getenv("CATALYST_REGION_URL")
	var isEnabled bool
	if catalystJSONURL == "" {
		common.Logger.Warn("The variable 'CATALYST_REGION_URL' is currently not set in the environment.  As a result, this process will not run.")
		isEnabled = false
	} else {
		isEnabled = true
		common.Logger.Info("CatalystDataManager JSON URL is set to retrieve regions from: %s", catalystJSONURL)
	}
	return &CatalystDataManager{
		catalystJSONURL: catalystJSONURL,
		isEnabled:       isEnabled,
	}
}

// UpdateCatalystRegions updates the regions in the database
// and returns the number of regions inserted and processed
func (c *CatalystDataManager) UpdateRegions() (int, int) {
	if !c.isEnabled {
		common.Logger.Trace("CatalystDataManager is not enabled.  Exiting.")
		return 0, 0
	}

	totalProcessed := 0
	regions, err := c.GetCatalystRegions()
	if err != nil {
		common.Logger.Error("Error getting catalyst regions: %s", err)
		return 0, 0
	}

	if len(regions) == 0 {
		common.Logger.Error("No regions found in catalyst data")
		return 0, 0
	}

	// update the regions in the database
	totalInserted, totalProcessed := Store.InsertRegions(regions)
	if totalInserted != len(regions) {
		//some may not get inserted if they already exist
		//so we will only throw a warning for this case
		common.Logger.Debug("Not all regions were inserted.  Only %d of %d were inserted", totalInserted, len(regions))
	}
	common.Logger.Debug("%d regions have been updated.", totalInserted)
	return totalInserted, totalProcessed
}

// GetCatalystRegions gets the regions data from the configured Catalyst JSON endpoint
func (c *CatalystDataManager) GetCatalystRegions() ([]*models.Region, error) {
	if !c.isEnabled {
		common.Logger.Debug("CatalystDataManager is not enabled.  Exiting.")
		return nil, nil
	}

	resp, err := http.Get(c.catalystJSONURL)
	if err != nil {
		return nil, fmt.Errorf("can't fetch the %s: %s", c.catalystJSONURL, err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("can't read the %s: %s", c.catalystJSONURL, err)
	}

	common.Logger.Debug("Catalyst JSON: %s", string(body))

	type catalystEnvData struct {
		Region    map[string]string `json:"full_name"`
		Urls      []string          `json:"urls"`
		GlobalUrl string            `json:"global_url"`
	}
	var catalystData map[string]catalystEnvData
	err = json.Unmarshal(body, &catalystData)
	if err != nil {
		return nil, fmt.Errorf("can't parse the %s: %s", c.catalystJSONURL, err)
	}
	//create an array of models.Region
	var regions []*models.Region
	for region := range catalystData["prod"].Region {
		regionName := strings.ToUpper(region)
		regionDisplayName := catalystData["prod"].Region[region]
		common.Logger.Debug("Extracted region from JSON: %s with a display name of %s", regionName, regionDisplayName)
		regions = append(regions, &models.Region{
			Name:        regionName,
			DisplayName: regionDisplayName,
			Type:        models.Transcoding.String(),
		})
	}
	return regions, nil
}
