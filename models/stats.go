package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type AggregatedStatsResults struct {
	Stats     []*Stats
	MedianRTT float64
}

func (a *AggregatedStatsResults) HasResults() bool {
	return a != nil && len(a.Stats) > 0
}

// AggregatedStats are the aggregated stats for an orchestrator
type AggregatedStats struct {
	ID             string  `json:"-" bson:"_id,omitempty"`
	SuccessRate    float64 `bson:"success_rate" json:"success_rate"`
	RoundTripScore float64 `bson:"round_trip_score" json:"round_trip_score"`
	TotalScore     float64 `bson:"score" json:"score"`
}

// Score is a sample of a single score for an orchestrator
type Score struct {
	Region       string  `json:"region" bson:"region"`
	Orchestrator string  `json:"orchestrator" bson:"orchestrator"`
	Value        float64 `json:"value" bson:"value"`
	Model        string  `json:"model" bson:"model"`
	Pipeline     string  `json:"pipeline" bson:"pipeline"`
}

// JobType custom types to reference either Transcoding or AI jobs
type JobType int

const (
	Unknown JobType = iota
	Transcoding
	AI
)

// String method to convert the enum to a lower case string
func (jt JobType) String() string {
	return [...]string{"unknown", "transcoding", "ai"}[jt]
}

func JobTypeFromString(s string) (JobType, error) {
	switch s {
	case "transcoding":
		return Transcoding, nil
	case "ai":
		return AI, nil
	default:
		return Unknown, errors.New("invalid JobType string")
	}
}

// Stats are the raw stats per test stream
type Stats struct {
	Region        string  `json:"region" bson:"-"`
	Orchestrator  string  `json:"orchestrator" bson:"orchestrator"`
	SuccessRate   float64 `json:"success_rate" bson:"success_rate"`
	RoundTripTime float64 `json:"round_trip_time" bson:"round_trip_time"`
	Errors        []Error `json:"errors" bson:"errors"`
	Timestamp     int64   `json:"timestamp" bson:"timestamp"`

	// Transcoding stats fields
	SegDuration      float64 `json:"seg_duration,omitempty" bson:"seg_duration,omitempty"`
	SegmentsSent     int     `json:"segments_sent,omitempty" bson:"segments_sent,omitempty"`
	SegmentsReceived int     `json:"segments_received,omitempty" bson:"segments_received,omitempty"`
	UploadTime       float64 `json:"upload_time,omitempty" bson:"upload_time,omitempty"`
	DownloadTime     float64 `json:"download_time,omitempty" bson:"download_time,omitempty"`
	TranscodeTime    float64 `json:"transcode_time,omitempty" bson:"transcode_time,omitempty"`

	// AI stats fields
	Model           string `json:"model,omitempty" bson:"model,omitempty"`
	ModelIsWarm     bool   `json:"model_is_warm,omitempty" bson:"model_is_warm,omitempty"`
	Pipeline        string `json:"pipeline,omitempty" bson:"pipeline,omitempty"`
	InputParameters string `json:"input_parameters,omitempty" bson:"input_parameters,omitempty"`
	ResponsePayload string `json:"response_payload,omitempty" bson:"response_payload,omitempty"`
}

type Error struct {
	ErrorCode string `json:"error_code" bson:"error_code"`
	Message   string `json:"message,omitempty" bson:"message,omitempty"`
	Count     int    `json:"count" bson:"count"`
}

type Region struct {
	Name        string `bson:"id" json:"id"`
	DisplayName string `bson:"name" json:"name"`
	Type        string `bson:"type" json:"type"`
}

type Pipeline struct {
	Name    string   `bson:"id" json:"id"`
	Models  []string `bson:"models" json:"models"`
	Regions []string `bson:"regions" json:"regions"`
}

type StatsQuery struct {
	Orchestrator string
	Region       string
	Model        string
	Pipeline     string
	Since        time.Time
	Until        time.Time
	JobType      JobType
	SortFields   []StatsQuerySortField
	Limit        int
}

func (s *Stats) JobType() string {
	if s.Model != "" && s.Pipeline != "" {
		return AI.String()
	}
	return Transcoding.String()
}

// StatsQuerySortField defines a field and its sort order.
type StatsQuerySortField struct {
	Field string
	Order SortOrder
}

// SortOrder represents the order direction for sorting.
type SortOrder int

const (
	SortOrderAsc SortOrder = iota
	SortOrderDesc
)

// String returns the string representation of the SortOrder.
func (s SortOrder) String() string {
	switch s {
	case SortOrderAsc:
		return "ASC"
	case SortOrderDesc:
		return "DESC"
	default:
		return "UNKNOWN"
	}
}

// String returns the string representation of the SortField, e.g., "success_rate DESC".
func (s StatsQuerySortField) String() string {
	return fmt.Sprintf("%s %s", s.Field, s.Order.String())
}

// NewSortField creates a new StatsQuerySortField with the given field and order.
func NewSortField(field string, order SortOrder) StatsQuerySortField {
	return StatsQuerySortField{
		Field: field,
		Order: order,
	}
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

// COMMON ERRORS
var ErrMissingPipeline = errors.New("pipeline required")
var ErrMissingModel = errors.New("model required")
