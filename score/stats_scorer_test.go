package score

import (
	"testing"

	"github.com/livepeer/leaderboard-serverless/db"
	"github.com/livepeer/leaderboard-serverless/models"
	"github.com/livepeer/leaderboard-serverless/testutils"
)

func TestMain(m *testing.M) {
	testutils.TestMain(m)
}

func TestRTTScoreCalc(t *testing.T) {

	aiTestStats := testutils.GetBestAIStats()

	// create four copies of the same stats with different RTT and SuccessRate
	// to test the RTT score calculation
	aiTestStats1 := aiTestStats
	aiTestStats1.Orchestrator = "testSLOWOrchestrator1"
	aiTestStats1.RoundTripTime = 8.0
	aiTestStats1.SuccessRate = 1.0

	aiTestStats2 := aiTestStats
	aiTestStats2.Orchestrator = "testAvgOrchestrator2"
	aiTestStats2.RoundTripTime = 2.0
	aiTestStats2.SuccessRate = 1.0

	aiTestStats3 := aiTestStats
	aiTestStats3.Orchestrator = "testFastButLimitedSuccessOrchestrator3"
	aiTestStats3.RoundTripTime = 1.0
	aiTestStats3.SuccessRate = 0.75

	aiTestStats4 := aiTestStats
	aiTestStats4.Orchestrator = "testAvgBufLimitedSuccessOrchestrator4"
	aiTestStats4.RoundTripTime = 2.0
	aiTestStats4.SuccessRate = 0.75

	aiTestStats5 := aiTestStats
	aiTestStats5.Orchestrator = "testFastandGoodOrchestrator5"
	aiTestStats5.RoundTripTime = 0.5
	aiTestStats5.SuccessRate = 0.95

	// create an array with the four stats
	aiTestStatsArray := []*models.Stats{&aiTestStats1, &aiTestStats2, &aiTestStats3, &aiTestStats4, &aiTestStats5}

	testutils.NewDB(t)
	for _, statsToInsert := range aiTestStatsArray {
		if err := db.Store.InsertStats(statsToInsert); err != nil {
			t.Fatalf("Unexpected error when inserting stats: %v", err)
		}
	}

	statsQuery := &models.StatsQuery{
		Model:    testutils.GetModel(),
		Pipeline: testutils.GetPipeline(),
		Since:    testutils.GetUnixTimeMinus24Hr(),
		Until:    testutils.GetUnixTimeInFiveSec(),
	}
	medianRTT, err := db.Store.MedianRTT(statsQuery)
	if err != nil {
		t.Fatalf("Unexpected error when getting median RTT: %v", err)
	}

	testAIStatsResults := &models.AggregatedStatsResults{Stats: aiTestStatsArray, MedianRTT: medianRTT}
	testAIAggregatedStats := CreateAggregatedStats(testAIStatsResults)

	// loop through the aggregated stats and check the RTT score calculation
	// with a map of Orchestrator to expected RTT score and Total Score
	expectedScores := map[string][]float64{
		//ORCH: 																	{RTT Score,         Total Score}
		"testSLOWOrchestrator1":                  {0.699751727323698, 0.8949131045632943},
		"testAvgOrchestrator2":                   {0.9146101038546527, 0.9701135363491284},
		"testFastButLimitedSuccessOrchestrator3": {0.956352499790037, 0.822223374926513},
		"testAvgBufLimitedSuccessOrchestrator4":  {0.9146101038546527, 0.8076135363491285},
		"testFastandGoodOrchestrator5":           {0.9779327685429285, 0.9597764689900249},
	}

	for _, stats := range aiTestStatsArray {

		if testAIAggregatedStats[stats.Orchestrator][stats.Region].RoundTripScore != expectedScores[stats.Orchestrator][0] {
			t.Errorf("Handler returned unexpected RTT score: got %v want %v", testAIAggregatedStats[stats.Orchestrator][stats.Region].RoundTripScore, expectedScores[stats.Orchestrator][0])
		}

		// validate the scores cames out as expected
		if testAIAggregatedStats[stats.Orchestrator][stats.Region].TotalScore != expectedScores[stats.Orchestrator][1] {
			t.Errorf("Handler returned unexpected total score score: got %v want %v", testAIAggregatedStats[stats.Orchestrator][stats.Region].TotalScore, expectedScores[stats.Orchestrator][1])
		}

	}
}
