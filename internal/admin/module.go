package admin

import (
	"database/sql"

	"oj-lite/internal/platform/logger"
)

// Module is the public entry point for the admin package; it holds auth, service, and HTTP handler dependencies.
type Module struct {
	localAuth *localAuth
	repo      *repo
	service   *service
	handler   *handler
}

func New(database *sql.DB, log *logger.Logger) *Module {
	repo := newRepo(database, log)
	service := newService(log, repo)

	module := &Module{
		localAuth: newLocalAuth(log),
		repo:      repo,
		service:   service,
	}
	module.handler = newHandler(log, service)

	return module
}
