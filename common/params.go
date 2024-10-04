package common

import (
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/livepeer/leaderboard-serverless/models"
)

// ParseStatsQueryParams parses and defaults the 'since', 'until', and 'region' parameters from the URL query.
func ParseStatsQueryParams(r *http.Request) (*models.StatsQuery, error) {
	queryParams := r.URL.Query()

	orch := strings.ToLower(queryParams.Get("orchestrator"))
	pipeline := queryParams.Get("pipeline")
	model := queryParams.Get("model")

	// Parse 'Since' parameter
	since, err := parseSince(r)
	if err != nil {
		return nil, err
	}
	// Parse 'Until' parameter
	until, err := parseUntil(r)
	if err != nil {
		return nil, err
	}

	// Parse 'Region' parameter
	region := strings.ToUpper(queryParams.Get("region"))

	jobTypeStr := queryParams.Get("job_type")
	var finalJobType models.JobType
	if jobTypeStr != "" {
		jobTypeParsed, err := models.JobTypeFromString(jobTypeStr)
		if err != nil {
			Logger.Warn("Invalid job_type parameter: %v", queryParams.Get("job_type"))
		}
		finalJobType = jobTypeParsed
	}

	// if model OR pipeline is supplied, thrown an error since both are required
	if model != "" && pipeline == "" {
		return nil, models.ErrMissingPipeline
	}
	if pipeline != "" && model == "" {
		return nil, models.ErrMissingModel
	}

	Logger.Debug("Parsed query parameters: since=%v, until=%v, region=%v, orchestrator=%v, pipeline=%v, model=%v, jobType=%v", since, until, region, orch, pipeline, model, finalJobType)

	return &models.StatsQuery{
		Since:        since,
		Until:        until,
		Region:       region,
		Orchestrator: orch,
		Pipeline:     pipeline,
		Model:        model,
		JobType:      finalJobType,
	}, nil
}

func GetDefaultSince() time.Time {
	startTimeWindow := EnvOrDefault("START_TIME_WINDOW", 24).(int)
	return time.Now().Add(time.Duration(-startTimeWindow) * time.Hour).UTC()
}

func parseSince(r *http.Request) (time.Time, error) {
	queryParams := r.URL.Query()

	// Parse 'Since' parameter
	defaultSince := GetDefaultSince()
	sinceStr := queryParams.Get("since")

	if sinceStr == "" {
		// Default to a configured time window
		Logger.Debug("Looking back for stats starting from the time: %v", defaultSince)
		return defaultSince, nil
	} else {
		// Parse 'since' as float64 to include fractional seconds
		sinceFloat, err := strconv.ParseFloat(sinceStr, 64)
		if err != nil {
			return defaultSince, err
		}
		sec, frac := math.Modf(sinceFloat)
		return time.Unix(int64(sec), int64(frac*1e9)).UTC(), nil
	}
}

func parseUntil(r *http.Request) (time.Time, error) {
	queryParams := r.URL.Query()

	// Parse 'Until' parameter
	untilStr := queryParams.Get("until")
	if untilStr == "" {
		return time.Now().UTC(), nil
	} else {
		// Parse 'until' as float64 to include fractional seconds
		untilFloat, err := strconv.ParseFloat(untilStr, 64)
		if err != nil {
			return time.Now().UTC(), err
		}
		sec, frac := math.Modf(untilFloat)
		return time.Unix(int64(sec), int64(frac*1e9)).UTC(), nil
	}
}
