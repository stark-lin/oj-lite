package httpx

import "github.com/gin-gonic/gin"

func BindJSONOrNotFound(c *gin.Context, target any) bool {
	if err := c.ShouldBindJSON(target); err != nil {
		AbortNotFoundDetails(c, "invalid request body", gin.H{
			"reason": err.Error(),
		})
		return false
	}

	return true
}

func PathParamInt64OrNotFound(c *gin.Context, name string) (int64, bool) {
	value := c.Param(name)
	parsed, ok := parseInt64(value)
	if !ok {
		AbortNotFoundDetails(c, "invalid path parameter", gin.H{
			"name":  name,
			"value": value,
		})
		return 0, false
	}

	return parsed, true
}
