package dto

import "time"

// Create order request
type CreateOrderRequest struct {
	Username      string `json:"username" validate:"required"`
	Password      string `json:"password" validate:"required"`
	Joki          string `json:"joki" validate:"required,min=3,max=100"`
	BuktiTransfer string `json:"bukti_transfer" validate:"required"`
}

// Validate request
func (r *CreateOrderRequest) Validate() map[string]string {
	errors := make(map[string]string)

	// Manual validation
	if r.Username == "" {
		errors["username"] = "Username wajib diisi"
	}

	if r.Password == "" {
		errors["password"] = "Password wajib diisi"
	}

	if r.Joki == "" {
		errors["joki"] = "Jenis joki wajib diisi"
	} else if len(r.Joki) < 3 {
		errors["joki"] = "Jenis joki minimal 3 karakter"
	} else if len(r.Joki) > 100 {
		errors["joki"] = "Jenis joki maksimal 100 karakter"
	}

	if r.BuktiTransfer == "" {
		errors["bukti_transfer"] = "Bukti transfer wajib diisi"
	}

	return errors
}

// Order detail response
type OrderResponse struct {
	ID            uint      `json:"id"`
	Username      string    `json:"username"`
	Password      string    `json:"password"`
	Joki          string    `json:"joki"`
	BuktiTransfer string    `json:"bukti_transfer"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Order list response
type OrderListResponse struct {
	ID            uint   `json:"id"`
	Joki          string `json:"joki"`
	BuktiTransfer string `json:"bukti_transfer"`
}
