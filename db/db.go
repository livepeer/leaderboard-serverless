package db

import (
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/livepeer/leaderboard-serverless/common"
	"github.com/livepeer/leaderboard-serverless/models"
	"github.com/livepeer/leaderboard-serverless/postgres"
)

var Store DB

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

func Start(connectionUrl string) error {
	if connectionUrl != "" {
		db, err := postgres.Start(connectionUrl)
		// if not error, cache the handle to the database and set up signal handler for graceful shutdown
		if err == nil {
			Store = db
			// Start the scheduler that reads catalyst json
			// and updates the regions
			go StartUpdateRegionsScheduler()

			go handleShutdown()
			common.Logger.Debug("Database connection pool startup completed.")
		}
		return err
	}
	return errors.New("no database specified")
}

func CacheDB() error {
	if Store != nil {
		return nil
	}

	//make sure POSTGRES environment variable is set
	postgresqlUrl := os.Getenv("POSTGRES")
	if postgresqlUrl == "" {
		common.Logger.Fatal("POSTGRES environment variable is not set")
	}

	return Start(postgresqlUrl)
}

func handleShutdown() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	if Store != nil {
		Store.Close()
	}
	os.Exit(0)
}
