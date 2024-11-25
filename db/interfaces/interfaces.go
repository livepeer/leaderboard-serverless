package interfaces

import "github.com/livepeer/leaderboard-serverless/models"

type DB interface {
	InsertStats(stats *models.Stats) error
	AggregatedStats(query *models.StatsQuery) (*models.AggregatedStatsResults, error)
	MedianRTT(query *models.StatsQuery) (float64, error)
	BestAIRegion(orchestratorId string) (*models.Stats, error)
	RawStats(query *models.StatsQuery) ([]*models.Stats, error)
	Regions() ([]*models.Region, error)
	InsertRegions(regions []*models.Region) (int, int)
	Pipelines(query *models.StatsQuery) ([]*models.Pipeline, error)
	Close()
}

type DBManager interface {
	UpdateRegions() (int, int)
}
