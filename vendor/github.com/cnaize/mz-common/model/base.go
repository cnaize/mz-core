package model

import (
	"time"
)

type Base struct {
	ID        uint      `json:"-" form:"-" gorm:"primary_key"`
	CreatedAt time.Time `json:"-" form:"-"`
	UpdatedAt time.Time `json:"-" form:"-"`
}
