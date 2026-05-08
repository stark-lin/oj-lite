package seed

import (
	"context"
	"database/sql"
)

func SeedDemoAccounts(ctx context.Context, database *sql.DB) error {
	return New(database).SeedDemoAccounts(ctx)
}
