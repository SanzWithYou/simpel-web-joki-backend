package utils

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validator instance
var Validator = validator.New()

// Validasi struct
func ValidateStruct(s interface{}) map[string]string {
	errors := make(map[string]string)

	err := Validator.Struct(s)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			fieldName := err.Field()

			// Pesan kustom
			switch err.Tag() {
			case "required":
				errors[fieldName] = fieldName + " is required"
			case "min":
				errors[fieldName] = fieldName + " must be at least " + err.Param() + " characters"
			case "max":
				errors[fieldName] = fieldName + " must be maximum " + err.Param() + " characters"
			case "email":
				errors[fieldName] = fieldName + " must be a valid email"
			default:
				errors[fieldName] = fieldName + " is invalid"
			}
		}
	}

	return errors
}

// Validasi password
func ValidatePassword(password string) map[string]string {
	errors := make(map[string]string)

	if len(password) < 8 {
		errors["password"] = "Password must be at least 8 characters"
	}

	if !regexp.MustCompile(`[A-Z]`).MatchString(password) {
		errors["password"] = "Password must contain at least one uppercase letter"
	}

	if !regexp.MustCompile(`[a-z]`).MatchString(password) {
		errors["password"] = "Password must contain at least one lowercase letter"
	}

	if !regexp.MustCompile(`[0-9]`).MatchString(password) {
		errors["password"] = "Password must contain at least one number"
	}

	return errors
}

// Bersihkan string
func SanitizeString(s string) string {
	// Trim
	s = strings.TrimSpace(s)
	// Hapus karakter khusus
	reg := regexp.MustCompile(`[^\w\s]`)
	s = reg.ReplaceAllString(s, "")
	return s
}

// Konversi map ke error
func MapToError(errMap map[string]string) error {
	if len(errMap) == 0 {
		return nil
	}

	var errMsgs []string
	for field, msg := range errMap {
		errMsgs = append(errMsgs, fmt.Sprintf("%s: %s", field, msg))
	}

	return errors.New(strings.Join(errMsgs, "; "))
}
