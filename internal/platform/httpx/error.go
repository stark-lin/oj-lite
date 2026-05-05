// Defines the shared error payload and HTTP error response helpers.

package httpx

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

type ErrorResponse struct {
	Error     APIError `json:"error"`
	RequestID string   `json:"request_id,omitempty"`
}

func Abort(c *gin.Context, status int, code, message string, details any) {
	c.AbortWithStatusJSON(status, ErrorResponse{
		Error: APIError{
			Code:    code,
			Message: message,
			Details: details,
		},
		RequestID: requestID(c),
	})
}

func AbortBadRequest(c *gin.Context, message string, details any) {
	Abort(c, http.StatusBadRequest, "bad_request", message, details)
}

func AbortValidation(c *gin.Context, message string, details any) {
	Abort(c, http.StatusBadRequest, "validation_error", message, details)
}

func AbortUnauthorized(c *gin.Context, message string) {
	Abort(c, http.StatusUnauthorized, "unauthorized", message, nil)
}

func AbortForbidden(c *gin.Context, message string) {
	Abort(c, http.StatusForbidden, "forbidden", message, nil)
}

func AbortNotFound(c *gin.Context, message string) {
	Abort(c, http.StatusNotFound, "not_found", message, nil)
}

func AbortNotFoundDetails(c *gin.Context, message string, details any) {
	Abort(c, http.StatusNotFound, "not_found", message, details)
}

func AbortNotImplemented(c *gin.Context, message string, details any) {
	Abort(c, http.StatusNotImplemented, "not_implemented", message, details)
}

func AbortInternal(c *gin.Context, err error) {
	Abort(c, http.StatusInternalServerError, "internal_error", "internal server error", gin.H{
		"reason": err.Error(),
	})
}
