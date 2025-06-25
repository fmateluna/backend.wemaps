package dto

import (
	"github.com/golang-jwt/jwt/v5"
)

type WeMapsAddress struct {
	FormattedAddress string  `json:"formatted_address"`
	Latitude         float64 `json:"latitude"`
	Longitude        float64 `json:"longitude"`
}

type Claims struct {
	UserAlias string `json:"user_alias"`
	jwt.RegisteredClaims
}
