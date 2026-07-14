package utils

import (
	"net/http"
	"regexp"

	"github.com/gofrs/uuid"
	"github.com/ms-kanban-server/internal/pkg/response"
	"golang.org/x/crypto/bcrypt"
)

func IsValidPassword(storedHash, enteredPassword string) bool {

	err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(enteredPassword))
	if err != nil {
		return true
	}
	return false
}

func ValidatedPassword(password string) bool {

	emailRegex := regexp.MustCompile(`^[a-z0-9A-Z._%+\-]{8,}$`)
	return emailRegex.MatchString(password)
}

func HashPassword(password string) (int, string, *response.Error) {

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		errorResponse := response.Error{
			Code:    response.ErrInternalServerError,
			Message: "InternalServerError",
			Details: []response.Details{
				{
					Field:   "Password",
					Message: "Error while hashing password : " + err.Error(),
				},
			},
		}
		return http.StatusInternalServerError, "", &errorResponse
	}
	return http.StatusOK, string(bytes), nil
}

func StringToUUID(idStr string) (uuid.UUID, *response.Error) {

	if idStr == "" {
		return uuid.Nil, nil
	}
	id, err := uuid.FromString(idStr)
	if err != nil {
		errorResponse := response.Error{
			Code:    response.ErrBadRequest,
			Message: "Invalid Format",
			Details: []response.Details{
				{
					Field:   "ID",
					Message: "Failed to parse UUID : " + err.Error(),
				},
			},
		}
		return uuid.Nil, &errorResponse
	}
	return id, nil
}
