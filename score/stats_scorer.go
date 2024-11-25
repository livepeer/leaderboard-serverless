package score

import (
	"math"

	"github.com/livepeer/leaderboard-serverless/common"
	"github.com/livepeer/leaderboard-serverless/models"
)

/** AI METRICS SCORING **/

// desiredScoreAtMedian is the k coefficient for exponential decay
// Set k such that the score equals the desired score
// at the median RTT.
// Median was chosen to be more resilient to outliers.
const desiredScoreAtMedian float64 = 0.8

// Weights for success rate and RTT score
// in the final score calculation
const weightSuccess float64 = 0.65
const weightRTT float64 = 0.35

func calculateTotalScore(stats *models.AggregatedStats, jobType string) float64 {
	// if not stats or success rate is 0, return 0
	if stats == nil || stats.SuccessRate == 0 {
		return 0
	}
	if jobType == models.AI.String() {
		return ((weightSuccess * stats.SuccessRate) + (weightRTT * stats.RoundTripScore))
	}

	return stats.SuccessRate * stats.RoundTripScore
}

func CreateAggregatedStats(aggrStatsResults *models.AggregatedStatsResults) map[string]map[string]*models.AggregatedStats {
	results := make(map[string]map[string]*models.AggregatedStats)
	common.Logger.Debug("Creating aggregated stats for %d stats", len(aggrStatsResults.Stats))

	for _, stat := range aggrStatsResults.Stats {
		_, ok := results[stat.Orchestrator]
		if !ok {
			results[stat.Orchestrator] = make(map[string]*models.AggregatedStats)
		}
		normalizedRTTScore := calculateRTTScore(stat, aggrStatsResults.MedianRTT)
		aggrStats := &models.AggregatedStats{
			ID:             stat.Orchestrator,
			SuccessRate:    stat.SuccessRate,
			RoundTripScore: normalizedRTTScore,
		}
		aggrStats.TotalScore = calculateTotalScore(aggrStats, stat.JobType())
		results[stat.Orchestrator][stat.Region] = aggrStats

		common.Logger.Trace("Stat object added with Orchestrator: %v, Region: %v, SuccessRate: %v, RoundTripTime: %v, SegDuration: %v, TotalScore: %v",
			stat.Orchestrator, stat.Region, stat.SuccessRate, stat.RoundTripTime, stat.SegDuration, aggrStats.TotalScore)
	}
	common.Logger.Trace("Compiled aggregated stats: %v", results)
	return results
}

// Calculate the RTT score for a given stat
func calculateRTTScore(stat *models.Stats, medianRtt float64) float64 {

	if stat.JobType() == models.AI.String() {
		return normalizeAndCalcRTTScore(medianRtt, stat)
	}
	return normalizeLatencyScore(calculateLatencyScore(stat))
}

// Calculate the latency score for a given stat.  This function
// applies only to transcoding jobs.  AI Jobs are scored differently.
func calculateLatencyScore(stat *models.Stats) float64 {
	common.Logger.Trace("Calculating latency score for stat: %v and jobType: %v", stat, stat.JobType())
	if stat == nil {
		return 0
	}

	segDuration := stat.SegDuration
	latency := stat.RoundTripTime
	if latency == 0 {
		return 0
	}
	//check for stat type (ai or transcoding) and calculate latency score accordingly
	if stat.JobType() == models.AI.String() {
		// issue a warning as this function is not intended to be called for AI jobs
		common.Logger.Warn("calculateLatencyScore called for AI job.  This function is intended for transcoding jobs.")
		return latency
	} else {
		return segDuration / latency
	}
}

// Normalize the latency score to a value between 0 and 1
func normalizeLatencyScore(score float64) float64 {
	return 1 - math.Pow(math.E, -score)
}

// CalculateScores calculates the final scores for the given stats
// using a combination of success rate and RTT scores through exponential decay E(x)=e^−kx
func normalizeAndCalcRTTScore(medianRTT float64, stat *models.Stats) float64 {

	common.Logger.Trace("Calculating RTT score for stat: %v and jobType: %v", stat, stat.JobType())

	// Calculate k based on desired score at median RTT
	k := -math.Log(desiredScoreAtMedian) / medianRTT

	// Compute Exponential Decay Score for RTTs
	expDecayScore := math.Exp(-k * stat.RoundTripTime)

	common.Logger.Trace("Mean RTT: %v, RoundTripTime: %v, expDecayScore: %v",
		medianRTT, stat.RoundTripTime, expDecayScore)

	return expDecayScore
}
