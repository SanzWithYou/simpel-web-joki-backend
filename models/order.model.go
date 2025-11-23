package models

import (
	"backend/utils"
	"time"

	"gorm.io/gorm"
)

type Order struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	Username      string `json:"-"`
	Password      string `json:"-"`
	Joki          string `json:"joki"`
	BuktiTransfer string `json:"bukti_transfer"`
}

func (o *Order) BeforeCreate(tx *gorm.DB) error {
	o.Joki = utils.SanitizeString(o.Joki)
	return nil
}
