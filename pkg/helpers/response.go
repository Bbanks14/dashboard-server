// File: pkg/helpers/response.go
package helpers

import (
	"github.com/gin-gonic/gin"
)

// Response is the standard API response structure
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// SuccessResponse returns a standardized success response
func SuccessResponse(ctx *gin.Context, statusCode int, data interface{}) {
	ctx.JSON(statusCode, Response{
		Success: true,
		Data:    data,
	})
}

// ErrorResponse returns a standardized error response
func ErrorResponse(ctx *gin.Context, statusCode int, message string) {
	ctx.JSON(statusCode, Response{
		Success: false,
		Message: message,
	})
}
