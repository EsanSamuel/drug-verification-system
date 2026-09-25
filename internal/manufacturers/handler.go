package manufacturers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/drug-verification/server/internal/response"
)

// Handler handles manufacturer HTTP requests.
type Handler struct {
	service *Service
}

// NewHandler creates a new manufacturer handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers manufacturer routes (all protected).
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	m := rg.Group("/manufacturers")
	m.POST("", h.Create)
	m.GET("", h.List)
	m.GET("/:id", h.Get)
	m.PATCH("/:id", h.Update)
	m.DELETE("/:id", h.Delete)
}

// Create handles manufacturer creation.
func (h *Handler) Create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	result, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		if isValidationErr(err) {
			response.ValidationError(c, err.Error())
			return
		}
		response.InternalError(c)
		return
	}

	response.Success(c, http.StatusCreated, result)
}

// Get handles retrieving a single manufacturer.
func (h *Handler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid manufacturer ID")
		return
	}

	result, err := h.service.Get(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, ErrManufacturerNotFound) {
			response.NotFound(c, "MANUFACTURER_NOT_FOUND", "Manufacturer not found")
			return
		}
		response.InternalError(c)
		return
	}

	response.Success(c, http.StatusOK, result)
}

// List handles listing manufacturers with pagination.
func (h *Handler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	results, total, err := h.service.List(c.Request.Context(), ListParams{
		Page:  page,
		Limit: limit,
	})
	if err != nil {
		response.InternalError(c)
		return
	}

	response.SuccessWithPagination(c, results, response.NewPagination(page, limit, int(total)))
}

// Update handles manufacturer updates.
func (h *Handler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid manufacturer ID")
		return
	}

	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	result, err := h.service.Update(c.Request.Context(), id, req)
	if err != nil {
		if errors.Is(err, ErrManufacturerNotFound) {
			response.NotFound(c, "MANUFACTURER_NOT_FOUND", "Manufacturer not found")
			return
		}
		if isValidationErr(err) {
			response.ValidationError(c, err.Error())
			return
		}
		response.InternalError(c)
		return
	}

	response.Success(c, http.StatusOK, result)
}

// Delete handles manufacturer removal.
func (h *Handler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid manufacturer ID")
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, ErrManufacturerNotFound) {
			response.NotFound(c, "MANUFACTURER_NOT_FOUND", "Manufacturer not found")
			return
		}
		if errors.Is(err, ErrManufacturerInUse) {
			response.Conflict(c, "MANUFACTURER_IN_USE", "Cannot delete manufacturer with existing drugs")
			return
		}
		response.InternalError(c)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func isValidationErr(err error) bool {
	msg := err.Error()
	return len(msg) > 0 && msg[0] >= 'a' && msg[0] <= 'z'
}
