package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// Regions (collections)
var Regions = []string{"MDW", "FRA", "SIN", "NYC", "LAX", "LON", "PRG"}

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
