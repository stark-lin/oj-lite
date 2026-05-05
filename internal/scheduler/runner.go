// Starts the background scheduling loop and continuously polls pending submissions.

package scheduler

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"oj-lite/internal/platform/config"
	"oj-lite/internal/platform/logger"
)

type runner struct {
	cfg     config.SchedulerConfig
	log     *logger.Logger
	service *service
}

func newRunner(database *sql.DB, log *logger.Logger, cfg config.SchedulerConfig) *runner {
	return &runner{
		cfg:     cfg,
		log:     log,
		service: newService(database, log),
	}
}

func (module *Module) Run(ctx context.Context) {
	module.runner.run(ctx)
}

func (runner *runner) run(ctx context.Context) {
	slotSem := make(chan struct{}, runner.cfg.Concurrency)
	var wg sync.WaitGroup

	defer wg.Wait()

	for {
		if ctx.Err() != nil {
			return
		}

		claimed, err := runner.service.claimPendingSubmissions(ctx, runner.cfg.FetchBatchSize)
		if err != nil {
			runner.log.Errorf("scheduler claim failed: %v", err)
			if !sleepWithContext(ctx, runner.cfg.IdleSleep) {
				return
			}
			continue
		}

		if len(claimed) == 0 {
			if !sleepWithContext(ctx, runner.cfg.IdleSleep) {
				return
			}
			continue
		}

		for _, script := range claimed {
			select {
			case <-ctx.Done():
				return
			case slotSem <- struct{}{}:
			}

			wg.Add(1)
			go func(s claimedSubmission) {
				defer wg.Done()
				defer func() { <-slotSem }()

				if err := runner.service.processClaimedScript(ctx, s); err != nil {
					runner.log.Errorf("submission %d ended with error: %v", s.ID, err)
				}
			}(script)
		}
	}
}

func sleepWithContext(ctx context.Context, duration time.Duration) bool {
	if duration <= 0 {
		duration = time.Second
	}

	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
