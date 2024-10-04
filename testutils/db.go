// file: testutils/db.go
package testutils

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/livepeer/leaderboard-serverless/common"
	"github.com/livepeer/leaderboard-serverless/db"
	"github.com/peterldowns/pgtestdb"
)

var postgres *embeddedpostgres.EmbeddedPostgres

// NewDB is a helper that creates a unique and isolated
// test database, fully migrated and ready for you to query.
func NewDB(t *testing.T) error {
	t.Helper()

	//running db connection url
	connectionURL := testDatabaseConfig.GetConnectionURL()
	parsedURL, err := url.Parse(connectionURL)
	if err != nil {
		return err
	}

	if parsedURL.User == nil {
		return fmt.Errorf("no user info in URL")
	}

	password, _ := parsedURL.User.Password()

	conf := pgtestdb.Config{
		DriverName: "postgres",
		User:       parsedURL.User.Username(),
		Password:   password,
		Host:       parsedURL.Hostname(),
		Port:       parsedURL.Port(),
		Options:    "sslmode=disable",
	}
	//we use a no-op migrator because the app will apply migrations at runtime
	var migrator pgtestdb.Migrator = pgtestdb.NoopMigrator{}
	pgTestDbConf := pgtestdb.Custom(t, conf, migrator)

	// IMPORTANT: need to set up the Cleanup after creating a db with pgtestdb
	// as pgtestdb registers a cleanup function that drops the database
	//and the database won't get dropped if it still has active connectoins
	//and t.Cleanup is called in a last added first called order
	t.Cleanup(func() {
		common.Logger.Debug("NewDB cleanup called")
		db.Store.Close()
	})

	// Parse the new connection string created from the template db
	// so we can get the unqiue databaes URL to be used for the application under test
	parsedURL, err = url.Parse(conf.URL())
	if err != nil {
		common.Logger.Info("Error parsing URL: %v", err)
		return err
	}
	parsedURL.Path = pgTestDbConf.Database
	common.Logger.Info("Running tests against database: %s", parsedURL.String())
	db.Start(parsedURL.String())

	// make sure the app has a handle to the backend store
	if err := db.CacheDB(); err != nil {
		common.Logger.Info("Expected no error when connecting to the database: %v", err)

		//the creation of the app db should not throw an error
		//so we must shut down the database server and fail hard
		ShutdownDatabase()

		//FailNow prevents further execution and ensures no pgtestdb resources are leaked
		t.FailNow()

		return err
	}
	common.Logger.Info("Successfully started the test database")
	return nil
}

var testDatabaseConfig embeddedpostgres.Config

// InitDB is a helper that returns an open connection to a unique and isolated test database
func InitDB(port uint32) error {
	if err := killActivePostgresInstances(); err != nil {
		common.Logger.Fatal("An error occurred while killing active postgres instances: %v", err)
	}
	testDatabaseConfig = embeddedpostgres.DefaultConfig().
		Port(port).
		Username("postgres").
		Password("postgres").
		Database("leaderboard").
		Version(embeddedpostgres.V16).
		StartTimeout(45 * time.Second).
		StartParameters(map[string]string{"max_connections": "200", "ssl": "off"})
	postgres = embeddedpostgres.NewDatabase(testDatabaseConfig)

	if err := postgres.Start(); err != nil {
		common.Logger.Info("An error occurred while starting the database: %v", err)
		return err
	}
	common.Logger.Info("Successfully started the test database")
	return nil
}

func killActivePostgresInstances() error {
	// only run this if this is linux
	if runtime.GOOS != "linux" {
		common.Logger.Info("Skipping killing active postgres instances because this is not a linux environment")
		return nil
	}
	// Find the PostgreSQL process ID
	cmd := exec.Command("pgrep", "-o postgres")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		//postgres process was likely not found, return no error
		common.Logger.Info("No postgres process found to kill")
		return nil
	}

	// Get the PID from the output
	pidStr := strings.TrimSpace(out.String())
	if pidStr == "" {
		return fmt.Errorf("no postgres process found")
	}

	// Convert PID to integer
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return fmt.Errorf("invalid PID: %v", err)
	}

	// Terminate the process
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process with PID %d: %v", pid, err)
	}

	if err := process.Kill(); err != nil {
		return fmt.Errorf("failed to kill process with PID %d: %v", pid, err)
	}

	fmt.Printf("Successfully killed postgres process with PID %d\n", pid)
	return nil
}

// ShutdownDatabase is a helper that cleans up the test database
func ShutdownDatabase() error {
	common.Logger.Info("Shutting down the test database")
	if postgres != nil {
		if err := postgres.Stop(); err != nil {
			common.Logger.Info("Error occured while shutting down the database: %v", err)
			return err
		}
		common.Logger.Info("Successfully shut down the test database")
	} else {
		common.Logger.Info("Unexpectedly found no database to clean up after tests")
	}
	return nil
}
