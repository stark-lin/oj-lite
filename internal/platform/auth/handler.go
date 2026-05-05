// Handles teacher and student login, sign-out, and self-service password changes.

package auth

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"

	"oj-lite/internal/platform/errs"
	"oj-lite/internal/platform/httpx"
	"oj-lite/internal/platform/logger"
	"oj-lite/internal/platform/password"
	"oj-lite/internal/platform/session"
)

type handler struct {
	log      *logger.Logger
	service  *service
	sessions *session.Manager
}

func newHandler(log *logger.Logger, service *service, sessions *session.Manager) *handler {
	return &handler{
		log:      log,
		service:  service,
		sessions: sessions,
	}
}

func (module *Module) Login(c *gin.Context) {
	module.handler.Login(c)
}

func (module *Module) Logout(c *gin.Context) {
	module.handler.Logout(c)
}

func (module *Module) GetMe(c *gin.Context) {
	module.handler.GetMe(c)
}

func (module *Module) ChangePassword(c *gin.Context) {
	module.handler.ChangePassword(c)
}

func (handler *handler) Login(c *gin.Context) {
	var request loginRequest
	if !httpx.BindJSON(c, &request) {
		return
	}

	if !validateLoginRequest(c, request) {
		return
	}

	result, err := handler.service.Login(c.Request.Context(), request.Username, request.Password)
	if err != nil {
		switch {
		case errors.Is(err, errInvalidCredentials), errors.Is(err, errAccountDisabled):
			handler.sessions.ClearAPISessionCookie(c)
			httpx.AbortUnauthorized(c, "invalid username or password")
		default:
			handler.log.Errorf("login failed: username=%s err=%v", request.Username, err)
			httpx.AbortInternal(c, err)
		}
		return
	}

	claims := handler.sessions.NewClaims(result.User.ID, result.User.Role, result.ClassroomID)
	if err := handler.sessions.SetAPISessionCookie(c, claims); err != nil {
		handler.log.Errorf("set api session cookie failed: user_id=%d err=%v", result.User.ID, err)
		httpx.AbortInternal(c, err)
		return
	}

	httpx.OK(c, gin.H{
		"user": newUserDTO(result.User),
	})
}

func (handler *handler) Logout(c *gin.Context) {
	handler.sessions.ClearAPISessionCookie(c)
	httpx.OK(c, gin.H{
		"ok": true,
	})
}

func (handler *handler) GetMe(c *gin.Context) {
	currentUser, ok := GetCurrentUser(c)
	if !ok {
		httpx.AbortUnauthorized(c, "missing authenticated user in request context")
		return
	}

	account, err := handler.service.GetUser(c.Request.Context(), currentUser.ID)
	if err != nil {
		if errs.IsUnauthenticated(err) {
			handler.sessions.ClearAPISessionCookie(c)
			httpx.AbortUnauthorized(c, "current user does not exist")
			return
		}

		handler.log.Errorf("get current user failed: user_id=%d err=%v", currentUser.ID, err)
		httpx.AbortInternal(c, err)
		return
	}

	httpx.OK(c, gin.H{
		"user": newUserDTO(account),
	})
}

func (handler *handler) ChangePassword(c *gin.Context) {
	currentUser, ok := GetCurrentUser(c)
	if !ok {
		httpx.AbortUnauthorized(c, "missing authenticated user in request context")
		return
	}

	var request changePasswordRequest
	if !httpx.BindJSONOrNotFound(c, &request) {
		return
	}

	if !validateAuthenticatedChangePasswordRequest(c, request) {
		return
	}

	err := handler.service.ChangePassword(
		c.Request.Context(),
		currentUser.ID,
		request.OldPassword,
		request.NewPassword,
	)
	if err != nil {
		switch {
		case errs.IsUnauthenticated(err):
			handler.sessions.ClearAPISessionCookie(c)
			httpx.AbortUnauthorized(c, "current user does not exist")
		case errors.Is(err, errPasswordMismatch):
			httpx.AbortNotFound(c, "old password is incorrect")
		case errors.Is(err, errAccountDisabled):
			httpx.AbortNotFound(c, "current account is unavailable")
		case errs.IsUnavailable(err):
			httpx.AbortNotFound(c, "change password is unavailable")
		default:
			handler.log.Errorf("change password failed: user_id=%d err=%v", currentUser.ID, err)
			httpx.AbortInternal(c, err)
		}
		return
	}

	claims := handler.sessions.NewClaims(currentUser.ID, currentUser.Role, currentUser.ClassroomID)
	if err := handler.sessions.SetAPISessionCookie(c, claims); err != nil {
		handler.log.Errorf("refresh api session cookie after password change failed: user_id=%d err=%v", currentUser.ID, err)
		httpx.AbortInternal(c, err)
		return
	}

	httpx.OK(c, gin.H{
		"ok": true,
	})
}

func (handler *handler) notImplemented(c *gin.Context, action string) {
	handler.log.Warnf("auth scaffold hit: action=%s method=%s path=%s", action, c.Request.Method, c.FullPath())
	httpx.AbortNotImplemented(c, "auth scaffold endpoint is not implemented yet", gin.H{
		"module": "auth",
		"action": action,
		"method": c.Request.Method,
		"path":   c.FullPath(),
	})
}

func validateLoginRequest(c *gin.Context, request loginRequest) bool {
	request.Username = strings.TrimSpace(request.Username)
	if request.Username == "" {
		httpx.AbortValidation(c, "username is required", gin.H{
			"field": "username",
		})
		return false
	}

	if err := password.ValidatePlaintext(request.Password); err != nil {
		httpx.AbortValidation(c, fmt.Sprintf("password length must be between %d and %d", password.MinLength, password.MaxLength), gin.H{
			"field": "password",
		})
		return false
	}

	return true
}

func validateAuthenticatedChangePasswordRequest(c *gin.Context, request changePasswordRequest) bool {
	if err := password.ValidatePlaintext(request.OldPassword); err != nil {
		httpx.AbortNotFoundDetails(c, fmt.Sprintf("old_password length must be between %d and %d", password.MinLength, password.MaxLength), gin.H{
			"field": "old_password",
		})
		return false
	}

	if err := password.ValidatePlaintext(request.NewPassword); err != nil {
		httpx.AbortNotFoundDetails(c, fmt.Sprintf("new_password length must be between %d and %d", password.MinLength, password.MaxLength), gin.H{
			"field": "new_password",
		})
		return false
	}

	if request.OldPassword == request.NewPassword {
		httpx.AbortNotFoundDetails(c, "new password must be different from old password", gin.H{
			"field": "new_password",
		})
		return false
	}

	return true
}
