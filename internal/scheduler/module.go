package scheduler

import (
	"database/sql"

	"oj-lite/internal/platform/config"
	"oj-lite/internal/platform/logger"
)

type Module struct {
	runner *runner
}

func New(database *sql.DB, log *logger.Logger, cfg config.SchedulerConfig) *Module {
	return &Module{
		runner: newRunner(database, log, cfg),
	}
}
