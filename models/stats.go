package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// Regions (collections)
var (
	Regions           []string
	RegionsLastUpdate time.Time
)

// AggregatedStats are the aggregated stats for an orchestrator
type AggregatedStats struct {
	ID             string  `json:"-" bson:"_id,omitempty"`
	SuccessRate    float64 `bson:"success_rate" json:"success_rate"`
	RoundTripScore float64 `bson:"round_trip_score" json:"round_trip_score"`
	TotalScore     float64 `bson:"score" json:"score"`
}

// Stats are the raw stats per test stream
type Stats struct {
	ID               string  `json:"-" bson:"_id,omitempty"`
	Region           string  `json:"region" bson:"-"`
	Orchestrator     string  `json:"orchestrator" bson:"orchestrator"`
	SegmentsSent     int     `json:"segments_sent" bson:"segments_sent"`
	SegmentsReceived int     `json:"segments_received" bson:"segments_received"`
	SuccessRate      float64 `json:"success_rate" bson:"success_rate"`
	SegDuration      float64 `json:"seg_duration" bson:"seg_duration"`
	UploadTime       float64 `json:"upload_time" bson:"upload_time"`
	DownloadTime     float64 `json:"download_time" bson:"download_time"`
	TranscodeTime    float64 `json:"transcode_time" bson:"transcode_time"`
	RoundTripTime    float64 `json:"round_trip_time" bson:"round_trip_time"`
	Errors           []Error `json:"errors" bson:"errors"`
	Timestamp        int64   `json:"timestamp" bson:"timestamp"`
}

type Error struct {
	ErrorCode string `json:"error_code" bson:"error_code"`
	Count     int    `json:"count" bson:"count"`
}

// Value implements the driver.Valuer interface from SQL
// This function simply results the JSON-encoded representation of the struct
func (s Stats) Value() (driver.Value, error) {
	return json.Marshal(s)
}

// Scan implements the sql.Scanner interface. This method simply decodes a JSON represenation of the struct
func (s *Stats) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &s)
}

func GetRegions() []string {
	if !RegionsLastUpdate.IsZero() && time.Since(RegionsLastUpdate).Seconds() < 60 {
		// return cached list of regions
		return Regions
	}

	catalystJSONURL, exists := os.LookupEnv("CATALYSTS_JSON")
	if !exists {
		catalystJSONURL = "https://livepeer.github.io/livepeer-infra/catalysts.json"
	}

	resp, err := http.Get(catalystJSONURL)
	if err != nil {
		log.Printf("Can't fetch the %s: %s", catalystJSONURL, err)
		return Regions // return previosuly cached list of regions
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Can't read the %s: %s", catalystJSONURL, err)
		return Regions // return previosuly cached list of regions
	}

	type catalystEnvData struct {
		Counts    map[string]int `json:"counts"`
		Urls      []string       `json:"urls"`
		GlobalUrl string         `json:"global_url"`
	}
	var catalystData map[string]catalystEnvData
	err = json.Unmarshal(body, &catalystData)
	if err != nil {
		log.Printf("Can't parse the %s: %s", catalystJSONURL, err)
		return Regions // return previosuly cached list of regions
	}
	Regions = []string{}
	for region := range catalystData["prod"].Counts {
		Regions = append(Regions, strings.ToUpper(region))
	}
	RegionsLastUpdate = time.Now()
	return Regions
}
