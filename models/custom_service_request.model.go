package models

import (
	"backend/utils"
	"time"

	"gorm.io/gorm"
)

type CustomServiceRequest struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	Name    string `json:"name"`
	Email   string `json:"email"`
	Service string `json:"service"`
}

func (csr *CustomServiceRequest) BeforeCreate(tx *gorm.DB) error {
	csr.Name = utils.SanitizeString(csr.Name)
	csr.Service = utils.SanitizeString(csr.Service)
	return nil
}
