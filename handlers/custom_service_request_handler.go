package handlers

import (
	"backend/models"
	"backend/utils"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// Handler struct
type CustomServiceRequestHandler struct {
	DB *gorm.DB
}

// Constructor
func NewCustomServiceRequestHandler(db *gorm.DB) *CustomServiceRequestHandler {
	return &CustomServiceRequestHandler{DB: db}
}

// Create request
func (h *CustomServiceRequestHandler) CreateCustomServiceRequest(c *fiber.Ctx) error {
	// Parse body
	var req models.CustomServiceRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	// Validate input
	if req.Name == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Name is required", nil)
	}
	if req.Email == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Email is required", nil)
	}
	if req.Service == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Service description is required", nil)
	}

	// Validate email
	if err := utils.Validator.Var(req.Email, "email"); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid email format", err)
	}

	// Save to DB
	if err := h.DB.Create(&req).Error; err != nil {
		utils.LogError(err)
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to save request", err)
	}

	// Send email
	go func() {
		if err := utils.SendCustomServiceRequestEmail(req.ID, req.Name, req.Email, req.Service); err != nil {
			utils.LogError(err)
		}
	}()

	return utils.SuccessResponse(c, fiber.StatusCreated, "Request submitted successfully", req)
}

// Delete request
func (h *CustomServiceRequestHandler) DeleteCustomServiceRequest(c *fiber.Ctx) error {
	// Parse ID
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid ID format", err)
	}

	// Soft delete
	if err := h.DB.Delete(&models.CustomServiceRequest{}, id).Error; err != nil {
		utils.LogError(err)
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete request", err)
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Request deleted successfully", nil)
}
