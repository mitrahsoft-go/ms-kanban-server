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
				Message:    "Invalid request payload in Handler Layer",
				Details: []response.Details{
					{Field: "body", Message: err.Error()},
				},
			},
		}
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
				Message:    "Validation failed in Handler Layer",
				Details:    details,
			},
		}

		g.JSON(errorResponse.Error.StatusCode, errorResponse)
		return
	}

	if utils.ValidatedPassword(payload.Password) {
		errorResponse := &response.ErrorResponse{
			Success: false,
			Error: response.Error{
				Code:       response.ErrValidation,
				StatusCode: http.StatusBadRequest,
				Message:    "Validation failed in Handler Layer",
				Details: []response.Details{
					{
						Field:   "password",
						Message: "must contain at least one uppercase/lowercase letter, one number and one special character",
					},
				},
			},
		}
		g.JSON(errorResponse.Error.StatusCode, errorResponse)
		return
	}

	err := h.service.SignUp(payload)
	if err != nil {
		g.JSON(err.StatusCode, err)
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
				Message:    "Failed converting json to struct in Handler Layer",
				Details: []response.Details{
					{Field: "body", Message: err.Error()},
				},
			},
		}

		g.JSON(errorResponse.Error.StatusCode, errorResponse)
		return
	}

	validate := validator.New()
	if err := validate.Struct(loginCredentials); err != nil {
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
		g.JSON(errorResponse.Error.StatusCode, errorResponse)
		return
	}

	_, _, err := h.service.SignIn(loginCredentials)
	if err != nil {
		g.JSON(err.StatusCode, err)
		return
	}

	successResponse := &response.SuccessResponse{
		Message:    "Successfully Logged in",
		StatusCode: http.StatusOK,
		Success:    true,
	}

	g.JSON(successResponse.StatusCode, successResponse)
}
