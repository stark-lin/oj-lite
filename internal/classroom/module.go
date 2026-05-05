package classroom

import (
	"database/sql"

	"oj-lite/internal/platform/logger"
)

// Module is the public entry point for the classroom package; it wires the repository, service, and HTTP handler.
type Module struct {
	repo    *repo
	service *service
	handler *handler
}

func New(database *sql.DB, log *logger.Logger) *Module {
	repo := newRepo(database, log)
	service := newService(log, repo)

	module := &Module{
		repo:    repo,
		service: service,
	}
	module.handler = newHandler(log, service)

	return module
}
