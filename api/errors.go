package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Error string `json:"Error"`
}

// RespondWithError sends a standardized error response based on status code
func RespondWithError(c *gin.Context, statusCode int, customMessage ...string) {
	var message string

	// If a custom message is provided, use it
	if len(customMessage) > 0 && customMessage[0] != "" {
		message = customMessage[0]
	} else {
		// Otherwise use standard messages based on status code
		switch statusCode {
		case http.StatusBadRequest: // 400
			message = "The request body is invalid"
		case http.StatusUnauthorized: // 401
			message = "Unauthorized"
		case http.StatusForbidden: // 403
			message = "You don't have permission on this resource"
		case http.StatusNotFound: // 404
			message = "Not found"
		case http.StatusMethodNotAllowed: // 405
			message = "Method not allowed"
		case http.StatusConflict: // 409
			message = "Enrollment data is invalid"
		case http.StatusInternalServerError: // 500
			message = "Internal server error"
		default:
			message = "An error occurred"
		}
	}

	c.JSON(statusCode, ErrorResponse{Error: message})
}
