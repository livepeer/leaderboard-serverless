package db

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/livepeer/leaderboard-serverless/common"
	"github.com/livepeer/leaderboard-serverless/models"
)

// Regions (collections)
var (
	Regions           []string
	RegionsLastUpdate time.Time
	mu                sync.RWMutex
)

func StartUpdateRegionsScheduler() {
	catalystJSONURL := common.EnvOrDefault("CATALYSTS_REGION_URL", "").(string)
	if catalystJSONURL == "" {
		common.Logger.Error("No catalyst JSON URL provided.  Will not run updater schedule.")
		return
	}

	interval := common.EnvOrDefault("CATALYSTS_REGION_UPDATE_INTERVAL", 60).(int)
	ticker := time.NewTicker(time.Duration(interval) * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		UpdateCatalystRegions()
	}
}

// UpdateCatalystRegions updates the regions in the database
// and returns the number of regions inserted and processed
func UpdateCatalystRegions() (int, int) {

	mu.Lock()
	defer mu.Unlock()

	totalProcessed := 0

	regions, err := GetCatalystRegions()
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
		common.Logger.Warn("Not all regions were inserted.  Only %d of %d were inserted", totalInserted, len(regions))
	}
	RegionsLastUpdate = time.Now()
	common.Logger.Info("%d regions have been updated.", totalInserted)
	return totalInserted, totalProcessed
}

func GetCatalystRegions() ([]*models.Region, error) {

	catalystJSONURL := common.EnvOrDefault("CATALYSTS_JSON", "https://livepeer.github.io/livepeer-infra/catalysts.json").(string)

	resp, err := http.Get(catalystJSONURL)
	if err != nil {
		return nil, fmt.Errorf("can't fetch the %s: %s", catalystJSONURL, err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("can't read the %s: %s", catalystJSONURL, err)
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
		return nil, fmt.Errorf("can't parse the %s: %s", catalystJSONURL, err)
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
