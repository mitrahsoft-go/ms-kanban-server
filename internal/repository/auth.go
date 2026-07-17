package repository

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/ms-kanban-server/internal/pkg/models"
	"github.com/ms-kanban-server/internal/pkg/response"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository interface {
	SignIn(email string) (models.User, *response.Error)
	SignInByID(id uuid.UUID) (models.User, *response.Error)
	SignUp(row models.User) *response.Error
	StoreRefreshToken(token models.RefreshToken) *response.Error
	GetRefreshToken(userID string) (models.RefreshToken, *response.Error)
	ChangePassword(password string, userID uuid.UUID) *response.Error
	UpdateUser(userID uuid.UUID, req models.User) *response.Error
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
				Message:    "Enter valid Email/Password",
				Details: []response.Details{
					{
						Field:   "Email/Password",
						Message: "User not found :" + email,
					},
				},
			}
			d.logger.Error("User not found in database",
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

		d.logger.Error("Database error occurred",
			zap.String("Email", email), zap.Error(err))
		return models.User{}, &errorResponse
	}

	return row, nil
}

func (d *authdatabase) SignInByID(id uuid.UUID) (models.User, *response.Error) {

	var row models.User

	if err := d.DB.Where("id = ?", id).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {

			d.logger.Error("The user associated with the refresh token could not be found",
				zap.Error(err))
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

		d.logger.Error("Failed to retrieve user",
			zap.Error(err))
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

		d.logger.Error("Database error occurred",
			zap.Error(err))
		return &errorResponse
	}

	return nil
}

func (d *authdatabase) StoreRefreshToken(token models.RefreshToken) *response.Error {

	err := d.DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "user_id"}, // Conflict target
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"token_hash",
			"user_agent",
			"ip_address",
			"expires_at",
			"revoked_at",
			"updated_at", // if your model has UpdatedAt
		}),
	}).Create(&token).Error

	if err != nil {
		errorResponse := response.Error{
			Code:       response.ErrInternalServerError,
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to store refresh token",
			Details: []response.Details{{
				Message: "Failed storing refresh token: " + err.Error(),
			}},
		}

		d.logger.Error("Database error occurred while storing refresh token",
			zap.Error(err))

		return &errorResponse
	}

	return nil
}

func (d *authdatabase) GetRefreshToken(userID string) (models.RefreshToken, *response.Error) {

	var token models.RefreshToken

	if err := d.DB.Where("user_id = ?", userID).First(&token).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {

			d.logger.Error("Database error occurred while storing refresh token",
				zap.Error(err))
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

		d.logger.Error("Database error occurred while storing refresh token",
			zap.Error(err))
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

func (d *authdatabase) ChangePassword(password string, userID uuid.UUID) *response.Error {

	if err := d.DB.
		Model(&models.User{}).
		Where("id = ?", userID).
		Update("password_hash", password).Error; err != nil {

		d.logger.Error("Database error occurred while updating user password",
			zap.Error(err))

		return &response.Error{
			Code:       response.ErrInternalServerError,
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to update password",
			Details: []response.Details{{
				Message: "Failed updating password: " + err.Error(),
			}},
		}
	}

	return nil
}

func (d *authdatabase) UpdateUser(userID uuid.UUID, req models.User) *response.Error {

	result := d.DB.
		Model(&models.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"full_name":  req.FullName,
			"username":   req.UserName,
			"avatar_url": req.AvatarURL,
			"timezone":   req.Timezone,
		})

	if result.Error != nil {

		d.logger.Error("Database error occurred while updating user",
			zap.Error(result.Error))

		return &response.Error{
			Code:       response.ErrInternalServerError,
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to update user",
			Details: []response.Details{{
				Message: "Failed updating user: " + result.Error.Error(),
			}},
		}
	}

	if result.RowsAffected == 0 {

		d.logger.Error("User not found while updating user",
			zap.String("user_id", fmt.Sprint(userID)))

		return &response.Error{
			Code:       response.ErrUnauthorized,
			StatusCode: http.StatusUnauthorized,
			Message:    "User not found",
			Details: []response.Details{{
				Field:   "user_id",
				Message: "The specified user does not exist",
			}},
		}
	}

	return nil
}
