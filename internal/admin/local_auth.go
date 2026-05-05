// Verifies the fixed admin username and local plaintext password.

package admin

import "oj-lite/internal/platform/logger"

type localAuth struct {
	log *logger.Logger
}

func newLocalAuth(log *logger.Logger) *localAuth {
	return &localAuth{log: log}
}
