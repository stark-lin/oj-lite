// Embeds migration files under `sql/` for runtime use.

package db

import (
	"embed"
)

var (
	// MigrationFS exposes embedded SQLite migration files under sql/*.sql.
	//
	//go:embed sql/*.sql
	MigrationFS embed.FS
)
