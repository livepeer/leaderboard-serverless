package testutils

import (
	"net"
	"os"
	"testing"

	"github.com/livepeer/leaderboard-serverless/common"
)

// findAvailablePort finds an available port on the system
func findAvailablePort() (int, error) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}

func TestMain(m *testing.M) {

	//check if LOG_LEVEL is set.  If not, set it to DEBUG.
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "debug"
	}
	common.Logger.SetLevel(logLevel)

	// Find a port to run the tests on
	port, err := findAvailablePort()
	if err != nil {
		common.Logger.Info("Failed to find an available port: %v", err)
		os.Exit(1)
	}
	portUint32 := uint32(port)
	common.Logger.Info("Using port: %d", portUint32)

	// Initialize the test database
	if err := InitDB(portUint32); err != nil {
		common.Logger.Info("Expected no error when initializing the database: %v", err)
		ShutdownDatabase()
		os.Exit(1)
	}

	// Run the tests
	code := m.Run()

	// Clean up resources if needed
	if err = ShutdownDatabase(); err != nil {
		common.Logger.Info("Expected no error when cleaning up the database: %v", err)
	}

	// Exit with the appropriate code
	os.Exit(code)
}
