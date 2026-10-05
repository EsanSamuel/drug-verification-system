package drugs

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/drug-verification/server/internal/middleware"
	"github.com/drug-verification/server/internal/response"
	"github.com/drug-verification/server/pkg/qr"
)

// Handler handles drug HTTP requests.
type Handler struct {
	service *Service
	qrGen   *qr.Generator
}

// NewHandler creates a new drug handler.
func NewHandler(service *Service, qrGen *qr.Generator) *Handler {
	return &Handler{service: service, qrGen: qrGen}
}

// RegisterRoutes registers drug routes (all protected).
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	d := rg.Group("/drugs")
	d.POST("", h.Create)
	d.GET("", h.List)
	d.GET("/:id", h.Get)
	d.PATCH("/:id", h.Update)
	d.DELETE("/:id", h.Delete)
	d.POST("/:id/units", h.GenerateUnits)
	d.GET("/:id/units", h.ListUnits)

	// Drug unit QR
	units := rg.Group("/drug-units")
	units.GET("/:id/qr", h.GetQR)
}

// Create handles drug creation.
func (h *Handler) Create(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "Authentication required")
		return
	}

	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	result, err := h.service.Create(c.Request.Context(), req, userID)
	if err != nil {
		if errors.Is(err, ErrDuplicateBatch) {
			response.Conflict(c, "DUPLICATE_BATCH", "A drug with this batch number already exists for this manufacturer")
			return
		}
		if isValidationErr(err) {
			response.ValidationError(c, err.Error())
			return
		}
		response.InternalError(c)
		return
	}

	response.Success(c, http.StatusCreated, result)
}

// Get handles retrieving a single drug.
func (h *Handler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid drug ID")
		return
	}

	result, err := h.service.Get(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, ErrDrugNotFound) {
			response.NotFound(c, "DRUG_NOT_FOUND", "Drug not found")
			return
		}
		response.InternalError(c)
		return
	}

	response.Success(c, http.StatusOK, result)
}

// List handles listing drugs with pagination and filtering.
func (h *Handler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	params := ListParams{
		Page:           page,
		Limit:          limit,
		Name:           c.Query("name"),
		GenericName:    c.Query("generic_name"),
		BatchNumber:    c.Query("batch_number"),
		Status:         c.Query("status"),
		ManufacturerID: c.Query("manufacturer_id"),
		ExpiryBefore:   c.Query("expiry_before"),
		ExpiryAfter:    c.Query("expiry_after"),
	}

	results, total, err := h.service.List(c.Request.Context(), params)
	if err != nil {
		response.InternalError(c)
		return
	}

	response.SuccessWithPagination(c, results, response.NewPagination(page, limit, int(total)))
}

// Update handles drug updates.
func (h *Handler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid drug ID")
		return
	}

	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	result, err := h.service.Update(c.Request.Context(), id, req)
	if err != nil {
		if errors.Is(err, ErrDrugNotFound) {
			response.NotFound(c, "DRUG_NOT_FOUND", "Drug not found")
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

// Delete handles drug deactivation (soft delete via status change).
func (h *Handler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid drug ID")
		return
	}

	if err := h.service.UpdateStatus(c.Request.Context(), id, "suspended"); err != nil {
		if errors.Is(err, ErrDrugNotFound) {
			response.NotFound(c, "DRUG_NOT_FOUND", "Drug not found")
			return
		}
		response.InternalError(c)
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Drug has been suspended"})
}

// GenerateUnits handles drug unit generation.
func (h *Handler) GenerateUnits(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid drug ID")
		return
	}

	var req GenerateUnitsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	units, err := h.service.GenerateUnits(c.Request.Context(), id, req.Quantity)
	if err != nil {
		if errors.Is(err, ErrDrugNotFound) {
			response.NotFound(c, "DRUG_NOT_FOUND", "Drug not found")
			return
		}
		if isValidationErr(err) {
			response.ValidationError(c, err.Error())
			return
		}
		response.InternalError(c)
		return
	}

	response.Success(c, http.StatusCreated, units)
}

// ListUnits handles listing drug units for a drug.
func (h *Handler) ListUnits(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid drug ID")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	results, total, err := h.service.ListUnits(c.Request.Context(), id, page, limit)
	if err != nil {
		response.InternalError(c)
		return
	}

	response.SuccessWithPagination(c, results, response.NewPagination(page, limit, int(total)))
}

// GetQR returns a QR code image for a drug unit.
func (h *Handler) GetQR(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid drug unit ID")
		return
	}

	unit, err := h.service.GetUnit(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, ErrDrugUnitNotFound) {
			response.NotFound(c, "DRUG_UNIT_NOT_FOUND", "Drug unit not found")
			return
		}
		response.InternalError(c)
		return
	}

	png, err := h.qrGen.GeneratePNG(unit.SerialNumber, 256)
	if err != nil {
		response.InternalError(c)
		return
	}

	c.Data(http.StatusOK, "image/png", png)
}

func isValidationErr(err error) bool {
	msg := err.Error()
	return len(msg) > 0 && !strings.Contains(msg, ":")
}
