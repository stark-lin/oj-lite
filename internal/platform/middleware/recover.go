// Recovers from panics and returns the shared 500 error response.

package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"oj-lite/internal/platform/httpx"
	"oj-lite/internal/platform/logger"
)

func Recovery(log *logger.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		log.Errorf("panic recovered: %v", recovered)
		httpx.AbortInternal(c, fmt.Errorf("panic: %v", recovered))
	})
}
