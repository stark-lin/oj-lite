// Verifies the business signed cookie and injects the current user into context.

package middleware

import (
	"github.com/gin-gonic/gin"

	"oj-lite/internal/platform/auth"
	"oj-lite/internal/platform/httpx"
	"oj-lite/internal/platform/logger"
	"oj-lite/internal/platform/session"
)

func APIAuth(sessions *session.Manager, log *logger.Logger, allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := sessions.ReadAPISessionCookie(c)
		if err != nil {
			sessions.ClearAPISessionCookie(c)
			if !session.IsMissingSession(err) {
				log.Warnf("read api session cookie failed: path=%s err=%v", c.FullPath(), err)
			}
			httpx.AbortUnauthorized(c, "missing or invalid api session")
			return
		}

		if len(allowedRoles) > 0 && !containsRole(allowedRoles, claims.Role) {
			httpx.AbortNotFound(c, "route not found")
			return
		}

		auth.SetCurrentUser(c, auth.CurrentUser{
			ID:          claims.UserID,
			Role:        claims.Role,
			ClassroomID: claims.ClassroomID,
		})

		if sessions.ShouldRefresh(claims) {
			refreshed := sessions.RefreshClaims(claims)
			if err := sessions.SetAPISessionCookie(c, refreshed); err != nil {
				log.Errorf("refresh api session cookie failed: user_id=%d err=%v", claims.UserID, err)
				httpx.AbortInternal(c, err)
				return
			}
		}

		c.Next()
	}
}

func containsRole(allowedRoles []string, target string) bool {
	for _, role := range allowedRoles {
		if role == target {
			return true
		}
	}

	return false
}
