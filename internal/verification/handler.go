package verification

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/drug-verification/server/internal/response"
)

// Handler handles verification HTTP requests.
type Handler struct {
	service *Service
}

// NewHandler creates a new verification handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterPublicRoutes registers public (unauthenticated) verification routes.
func (h *Handler) RegisterPublicRoutes(rg *gin.RouterGroup) {
	rg.GET("/verify/:serial", h.Verify)
}

// RegisterProtectedRoutes registers authenticated verification log routes.
func (h *Handler) RegisterProtectedRoutes(rg *gin.RouterGroup) {
	rg.GET("/verification-logs", h.ListLogs)
	rg.GET("/drugs/:id/verification-logs", h.ListLogsByDrug)
}

// Verify handles public drug verification.
func (h *Handler) Verify(c *gin.Context) {
	serial := c.Param("serial")

	result, err := h.service.Verify(c.Request.Context(), VerifyRequest{
		SerialNumber: serial,
		IPAddress:    c.ClientIP(),
		UserAgent:    c.Request.UserAgent(),
	})
	if err != nil {
		response.InternalError(c)
		return
	}

	c.JSON(http.StatusOK, result)
}

// ListLogs handles listing verification logs.
func (h *Handler) ListLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	params := LogListParams{
		Page:         page,
		Limit:        limit,
		SerialNumber: c.Query("serial_number"),
		Result:       c.Query("result"),
		FromDate:     c.Query("from_date"),
		ToDate:       c.Query("to_date"),
	}

	results, total, err := h.service.ListLogs(c.Request.Context(), params)
	if err != nil {
		response.InternalError(c)
		return
	}

	response.SuccessWithPagination(c, results, response.NewPagination(page, limit, int(total)))
}

// ListLogsByDrug handles listing verification logs for a specific drug.
func (h *Handler) ListLogsByDrug(c *gin.Context) {
	drugID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid drug ID")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	params := LogListParams{
		Page:     page,
		Limit:    limit,
		Result:   c.Query("result"),
		FromDate: c.Query("from_date"),
		ToDate:   c.Query("to_date"),
	}

	results, total, err := h.service.ListLogsByDrug(c.Request.Context(), drugID, params)
	if err != nil {
		response.InternalError(c)
		return
	}

	response.SuccessWithPagination(c, results, response.NewPagination(page, limit, int(total)))
}
