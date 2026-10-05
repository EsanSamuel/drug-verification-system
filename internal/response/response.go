// Package response provides consistent JSON response helpers for the API.
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Pagination holds pagination metadata.
type Pagination struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// Success sends a single resource response.
func Success(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{
		"success": true,
		"data":    data,
	})
}

// SuccessWithPagination sends a paginated collection response.
func SuccessWithPagination(c *gin.Context, data any, pagination Pagination) {
	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"data":       data,
		"pagination": pagination,
	})
}

// APIError represents a structured API error.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Error sends a structured error response.
func Error(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{
		"success": false,
		"error":   message,
		"code":    code,
	})
}

// ValidationError sends a 422 with validation details.
func ValidationError(c *gin.Context, message string) {
	Error(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", message)
}

// BadRequest sends a 400 error.
func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, "BAD_REQUEST", message)
}

// Unauthorized sends a 401 error.
func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, "UNAUTHORIZED", message)
}

// Forbidden sends a 403 error.
func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, "FORBIDDEN", message)
}

// NotFound sends a 404 error.
func NotFound(c *gin.Context, code, message string) {
	Error(c, http.StatusNotFound, code, message)
}

// Conflict sends a 409 error.
func Conflict(c *gin.Context, code, message string) {
	Error(c, http.StatusConflict, code, message)
}

// InternalError sends a 500 error. Never expose internal details.
func InternalError(c *gin.Context) {
	Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred")
}

// NewPagination builds pagination metadata.
func NewPagination(page, limit, total int) Pagination {
	totalPages := 0
	if limit > 0 {
		totalPages = (total + limit - 1) / limit
	}
	return Pagination{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}
}
