package auth

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/drug-verification/server/internal/middleware"
	"github.com/drug-verification/server/internal/response"
)

// Handler handles authentication HTTP requests.
type Handler struct {
	service *Service
}

// NewHandler creates a new auth handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers auth routes.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	auth := rg.Group("/auth")
	auth.POST("/register", h.Register)
	auth.POST("/login", h.Login)
}

// RegisterProtectedRoutes registers authenticated auth routes.
func (h *Handler) RegisterProtectedRoutes(rg *gin.RouterGroup) {
	auth := rg.Group("/auth")
	auth.GET("/me", h.Me)
}

// Register handles user registration.
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	result, err := h.service.Register(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, ErrEmailTaken) {
			response.Conflict(c, "EMAIL_TAKEN", "Email already registered")
			return
		}
		// Validation errors from validateRegister start with the field name
		if isValidationError(err) {
			response.ValidationError(c, err.Error())
			return
		}
		response.InternalError(c)
		return
	}

	response.Success(c, http.StatusCreated, result)
}

// Login handles user authentication.
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	result, err := h.service.Login(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			response.Unauthorized(c, "Invalid email or password")
			return
		}
		if isValidationError(err) {
			response.ValidationError(c, err.Error())
			return
		}
		response.InternalError(c)
		return
	}

	response.Success(c, http.StatusOK, result)
}

// Me returns the current authenticated user.
func (h *Handler) Me(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "Authentication required")
		return
	}

	user, err := h.service.GetCurrentUser(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			response.NotFound(c, "USER_NOT_FOUND", "User not found")
			return
		}
		response.InternalError(c)
		return
	}

	response.Success(c, http.StatusOK, user)
}

// isValidationError checks if an error is a validation error (simple message, not wrapped).
func isValidationError(err error) bool {
	// Validation errors from our validate functions are simple fmt.Errorf messages
	// without colons, while internal errors are wrapped with context prefixes like "auth:"
	msg := err.Error()
	return len(msg) > 0 && !strings.Contains(msg, ":")
}
