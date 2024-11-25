package assets

import (
	"embed"
	"io/fs"
)

//go:embed migrations/*.sql
var MigrationFiles embed.FS

// Path is the path to the migrations directory.
const Path = "migrations"

// GetMigrations returns the embedded SQL files.
func GetMigrations() fs.FS {
	return MigrationFiles
}
