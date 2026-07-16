package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/ms-kanban-server/internal/handlers/dto"
	"github.com/ms-kanban-server/internal/pkg/response"
	"github.com/ms-kanban-server/internal/pkg/utils"
	"github.com/ms-kanban-server/internal/services"
	"go.uber.org/zap"
)

func InitAuthHandler(service services.Service, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		service: service,
		logger:  logger,
	}
}

type AuthHandler struct {
	service services.Service
	logger  *zap.Logger
}

func (h *AuthHandler) SignUp(g *gin.Context) {

	var payload dto.SignUpRequest

	if err := g.Bind(&payload); err != nil {
		errorResponse := &response.ErrorResponse{
			Success: false,
			Error: response.Error{
				Code:       response.ErrBadRequest,
				StatusCode: http.StatusBadRequest,
				Message:    "Invalid request payload",
				Details: []response.Details{
					{
						Field:   "body",
						Message: err.Error()},
				},
			},
		}

		h.logger.Error("Invalid request payload in Handler Layer",
			zap.Error(err))

		g.JSON(errorResponse.Error.StatusCode, errorResponse)
		return
	}

	validate := validator.New()
	if err := validate.Struct(payload); err != nil {
		var details []response.Details
		for _, fieldErr := range err.(validator.ValidationErrors) {
			details = append(details, response.Details{
				Field:   fieldErr.Field(),
				Message: fmt.Sprintf("failed on '%s' validation in Handler Layer", fieldErr.Tag()),
			})
		}

		errorResponse := &response.ErrorResponse{
			Success: false,
			Error: response.Error{
				Code:       response.ErrValidation,
				StatusCode: http.StatusBadRequest,
				Message:    "Validation failed",
				Details:    details,
			},
		}

		h.logger.Error("Validation failed in Handler Layer",
			zap.Error(err))

		g.JSON(errorResponse.Error.StatusCode, errorResponse)
		return
	}

	if utils.ValidatedPassword(payload.Password) {
		errorResponse := &response.ErrorResponse{
			Success: false,
			Error: response.Error{
				Code:       response.ErrValidation,
				StatusCode: http.StatusBadRequest,
				Message:    "Validation failed",
				Details: []response.Details{
					{
						Field:   "password",
						Message: "must contain at least one uppercase/lowercase letter, one number and one special character",
					},
				},
			},
		}

		h.logger.Error("Validation failed for password in Handler Layer")
		g.JSON(errorResponse.Error.StatusCode, errorResponse)
		return
	}

	err := h.service.SignUp(payload)
	if err != nil {
		errorResponse := &response.ErrorResponse{
			Success: false,
			Error:   *err,
		}
		g.JSON(err.StatusCode, errorResponse)
		return
	}

	successResponse := &response.SuccessResponse{
		Message:    "Successfully Created",
		StatusCode: http.StatusCreated,
		Success:    true,
	}

	g.JSON(successResponse.StatusCode, successResponse)

}

func (h *AuthHandler) SignIn(g *gin.Context) {

	var loginCredentials dto.SignInRequest

	if err := g.ShouldBindJSON(&loginCredentials); err != nil {
		errorResponse := &response.ErrorResponse{
			Success: false,
			Error: response.Error{
				Code:       response.ErrBadRequest,
				StatusCode: http.StatusBadRequest,
				Message:    "Failed converting json to struct",
				Details: []response.Details{
					{
						Field:   "body",
						Message: err.Error()},
				},
			},
		}

		h.logger.Error("Invalid request payload in Handler Layer",
			zap.Error(err))

		g.JSON(errorResponse.Error.StatusCode, errorResponse)
		return
	}

	validate := validator.New()
	if err := validate.Struct(loginCredentials); err != nil {
		var details []response.Details
		for _, fieldErr := range err.(validator.ValidationErrors) {
			details = append(details, response.Details{
				Field:   fieldErr.Field(),
				Message: fmt.Sprintf("failed on '%s' validation", fieldErr.Tag()),
			})
		}

		errorResponse := &response.ErrorResponse{
			Success: false,
			Error: response.Error{
				Code:       response.ErrValidation,
				StatusCode: http.StatusBadRequest,
				Message:    "Validation failed",
				Details:    details,
			},
		}

		h.logger.Error("Validation failed in Handler Layer",
			zap.Error(err))

		g.JSON(errorResponse.Error.StatusCode, errorResponse)
		return
	}

	tokens, err := h.service.SignIn(loginCredentials)
	if err != nil {
		errorResponse := &response.ErrorResponse{
			Success: false,
			Error:   *err,
		}
		g.JSON(err.StatusCode, errorResponse)
		return
	}

	successResponse := &response.SuccessResponse{
		Message:    "Successfully Logged in",
		StatusCode: http.StatusOK,
		Success:    true,
		Data:       tokens,
	}

	g.JSON(successResponse.StatusCode, successResponse)
}

func (h *AuthHandler) RequestPasswordReset(g *gin.Context) {
	var payload dto.PasswordResetRequest
	if err := g.ShouldBindJSON(&payload); err != nil {
		h.logger.Error("Invalid request payload in Handler Layer", zap.Error(err))
		g.JSON(http.StatusBadRequest, &response.ErrorResponse{Success: false, Error: response.Error{Code: response.ErrBadRequest, StatusCode: http.StatusBadRequest, Message: "Invalid request payload", Details: []response.Details{{Field: "body", Message: err.Error()}}}})
		return
	}

	if err := validator.New().Struct(payload); err != nil {
		g.JSON(http.StatusBadRequest, &response.ErrorResponse{Success: false, Error: response.Error{Code: response.ErrValidation, StatusCode: http.StatusBadRequest, Message: "Validation failed", Details: []response.Details{{Field: "email", Message: "must be a valid email"}}}})
		return
	}

	if err := h.service.RequestPasswordReset(payload.Email); err != nil {
		g.JSON(err.StatusCode, &response.ErrorResponse{Success: false, Error: *err})
		return
	}

	g.JSON(http.StatusOK, &response.SuccessResponse{Success: true, StatusCode: http.StatusOK, Message: "A password reset OTP has been sent to your email address"})
}

func (h *AuthHandler) ResetPassword(g *gin.Context) {
	var payload dto.ResetPasswordRequest
	if err := g.ShouldBindJSON(&payload); err != nil {
		h.logger.Error("Invalid request payload in Handler Layer", zap.Error(err))
		g.JSON(http.StatusBadRequest, &response.ErrorResponse{Success: false, Error: response.Error{Code: response.ErrBadRequest, StatusCode: http.StatusBadRequest, Message: "Invalid request payload", Details: []response.Details{{Field: "body", Message: err.Error()}}}})
		return
	}

	if err := validator.New().Struct(payload); err != nil {
		g.JSON(http.StatusBadRequest, &response.ErrorResponse{Success: false, Error: response.Error{Code: response.ErrValidation, StatusCode: http.StatusBadRequest, Message: "Validation failed", Details: []response.Details{{Field: "otp", Message: "OTP is required"}}}})
		return
	}

	if utils.ValidatedPassword(payload.NewPassword) {
		g.JSON(http.StatusBadRequest, &response.ErrorResponse{Success: false, Error: response.Error{Code: response.ErrValidation, StatusCode: http.StatusBadRequest, Message: "Validation failed", Details: []response.Details{{Field: "new_password", Message: "must contain at least one uppercase/lowercase letter, one number and one special character"}}}})
		return
	}

	if err := h.service.ResetPassword(payload); err != nil {
		g.JSON(err.StatusCode, &response.ErrorResponse{Success: false, Error: *err})
		return
	}

	g.JSON(http.StatusOK, &response.SuccessResponse{Success: true, StatusCode: http.StatusOK, Message: "Password reset successfully"})
}

func (h *AuthHandler) RefreshToken(g *gin.Context) {

	var payload dto.RefreshTokenRequest

	if err := g.ShouldBindJSON(&payload); err != nil {
		errorResponse := &response.ErrorResponse{
			Success: false,
			Error: response.Error{
				Code:       response.ErrBadRequest,
				StatusCode: http.StatusBadRequest,
				Message:    "Invalid request payload in Handler Layer",
				Details: []response.Details{{
					Field:   "body",
					Message: err.Error(),
				}},
			},
		}

		h.logger.Error("Invalid request payload in Handler Layer",
			zap.Error(err))

		g.JSON(errorResponse.Error.StatusCode, errorResponse)
		return
	}

	validate := validator.New()
	if err := validate.Struct(payload); err != nil {
		var details []response.Details
		for _, fieldErr := range err.(validator.ValidationErrors) {
			details = append(details, response.Details{
				Field:   fieldErr.Field(),
				Message: fmt.Sprintf("failed on '%s' validation in Handler Layer", fieldErr.Tag()),
			})
		}
		errorResponse := &response.ErrorResponse{
			Success: false,
			Error: response.Error{
				Code:       response.ErrValidation,
				StatusCode: http.StatusBadRequest,
				Message:    "Validation failed",
				Details:    details,
			},
		}

		h.logger.Error("Validation failed in Handler Layer",
			zap.Error(err))

		g.JSON(errorResponse.Error.StatusCode, errorResponse)
		return
	}

	userID, exist := g.Get("user_id")
	if !exist {
		errorResponse := &response.ErrorResponse{
			Success: false,
			Error: response.Error{
				Code:       response.ErrValidation,
				StatusCode: http.StatusBadRequest,
				Message:    "invalid User ID",
				Details: []response.Details{
					{
						Field:   "User ID",
						Message: "User Id missing",
					},
				},
			},
		}

		h.logger.Error("User Id missing  in Handler Layer")

		g.JSON(errorResponse.Error.StatusCode, errorResponse)
		return
	}
	payload.UserID = userID.(string)

	tokens, err := h.service.RefreshToken(payload)
	if err != nil {
		errorResponse := &response.ErrorResponse{
			Success: false,
			Error:   *err,
		}
		g.JSON(err.StatusCode, errorResponse)
		return
	}

	successResponse := &response.SuccessResponse{
		Message:    "Token refreshed successfully",
		StatusCode: http.StatusOK,
		Success:    true,
		Data:       tokens,
	}
	g.JSON(successResponse.StatusCode, successResponse)
}
