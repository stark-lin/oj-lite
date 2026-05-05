// Provides common user operations, including creating teachers and students, changing passwords, and updating account status.

package user

import "oj-lite/internal/platform/logger"

type service struct {
	log *logger.Logger
}

func newService(log *logger.Logger) *service {
	return &service{log: log}
}
