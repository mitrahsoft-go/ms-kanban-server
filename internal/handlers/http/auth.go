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

func (h *AuthHandler) Logout(g *gin.Context) {

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
	id := userID.(string)

	err := h.service.Logout(id)
	if err != nil {
		errorResponse := &response.ErrorResponse{
			Success: false,
			Error:   *err,
		}
		g.JSON(err.StatusCode, errorResponse)
		return
	}

	successResponse := &response.SuccessResponse{
		Message:    "Logedout successfully",
		StatusCode: http.StatusOK,
		Success:    true,
	}
	g.JSON(successResponse.StatusCode, successResponse)
}

func (h *AuthHandler) ChangePassword(g *gin.Context) {

	var payload dto.ChangePasswordRequest

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
	userIDStr := userID.(string)

	id, errorResponse := utils.StringToUUID(userIDStr)
	if errorResponse != nil {
		h.logger.Error("Failed to convert the string into UUID in Handler layer")
		return
	}
	payload.UserID = id

	err := h.service.ChangePassword(payload)
	if err != nil {
		errorResponse := &response.ErrorResponse{
			Success: false,
			Error:   *err,
		}
		g.JSON(err.StatusCode, errorResponse)
		return
	}

	successResponse := &response.SuccessResponse{
		Message:    "Password Changed successfully",
		StatusCode: http.StatusOK,
		Success:    true,
	}
	g.JSON(successResponse.StatusCode, successResponse)

}
