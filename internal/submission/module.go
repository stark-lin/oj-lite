package submission

import (
	"database/sql"

	"oj-lite/internal/platform/logger"
)

// Module is the public entry point for the submission package; it wires the repository, service, and HTTP handler.
type Module struct {
	repo    *repo
	service *service
	handler *handler
}

func New(database *sql.DB, log *logger.Logger) *Module {
	repo := newRepo(database, log)
	module := &Module{
		repo:    repo,
		service: newService(log, repo),
	}
	module.handler = newHandler(log, module.service)

	return module
}
