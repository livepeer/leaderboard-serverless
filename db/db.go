package db

import (
	"errors"
	"os"

	"github.com/livepeer/leaderboard-serverless/models"
	"github.com/livepeer/leaderboard-serverless/mongo"
	"github.com/livepeer/leaderboard-serverless/postgres"
)

var Store DB

type DB interface {
	InsertStats(stats *models.Stats) error
	AggregatedStats(orch, region, since string) ([]*models.AggregatedStats, error)
	RawStats(orch, region, since string) ([]*models.Stats, error)
}

func Start() (DB, error) {
	if os.Getenv("POSTGRES") != "" {
		return postgres.Start()
	}

	if os.Getenv("MONGO") != "" {
		return mongo.Start()
	}

	return nil, errors.New("no database specified")
}

func CacheDB() error {
	if Store != nil {
		return nil
	}

	var err error
	Store, err = Start()

	return err
}
