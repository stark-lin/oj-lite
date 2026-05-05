// Parses request bodies, path parameters, and query parameters.

package httpx

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func BindJSON(c *gin.Context, target any) bool {
	if err := c.ShouldBindJSON(target); err != nil {
		AbortBadRequest(c, "invalid request body", gin.H{
			"reason": err.Error(),
		})
		return false
	}

	return true
}

func PathParamInt64(c *gin.Context, name string) (int64, bool) {
	value := c.Param(name)
	parsed, ok := parseInt64(value)
	if !ok {
		AbortBadRequest(c, "invalid path parameter", gin.H{
			"name":  name,
			"value": value,
		})
		return 0, false
	}

	return parsed, true
}

func parseInt64(value string) (int64, bool) {
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, false
	}

	return parsed, true
}

func QueryInt(c *gin.Context, name string, fallback int) (int, bool) {
	value := c.Query(name)
	if value == "" {
		return fallback, true
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		AbortBadRequest(c, "invalid query parameter", gin.H{
			"name":  name,
			"value": value,
		})
		return 0, false
	}

	return parsed, true
}
