// Enforces the local admin localhost access restriction.

package middleware

import (
	"net"

	"github.com/gin-gonic/gin"

	"oj-lite/internal/platform/auth"
	"oj-lite/internal/platform/httpx"
)

func AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isLocalIP(c.ClientIP()) {
			httpx.AbortNotFound(c, "route not found")
			return
		}

		auth.SetCurrentUser(c, auth.CurrentUser{
			Username: "admin",
			Role:     "admin",
		})

		c.Next()
	}
}

func isLocalIP(value string) bool {
	ip := net.ParseIP(value)
	if ip == nil {
		return false
	}

	return ip.IsLoopback()
}
