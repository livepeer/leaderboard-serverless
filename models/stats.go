package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Regions (collections)
var Regions = []string{"MDW", "FRA", "SIN"}

// AggregatedStats are the aggregated stats for an orchestrator
type AggregatedStats struct {
	ID          string  `json:"-" bson:"_id,omitempty"`
	Score       float64 `bson:"score" json:"score"`
	SuccessRate float64 `bson:"success_rate" json:"success_rate"`
}

// Stats are the raw stats per test stream
type Stats struct {
	ID                primitive.ObjectID `json:"-" bson:"_id,omitempty"`
	Region            string             `json:"region" bson:"-"`
	Orchestrator      string             `json:"orchestrator" bson:"orchestrator"`
	SegmentsSent      int                `json:"segments_sent" bson:"segments_sent"`
	SegmentsReceived  int                `json:"segments_received" bson:"segments_received"`
	SuccessRate       float64            `json:"success_rate" bson:"success_rate"`
	AvgSegDuration    float64            `json:"avg_seg_duration" bson:"avg_seg_duration"`
	AvgUploadTime     float64            `json:"avg_upload_time" bson:"avg_upload_time"`
	AvgUploadScore    float64            `json:"avg_upload_score" bson:"avg_upload_score"`
	AvgDownloadTime   float64            `json:"avg_download_time" bson:"avg_download_time"`
	AvgDownloadScore  float64            `json:"avg_download_score" bson:"avg_download_score"`
	AvgTranscodeTime  float64            `json:"avg_transcode_time" bson:"avg_transcode_time"`
	AvgTranscodeScore float64            `json:"avg_transcode_score" bson:"avg_transcode_score"`
	RoundTripScore    float64            `json:"round_trip_score" bson:"round_trip_score"`
	Errors            []Error            `json:"errors" bson:"errors"`
	Timestamp         int64              `json:"timestamp" bson:"timestamp"`
}

type Error struct {
	ErrorCode string `json:"error_code" bson:"error_code"`
	Count     int    `json:"count" bson:"count"`
}

// Value implements the driver.Valuer interface from SQL
// This function simply results the JSON-encoded representation of teh struct
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
