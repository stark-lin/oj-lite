package auth

import (
	"oj-lite/internal/platform/logger"
	"oj-lite/internal/platform/session"
	"oj-lite/internal/platform/user"
)

// Module is the public entry point for the auth package; it wires authentication logic and HTTP handlers.
type Module struct {
	handler *handler
}

func New(log *logger.Logger, sessions *session.Manager, users *user.Store) *Module {
	service := newService(log, users)

	return &Module{
		handler: newHandler(log, service, sessions),
	}
}
