package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofrs/uuid"
	"github.com/ms-kanban-server/config"
	"github.com/ms-kanban-server/internal/handlers/dto"
	"github.com/ms-kanban-server/internal/middleware"
	mail "github.com/ms-kanban-server/internal/pkg/email"
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
	RequestPasswordReset(email string) *response.Error
	ResetPassword(credentials dto.ResetPasswordRequest) *response.Error
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

	if !utils.IsValidPassword(result.PasswordHash, credentials.Password) {
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

		s.logger.Error("Failed to set the expire time in Middleware Layer",
			zap.String("ERROR : ", fmt.Sprintf("%v", err)))
		return nil, err
	}

	refreshExpiresIn, err := utils.StringToInt(config.GetEnv("REFRESH_TOKEN_EXPIRY", "604800"))
	if err != nil {

		s.logger.Error("Failed to set the expire time in Middleware Layer",
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

	if !utils.IsValidPassword(oldToken.TokenHash, credentials.RefreshToken) {
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

		s.logger.Error("Failed to set the expire time in Middleware Layer",
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

func (s *authservice) RequestPasswordReset(email string) *response.Error {
	user, err := s.Repo.RequestPasswordReset(email)
	if err != nil {
		return err
	}
	if user.ID == uuid.Nil {
		s.logger.Warn("Password reset requested for unknown email", zap.String("email", email))
		return &response.Error{Code: response.ErrUnauthorized, StatusCode: http.StatusUnauthorized, Message: "Invalid OTP", Details: []response.Details{{Field: "email", Message: "The provided email does not match a known account"}}}
	}

	otpValue := generateOTP(6)
	otpExpiryMinutes, parseErr := strconv.Atoi(config.GetEnv("OTP_EXPIRY_MINUTES", "15"))
	if parseErr != nil || otpExpiryMinutes <= 0 {
		otpExpiryMinutes = 15
	}
	expiresAt := time.Now().Add(time.Duration(otpExpiryMinutes) * time.Minute)
	hashedOTP, hashErr := utils.HashPassword(otpValue)
	if hashErr != nil {
		return &response.Error{Code: response.ErrInternalServerError, StatusCode: http.StatusInternalServerError, Message: "Failed to secure OTP", Details: []response.Details{{Field: "otp", Message: hashErr.Message}}}
	}

	otpRecord := models.PasswordResetOTP{
		UserID:    user.ID,
		OTPHash:   hashedOTP,
		ExpiresAt: expiresAt,
	}
	if invalidateErr := s.Repo.InvalidatePasswordResetOTPs(user.ID); invalidateErr != nil {
		return invalidateErr
	}
	if saveErr := s.Repo.SavePasswordResetOTP(otpRecord); saveErr != nil {
		return saveErr
	}

	if err := mail.SendPasswordResetOTP(user.Email, otpValue); err != nil {
		s.logger.Error("Failed to send password reset OTP", zap.String("email", user.Email), zap.Error(err))
		return &response.Error{Code: response.ErrInternalServerError, StatusCode: http.StatusInternalServerError, Message: "Failed to send password reset OTP", Details: []response.Details{{Field: "email", Message: err.Error()}}}
	}

	s.logger.Info("Password reset OTP generated", zap.String("email", user.Email))
	return nil
}

func generateOTP(length int) string {
	chars := "0123456789"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		result[i] = chars[n.Int64()]
	}
	return string(result)
}

func (s *authservice) ResetPassword(credentials dto.ResetPasswordRequest) *response.Error {
	if utils.ValidatedPassword(credentials.NewPassword) {
		return &response.Error{Code: response.ErrBadRequest, StatusCode: http.StatusBadRequest, Message: "BadRequest", Details: []response.Details{{Field: "new_password", Message: "Password must meet the minimum complexity requirements"}}}
	}

	user, err := s.Repo.RequestPasswordReset(credentials.Email)
	if err != nil {
		return err
	}
	if user.ID == uuid.Nil {
		return &response.Error{Code: response.ErrUnauthorized, StatusCode: http.StatusUnauthorized, Message: "Invalid OTP", Details: []response.Details{{Field: "email", Message: "The provided email does not match a known account"}}}
	}

	otpRecord, otpErr := s.Repo.GetPasswordResetOTP(user.ID, credentials.OTP)
	if otpErr != nil {
		return otpErr
	}
	if otpRecord.ExpiresAt.Before(time.Now()) || otpRecord.UsedAt != nil || !utils.IsValidPassword(otpRecord.OTPHash, credentials.OTP) {
		return &response.Error{Code: response.ErrUnauthorized, StatusCode: http.StatusUnauthorized, Message: "Invalid OTP", Details: []response.Details{{Field: "otp", Message: "The provided OTP is invalid or expired"}}}
	}

	passwordHash, hashErr := utils.HashPassword(credentials.NewPassword)
	if hashErr != nil {
		return hashErr
	}
	if updateErr := s.Repo.UpdateUserPassword(user.ID, passwordHash); updateErr != nil {
		return updateErr
	}
	if revokeErr := s.Repo.RevokeRefreshTokens(user.ID); revokeErr != nil {
		return revokeErr
	}

	usedAt := time.Now()
	otpRecord.UsedAt = &usedAt
	if saveErr := s.Repo.SavePasswordResetOTP(otpRecord); saveErr != nil {
		return saveErr
	}

	s.logger.Info("Password reset completed", zap.String("email", credentials.Email))
	return nil
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
		return &response.Error{
			Code:       response.ErrBadRequest,
			StatusCode: errorResponse.StatusCode,
			Message:    "BadRequest",
			Details: []response.Details{
				{
					Field:   "Password",
					Message: "Invalid password format",
				},
			},
		}
	}

	result := models.User{
		Email:        credentials.Email,
		PasswordHash: passwordhash,
		Role:         string(credentials.Role),
		FullName:     credentials.FullName,
		AvatarURL:    credentials.AvatarURL,
		Timezone:     credentials.Timezone,
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

	organizationID, errorResponse := utils.StringToUUID(credentials.OrganizationID)
	if errorResponse != nil {
		return errorResponse
	}

	if organizationID != uuid.Nil {
		result.OrganizationID = &organizationID
	}

	return s.Repo.SignUp(result)

}
