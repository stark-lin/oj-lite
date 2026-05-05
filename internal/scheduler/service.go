// Encapsulates core scheduler orchestration across the submission and judge modules.

package scheduler

import (
	"context"
	"database/sql"

	"oj-lite/internal/platform/logger"
)

type service struct {
	lease  *lease
	log    *logger.Logger
	worker *worker
}

func newService(database *sql.DB, log *logger.Logger) *service {
	return &service{
		lease:  newLease(database, log),
		log:    log,
		worker: newWorker(database, log),
	}
}

func (service *service) claimPendingSubmissions(ctx context.Context, limit int) ([]claimedSubmission, error) {
	return service.lease.claimPendingSubmissions(ctx, limit)
}

func (service *service) processClaimedScript(ctx context.Context, script claimedSubmission) error {
	return service.worker.processClaimedScript(ctx, script)
}
