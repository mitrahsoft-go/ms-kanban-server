package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt/v5"
	"github.com/ms-kanban-server/config"
	"github.com/ms-kanban-server/internal/handlers/dto"
	"github.com/ms-kanban-server/internal/pkg/response"
	"go.uber.org/zap"
)

func GenerateJWT(role string, id uuid.UUID, logger *zap.Logger) (string, *response.Error) {
	return generateJWT(role, id, 15*time.Minute, logger)
}

func GenerateRefreshJWT(role string, id uuid.UUID, logger *zap.Logger) (string, *response.Error) {
	return generateJWT(role, id, 1*time.Hour, logger)
}

func generateJWT(role string, id uuid.UUID, ttl time.Duration, logger *zap.Logger) (string, *response.Error) {
	var jwtKey = config.GetEnv("JWT_SECRET_KEY", "")

	expirationTime := time.Now().Add(ttl)

	claims := &dto.ClaimsJWT{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Role:   role,
		UserId: id,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(jwtKey))
	if err != nil {
		errorResponse := response.Error{
			Code:       response.ErrInternalServerError,
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed generating the token",
			Details: []response.Details{
				{
					Field:   "token",
					Message: err.Error()},
			},
		}
		logger.Error("Failed generating the token in Middleware Layer",
			zap.String("ERROR : ", fmt.Sprintf("%v", errorResponse)))
		return "", &errorResponse
	}

	return tokenString, nil
}
