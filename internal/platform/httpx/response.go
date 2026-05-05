// Writes consistent JSON success and paginated responses.

package httpx

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const ContextKeyRequestID = "request_id"

type SuccessResponse struct {
	Data      any    `json:"data"`
	RequestID string `json:"request_id,omitempty"`
}

type Pagination struct {
	Page  int   `json:"page"`
	Size  int   `json:"size"`
	Total int64 `json:"total"`
}

type PaginatedResponse struct {
	Data       any        `json:"data"`
	Pagination Pagination `json:"pagination"`
	RequestID  string     `json:"request_id,omitempty"`
}

func OK(c *gin.Context, data any) {
	JSON(c, http.StatusOK, data)
}

func Created(c *gin.Context, data any) {
	JSON(c, http.StatusCreated, data)
}

func JSON(c *gin.Context, status int, data any) {
	c.JSON(status, SuccessResponse{
		Data:      data,
		RequestID: requestID(c),
	})
}

func Paginated(c *gin.Context, data any, page, size int, total int64) {
	c.JSON(http.StatusOK, PaginatedResponse{
		Data: data,
		Pagination: Pagination{
			Page:  page,
			Size:  size,
			Total: total,
		},
		RequestID: requestID(c),
	})
}

func NoContent(c *gin.Context) {
	if rid := requestID(c); rid != "" {
		c.Header("X-Request-ID", rid)
	}
	c.Status(http.StatusNoContent)
}

func requestID(c *gin.Context) string {
	value, ok := c.Get(ContextKeyRequestID)
	if !ok {
		return ""
	}

	requestID, ok := value.(string)
	if !ok {
		return ""
	}

	return requestID
}
