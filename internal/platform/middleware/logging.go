// Logs basic access data for each request.

package middleware

import (
	"time"

	"github.com/gin-gonic/gin"

	"oj-lite/internal/platform/httpx"
	"oj-lite/internal/platform/logger"
)

func Logging(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startedAt := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		requestID, _ := c.Get(httpx.ContextKeyRequestID)
		log.Infof(
			"request_id=%v method=%s path=%s status=%d latency=%s client_ip=%s",
			requestID,
			method,
			path,
			c.Writer.Status(),
			time.Since(startedAt),
			c.ClientIP(),
		)
	}
}
