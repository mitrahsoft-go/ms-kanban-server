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
	SignIn(email string) (models.User, *response.Error)
	SignInByID(id uint) (models.User, *response.Error)
	SignUp(row models.User) *response.Error
	StoreRefreshToken(token models.RefreshToken) *response.Error
	GetRefreshToken(userID string) (models.RefreshToken, *response.Error)
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

func (d *authdatabase) SignIn(email string) (models.User, *response.Error) {

	var row models.User

	err := d.DB.Where("email = ?", email).First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			errorResponse := response.Error{
				Code:       response.ErrUnauthorized,
				StatusCode: http.StatusUnauthorized,
				Message:    "Enter valid Email or Password before login",
				Details: []response.Details{
					{
						Field:   "Email/Password",
						Message: "User not found :" + email,
					},
				},
			}
			d.logger.Error("User not found in database in Repository layer",
				zap.String("Email", email), zap.Error(err))
			return models.User{}, &errorResponse
		}

		errorResponse := response.Error{
			Code:       response.ErrInternalServerError,
			StatusCode: http.StatusInternalServerError,
			Message:    "InternalServerError",
			Details: []response.Details{
				{
					Message: "Failed to Login : " + err.Error(),
				},
			},
		}

		d.logger.Error("Database error occurred in Repository layer",
			zap.String("Email", email), zap.Error(err))
		return models.User{}, &errorResponse
	}

	return row, nil
}

func (d *authdatabase) SignInByID(id uint) (models.User, *response.Error) {

	var row models.User
	
	if err := d.DB.Where("id = ?", id).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.User{}, &response.Error{
				Code:       response.ErrUnauthorized,
				StatusCode: http.StatusUnauthorized,
				Message:    "User not found",
				Details: []response.Details{{
					Field:   "user",
					Message: "The user associated with the refresh token could not be found",
				}},
			}
		}
		return models.User{}, &response.Error{
			Code:       response.ErrInternalServerError,
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to retrieve user",
			Details: []response.Details{{
				Message: "Failed querying user : " + err.Error(),
			}},
		}
	}
	return row, nil
}

func (d *authdatabase) SignUp(row models.User) *response.Error {

	if err := d.DB.Create(&row).Error; err != nil {
		errorResponse := response.Error{
			Code:       response.ErrInternalServerError,
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to Register",
			Details: []response.Details{
				{
					Message: "Failed inserting the row : " + err.Error(),
				},
			},
		}

		d.logger.Error("Database error occurred in Repository layer",
			zap.Error(err))
		return &errorResponse
	}

	return nil
}

func (d *authdatabase) StoreRefreshToken(token models.RefreshToken) *response.Error {

	if err := d.DB.Create(&token).Error; err != nil {
		errorResponse := response.Error{
			Code:       response.ErrInternalServerError,
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to store refresh token",
			Details: []response.Details{{
				Message: "Failed inserting refresh token : " + err.Error(),
			}},
		}
		d.logger.Error("Database error occurred while storing refresh token in Repository layer",
			zap.Error(err))
		return &errorResponse
	}

	return nil
}

func (d *authdatabase) GetRefreshToken(userID string) (models.RefreshToken, *response.Error) {

	var token models.RefreshToken

	if err := d.DB.Where("userID = ?", userID).First(&token).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.RefreshToken{}, &response.Error{
				Code:       response.ErrUnauthorized,
				StatusCode: http.StatusUnauthorized,
				Message:    "Invalid refresh token",
				Details: []response.Details{{
					Field:   "refresh_token",
					Message: "The refresh token was not found",
				}},
			}
		}
		return models.RefreshToken{}, &response.Error{
			Code:       response.ErrInternalServerError,
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to read refresh token",
			Details: []response.Details{{
				Message: "Failed querying refresh token : " + err.Error(),
			}},
		}
	}

	return token, nil
}
