package utils

import (
	"github.com/gofiber/fiber/v2"
)

// Response struct
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// Meta struct
type Meta struct {
	Total   int64 `json:"total"`
	Page    int   `json:"page"`
	PerPage int   `json:"per_page"`
}

// Success response
func SuccessResponse(c *fiber.Ctx, statusCode int, message string, data interface{}) error {
	return c.Status(statusCode).JSON(Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Success with meta
func SuccessResponseWithMeta(c *fiber.Ctx, statusCode int, message string, data interface{}, meta *Meta) error {
	return c.Status(statusCode).JSON(Response{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

// Error response
func ErrorResponse(c *fiber.Ctx, statusCode int, message string, err error) error {
	errorMsg := ""
	if err != nil {
		errorMsg = err.Error()
	}
	return c.Status(statusCode).JSON(Response{
		Success: false,
		Message: message,
		Error:   errorMsg,
	})
}

// Validation error
func ValidationErrorResponse(c *fiber.Ctx, message string, errors map[string]string) error {
	return c.Status(fiber.StatusUnprocessableEntity).JSON(Response{
		Success: false,
		Message: message,
		Error:   "Validation failed",
		Data:    errors,
	})
}
