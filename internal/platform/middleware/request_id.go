// Injects a request id into each request for log tracing.

package middleware

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"

	"oj-lite/internal/platform/httpx"
)

var requestCounter uint64

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = fmt.Sprintf("req-%d-%06d", time.Now().UnixNano(), atomic.AddUint64(&requestCounter, 1)%1000000)
		}

		c.Set(httpx.ContextKeyRequestID, requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}
