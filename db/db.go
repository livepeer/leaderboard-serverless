package db

import (
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/livepeer/leaderboard-serverless/common"
	"github.com/livepeer/leaderboard-serverless/db/cache"
	"github.com/livepeer/leaderboard-serverless/db/interfaces"
	"github.com/livepeer/leaderboard-serverless/postgres"
)

var Store interfaces.DB

func Start(connectionUrl string) error {
	if connectionUrl != "" {
		db, err := postgres.Start(connectionUrl, cache.NewCache(), NewCatalystDataManager())
		// if not error, cache the handle to the database and set up signal handler for graceful shutdown
		if err == nil {
			Store = db

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
