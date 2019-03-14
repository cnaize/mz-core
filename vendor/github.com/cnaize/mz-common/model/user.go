package model

import (
	"github.com/dgrijalva/jwt-go"
	"time"
)

type User struct {
	Base
	Username  string     `json:"username" gorm:"unique_index"`
	Token     string     `json:"token" form:"token"`
	PassHash  string     `json:"-"`
	DeletedAt *time.Time `json:"-"`
}

type Token struct {
	jwt.StandardClaims
	Username string
}
