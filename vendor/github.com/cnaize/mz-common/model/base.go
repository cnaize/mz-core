package model

import (
	"time"
)

type Base struct {
	ID        uint      `json:"id,omitempty" gorm:"primary_key" form:"id"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}
