package model

import (
	"github.com/dgrijalva/jwt-go"
	"time"
)

type User struct {
	Base
	Username  string     `json:"username" form:"-" gorm:"unique_index"`
	Token     string     `json:"token" form:"token"`
	PassHash  string     `json:"-" form:"-"`
	DeletedAt *time.Time `json:"-" form:"-"`
}

type Token struct {
	jwt.StandardClaims
	Username string
}
