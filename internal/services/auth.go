package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofrs/uuid"
	"github.com/ms-kanban-server/config"
	"github.com/ms-kanban-server/internal/handlers/dto"
	"github.com/ms-kanban-server/internal/middleware"
	"github.com/ms-kanban-server/internal/pkg/models"
	"github.com/ms-kanban-server/internal/pkg/response"
	"github.com/ms-kanban-server/internal/pkg/utils"
	"github.com/ms-kanban-server/internal/repository"
	"go.uber.org/zap"
)

type Service interface {
	SignIn(credentials dto.SignInRequest) (*dto.AuthTokensResponse, *response.Error)
	RefreshToken(credentials dto.RefreshTokenRequest) (*dto.AuthTokensResponse, *response.Error)
	SignUp(credentials dto.SignUpRequest) *response.Error
	Logout(UserID string) *response.Error
	ChangePassword(payload dto.ChangePasswordRequest) *response.Error
	UpdateUser(payload dto.UpdateUserRequest, userID uuid.UUID) *response.Error
}

func InitAuthService(repo repository.Repository, logger *zap.Logger) Service {
	return &authservice{
		Repo:   repo,
		logger: logger,
	}
}

type authservice struct {
	Repo   repository.Repository
	logger *zap.Logger
}

func (s *authservice) SignIn(credentials dto.SignInRequest) (*dto.AuthTokensResponse, *response.Error) {

	result, err := s.Repo.SignIn(credentials.Email)
	if err != nil {
		s.logger.Warn("Login failed during user lookup",
			zap.String("email", credentials.Email),
			zap.String("error", err.Message))
		return nil, err
	}

	if !result.IsActive {
		s.logger.Warn("Login rejected for inactive user",
			zap.String("email", credentials.Email))
		return nil, &response.Error{
			Code:       response.ErrForbidden,
			StatusCode: http.StatusForbidden,
			Message:    "Account is inactive",
			Details: []response.Details{{
				Field:   "account",
				Message: "The account is deactivated or locked",
			}},
		}
	}

	if utils.IsValidPassword(result.PasswordHash, credentials.Password) {
		s.logger.Warn("Login failed due to invalid credentials",
			zap.String("email", credentials.Email))
		return nil, &response.Error{
			Code:       response.ErrUnauthorized,
			StatusCode: http.StatusUnauthorized,
			Message:    "Email or password is incorrect",
			Details: []response.Details{{
				Field:   "credentials",
				Message: "The provided email or password is invalid",
			}},
		}
	}

	accessToken, tokenErr := middleware.GenerateJWT(result.Role, result.ID, s.logger)
	if tokenErr != nil {
		return nil, tokenErr
	}

	refreshTokenValue, refreshTokenErr := generateRefreshTokenValue()
	if refreshTokenErr != nil {
		return nil, &response.Error{
			Code:       response.ErrInternalServerError,
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to create refresh token",
			Details: []response.Details{{
				Field:   "refresh_token",
				Message: refreshTokenErr.Error(),
			}},
		}
	}

	hashedRefreshToken, hashErr := utils.HashPassword(refreshTokenValue)
	if hashErr != nil {
		return nil, hashErr
	}

	expiresIn, err := utils.StringToInt(config.GetEnv("JWT_EXPIRY", "900"))
	if err != nil {

		s.logger.Error("Failed to set the expire time in service Layer",
			zap.String("ERROR : ", fmt.Sprintf("%v", err)))
		return nil, err
	}

	refreshExpiresIn, err := utils.StringToInt(config.GetEnv("REFRESH_TOKEN_EXPIRY", "604800"))
	if err != nil {

		s.logger.Error("Failed to set the expire time in service Layer",
			zap.String("ERROR : ", fmt.Sprintf("%v", err)))
		return nil, err
	}
	expiresAt := time.Now().Add(time.Duration(refreshExpiresIn) * time.Second)
	storeErr := s.Repo.StoreRefreshToken(models.RefreshToken{
		UserID:    result.ID,
		TokenHash: hashedRefreshToken,
		ExpiresAt: expiresAt,
	})
	if storeErr != nil {
		return nil, storeErr
	}

	return &dto.AuthTokensResponse{
		AccessToken:      accessToken,
		RefreshToken:     refreshTokenValue,
		TokenType:        "Bearer",
		ExpiresIn:        expiresIn,
		RefreshExpiresIn: refreshExpiresIn,
	}, nil
}

func generateRefreshTokenValue() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (s *authservice) RefreshToken(credentials dto.RefreshTokenRequest) (*dto.AuthTokensResponse, *response.Error) {

	oldToken, err := s.Repo.GetRefreshToken(credentials.UserID)
	if err != nil {
		return nil, err
	}

	if utils.IsValidPassword(oldToken.TokenHash, credentials.RefreshToken) {
		return nil, &response.Error{
			Code:       response.ErrUnauthorized,
			StatusCode: http.StatusUnauthorized,
			Message:    "Wrong Refresh",
			Details: []response.Details{{
				Field:   "refresh_token",
				Message: "The refresh token is wrong, Give correct refresh token",
			}},
		}
	}

	if time.Now().After(oldToken.ExpiresAt) {
		return nil, &response.Error{
			Code:       response.ErrUnauthorized,
			StatusCode: http.StatusUnauthorized,
			Message:    "Refresh token expired",
			Details: []response.Details{{
				Field:   "refresh_token",
				Message: "The refresh token has expired",
			}},
		}
	}

	user, userErr := s.Repo.SignInByID(oldToken.UserID)
	if userErr != nil {
		return nil, userErr
	}
	if !user.IsActive {
		return nil, &response.Error{
			Code:       response.ErrForbidden,
			StatusCode: http.StatusForbidden,
			Message:    "Account is inactive",
			Details: []response.Details{{
				Field:   "account",
				Message: "The account is deactivated or locked",
			}},
		}
	}

	accessToken, tokenErr := middleware.GenerateJWT(user.Role, user.ID, s.logger)
	if tokenErr != nil {
		return nil, tokenErr
	}

	expiresIn, err := utils.StringToInt(config.GetEnv("JWT_EXPIRY", "900"))
	if err != nil {

		s.logger.Error("Failed to set the expire time in service Layer",
			zap.String("ERROR : ", fmt.Sprintf("%v", err)))
		return nil, err
	}

	return &dto.AuthTokensResponse{
		AccessToken:      accessToken,
		RefreshToken:     credentials.RefreshToken,
		TokenType:        "Bearer",
		ExpiresIn:        expiresIn,
		RefreshExpiresIn: int(time.Until(oldToken.ExpiresAt).Seconds()),
	}, nil
}

func (s *authservice) SignUp(credentials dto.SignUpRequest) *response.Error {

	validate := validator.New()
	err := validate.Struct(credentials)
	if err != nil {
		s.logger.Error(" Validated failure in Email/Password before login in service layer",
			zap.String("Email", credentials.Email), zap.Error(err))
		return &response.Error{
			Code:       response.ErrBadRequest,
			StatusCode: http.StatusBadRequest,
			Message:    "BadRequest",
			Details: []response.Details{
				{
					Field:   "Email/Password",
					Message: "Invalid email or password format",
				},
			},
		}
	}

	if utils.ValidatedPassword(credentials.Password) {
		s.logger.Error("Validated failure in Password before login in service layer",
			zap.String("Email", credentials.Email), zap.Error(err))
		return &response.Error{
			Code:       response.ErrBadRequest,
			StatusCode: http.StatusBadRequest,
			Message:    "BadRequest",
			Details: []response.Details{
				{
					Field:   "Password",
					Message: "Invalid password format",
				},
			},
		}

	}

	passwordhash, errorResponse := utils.HashPassword(credentials.Password)
	if errorResponse != nil {
		s.logger.Error("Failed Hashing Password before login in service layer",
			zap.String("Email", credentials.Email), zap.Error(err))
		return errorResponse
	}

	role := dto.Role(credentials.Role)

	if err := role.Validate(); err != nil {
		s.logger.Error(" Invalid role  in service layer",
			zap.String("Role", string(credentials.Role)), zap.Error(err))
		return &response.Error{
			Code:       response.ErrBadRequest,
			StatusCode: http.StatusBadRequest,
			Message:    "BadRequest",
			Details: []response.Details{
				{
					Field:   "Role",
					Message: "Invalid role",
				},
			},
		}
	}

	result := models.User{
		Email:        credentials.Email,
		PasswordHash: passwordhash,
		Role:         string(credentials.Role),
		FullName:     credentials.FullName,
		UserName:     credentials.UserName,
		AvatarURL:    credentials.AvatarURL,
		Timezone:     credentials.Timezone,
	}

	organizationID, errorResponse := utils.StringToUUID(credentials.OrganizationID)
	if errorResponse != nil {
		s.logger.Error("Failed to convert the string into UUID in service layer",
			zap.String("Email", credentials.Email), zap.Error(err))
		return errorResponse
	}

	if organizationID != uuid.Nil {
		result.OrganizationID = &organizationID
	}

	return s.Repo.SignUp(result)

}

func (s *authservice) Logout(UserID string) *response.Error {

	oldToken, err := s.Repo.GetRefreshToken(UserID)
	if err != nil {
		return err
	}

	expiresAt := time.Now()

	s.Repo.StoreRefreshToken(models.RefreshToken{
		UserID:    oldToken.ID,
		ExpiresAt: expiresAt,
	})

	return nil
}

func (s *authservice) ChangePassword(payload dto.ChangePasswordRequest) *response.Error {

	result, err := s.Repo.SignInByID(payload.UserID)
	if err != nil {
		s.logger.Warn("Login failed during user lookup",
			zap.String("error", err.Message))
		return err
	}

	if utils.IsValidPassword(result.PasswordHash, payload.OldPassword) {
		s.logger.Warn("Login failed due to invalid credentials")
		return &response.Error{
			Code:       response.ErrUnauthorized,
			StatusCode: http.StatusUnauthorized,
			Message:    "password is incorrect",
			Details: []response.Details{{
				Field:   "Password",
				Message: "The provided password is invalid",
			}},
		}
	}

	if utils.ValidatedPassword(payload.NewPassword) {
		s.logger.Error("Validated failure in Password before login in service layer")
		return &response.Error{
			Code:       response.ErrBadRequest,
			StatusCode: http.StatusBadRequest,
			Message:    "BadRequest",
			Details: []response.Details{
				{
					Field:   "Password",
					Message: "Invalid password format",
				},
			},
		}

	}

	passwordhash, errorResponse := utils.HashPassword(payload.NewPassword)
	if errorResponse != nil {
		s.logger.Error("Failed Hashing Password before login in service layer",
			zap.String("Email", result.Email))
		return errorResponse
	}

	return s.Repo.ChangePassword(passwordhash, payload.UserID)

}

func (s *authservice) UpdateUser(payload dto.UpdateUserRequest, userID uuid.UUID) *response.Error {

	if len(payload.FullName) >30 {
		s.logger.Error("Validated failure in Full Name  in service layer")
		return &response.Error{
			Code:       response.ErrBadRequest,
			StatusCode: http.StatusBadRequest,
			Message:    "BadRequest",
			Details: []response.Details{
				{
					Field:   "Full Name",
					Message: "Invalid Full Name format",
				},
			},
		}
	}

	if len(payload.UserName) > 30 {
		s.logger.Error("Validated failure in Full Name  in service layer")
		return &response.Error{
			Code:       response.ErrBadRequest,
			StatusCode: http.StatusBadRequest,
			Message:    "BadRequest",
			Details: []response.Details{
				{
					Field:   "UserName",
					Message: "Invalid UserName format",
				},
			},
		}
	}

	req := models.User{
		FullName:  payload.FullName,
		UserName:  payload.UserName,
		AvatarURL: payload.AvatarURL,
		Timezone:  payload.Timezone,
	}

	return s.Repo.UpdateUser(userID, req)

}
