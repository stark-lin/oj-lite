package seed

import (
	"context"
	"database/sql"
)

// Module wires the seed repository and service.
type Module struct {
	repo    *repo
	service *service
}

func New(database *sql.DB) *Module {
	repo := newRepo(database)

	return &Module{
		repo:    repo,
		service: newService(repo),
	}
}

func (module *Module) SeedDemoAccounts(ctx context.Context) error {
	return module.service.SeedDemoAccounts(ctx)
}
