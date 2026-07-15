package dto

import (
	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt/v5"
)

type ClaimsJWT struct {
	Role   string    `json:"role"`
	UserId uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}
