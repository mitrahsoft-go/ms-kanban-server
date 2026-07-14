package repository

import (
	"errors"
	"net/http"

	"github.com/ms-kanban-server/internal/pkg/models"
	"github.com/ms-kanban-server/internal/pkg/response"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Repository interface {
	SignIn(email string) (models.User, int, *response.Error)
	SignUp(row models.User) (int, *response.Error)
}

func InitAuthRepository(db *gorm.DB, logger *zap.Logger) Repository {
	return &authdatabase{
		DB:     db,
		logger: logger,
	}
}

type authdatabase struct {
	DB     *gorm.DB
	logger *zap.Logger
}

func (d *authdatabase) SignIn(email string) (models.User, int, *response.Error) {

	var row models.User

	err := d.DB.Preload("Role").Where("email = ?", email).First(&row).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			errorResponse := response.Error{
				Code:    response.ErrUnauthorized,
				Message: "Enter valid Email or Password before login",
				Details: []response.Details{
					{
						Field:   "Email/Password",
						Message: "User not found :" + email,
					},
				},
			}
			d.logger.Error("User not found in database in Repository layer",
				zap.String("Email", email), zap.Error(err))
			return models.User{}, http.StatusUnauthorized, &errorResponse
		}
		errorResponse := response.Error{
			Code:    response.ErrInternalServerError,
			Message: "InternalServerError",
			Details: []response.Details{
				{
					Field:   "Database",
					Message: "Database error : " + err.Error(),
				},
			},
		}
		d.logger.Error("Database error occurred in Repository layer",
			zap.String("Email", email), zap.Error(err))
		return models.User{}, http.StatusInternalServerError, &errorResponse
	}

	return row, http.StatusOK, nil
}

func (d *authdatabase) SignUp(row models.User) (int, *response.Error) {

	if err := d.DB.Create(&row).Error; err != nil {
		errorResponse := response.Error{
			Code:    response.ErrInternalServerError,
			Message: "Failed to Register",
			Details: []response.Details{
				{
					Field:   "Database",
					Message: "Failed inserting the row : " + err.Error(),
				},
			},
		}
		d.logger.Error("Database error occurred in Repository layer",
			zap.Error(err))
		return http.StatusInternalServerError, &errorResponse
	}

	return http.StatusCreated, nil
}
