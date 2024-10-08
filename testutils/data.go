package testutils

import (
	"fmt"
	"strconv"
	"time"

	"github.com/livepeer/leaderboard-serverless/models"
)

// IMPORTANT: DO NOT CHANGE DATA IN THIS FILE WIHTOUT RUNNING ALL TESTS AFTERWARDS!
var testOrchId = "0x5c0e79538f4d17a668568c4031e4a1488d71df1a"
var testRegion = "MDW"
var testPipeline = "text-to-image"
var testModel = "ByteDance/SDXL-Lightning"
var testTransStats = &models.Stats{
	Region:           testRegion,
	Orchestrator:     testOrchId,
	UploadTime:       0.16989430653333334,
	SegDuration:      2.0804666666666662,
	SuccessRate:      1,
	DownloadTime:     0.11656354278571428,
	SegmentsSent:     15,
	TranscodeTime:    0.3484082023238095,
	RoundTripTime:    0.6348660516428571,
	SegmentsReceived: 30,
}
var testAIStats = &models.Stats{
	Model: testModel,
	Errors: []models.Error{
		{
			Count:     1,
			ErrorCode: "HTTP-STATUS-503",
		},
	},
	Region:          testRegion,
	Pipeline:        testPipeline,
	Orchestrator:    testOrchId,
	SuccessRate:     0,
	RoundTripTime:   2.21572637,
	InputParameters: fmt.Sprintf(`{"guidance_scale":2,"height":512,"pipeline":"%s","model_id":"%s","num_images_per_prompt":1,"num_inference_steps":6,"prompt":"a bear","safety_check":false,"width":512}`, testPipeline, testModel),
	ResponsePayload: `{"error":{"message":"Service Unavailable Error"}}`,
}

var bestAIStatsLAX = &models.Stats{
	Model:           testModel,
	Errors:          nil,
	Region:          "LAX",
	Pipeline:        testPipeline,
	Orchestrator:    testOrchId,
	SuccessRate:     1,
	RoundTripTime:   0.1,
	InputParameters: "",
	ResponsePayload: "",
}

var testPipelineJson = fmt.Sprintf(`{
  "pipelines": [
    {
      "id": "%s",
      "models": [
        "%s"
      ],
      "regions": [
        "%s"
      ]
    }
  ]
}`, testPipeline, testModel, testRegion)

var testNewRegion = &models.Region{
	Name:        "NPL",
	DisplayName: "Northpole",
	Type:        models.Transcoding.String(),
}

func GetNewRegion() *models.Region {
	return testNewRegion
}

func GetOrchestratorID() string {
	return testOrchId
}

func GetPipeline() string {
	return testPipeline
}

func GetPipelineJson() string {
	return testPipelineJson
}

func GetModel() string {
	return testModel
}

func GetTranscodingStats() models.Stats {
	return *testTransStats
}

func GetAIStats() models.Stats {
	return *testAIStats
}

func GetBestAIStats() models.Stats {
	return *bestAIStatsLAX
}

func GetUnixTimeInFiveSec() time.Time {
	return time.Now().Add(5 * time.Second)
}

func GetUnixTimeInFiveSecStr() string {
	return ConvertUnixTimeToStrimg(GetUnixTimeInFiveSec())
}

func GetUnixTimeMinus24Hr() time.Time {
	return time.Now().Add(-24 * time.Hour)
}

func GetUnixTimeMinus24HrStr() string {
	return ConvertUnixTimeToStrimg(GetUnixTimeMinus24Hr())
}

func GetUnixTimeMinusTenSec() time.Time {
	return time.Now().Add(-10 * time.Second)
}

func GetUnixTimeMinusTenSecStr() string {
	return ConvertUnixTimeToStrimg(GetUnixTimeMinusTenSec())
}

func ConvertUnixTimeToStrimg(t time.Time) string {
	unixNano := t.UnixNano()
	unixFloat := float64(unixNano) / 1e9 // Convert nanoseconds to seconds with fractions
	return strconv.FormatFloat(unixFloat, 'f', -1, 64)
}
