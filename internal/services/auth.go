package services

import (
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gofrs/uuid"
	"github.com/ms-kanban-server/internal/handlers/dto"
	"github.com/ms-kanban-server/internal/pkg/models"
	"github.com/ms-kanban-server/internal/pkg/response"
	"github.com/ms-kanban-server/internal/pkg/utils"
	"github.com/ms-kanban-server/internal/repository"
	"go.uber.org/zap"
)

type Service interface {
	SignIn(credentials dto.SignInRequest) (uuid.UUID, int, string, *response.Error)
	SignUp(credentials dto.SignUpRequest) (int, *response.Error)
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

func (s *authservice) SignIn(credentials dto.SignInRequest) (uuid.UUID, int, string, *response.Error) {

	result, code, err := s.Repo.SignIn(credentials.Email)
	if err != nil {
		s.logger.Error("Error occurred during SignIn in service layer",
			zap.String("Email", credentials.Email))
		return uuid.Nil, code, "", err
	}

	if utils.IsValidPassword(result.PasswordHash, credentials.Password) {

		s.logger.Error(" Validated failure in Password before login in service layer",
			zap.String("Email", credentials.Email))
		return uuid.Nil, code, "", &response.Error{
			Code:    response.ErrUnauthorized,
			Message: "Enter valid Email or Password before login",
			Details: []response.Details{
				{
					Field:   "Email/Password",
					Message: "User not found :" + credentials.Email,
				},
			},
		}
	}
	return result.ID, http.StatusOK, string(result.Role), nil
}

func (s *authservice) SignUp(credentials dto.SignUpRequest) (int, *response.Error) {

	validate := validator.New()
	err := validate.Struct(credentials)
	if err != nil {
		s.logger.Error(" Validated failure in Email/Password before login in service layer",
			zap.String("Email", credentials.Email), zap.Error(err))
		return http.StatusBadRequest, &response.Error{
			Code:    response.ErrBadRequest,
			Message: "BadRequest",
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
		return http.StatusBadRequest, &response.Error{
			Code:    response.ErrBadRequest,
			Message: "BadRequest",
			Details: []response.Details{
				{
					Field:   "Password",
					Message: "Invalid password format",
				},
			},
		}

	}

	code, passwordhash, errorResponse := utils.HashPassword(credentials.Password)
	if errorResponse != nil {
		s.logger.Error("Error occurred while hashing password in service layer",
			zap.String("Email", credentials.Email), zap.Error(fmt.Errorf("%v", errorResponse)))
		return code, &response.Error{
			Code:    response.ErrBadRequest,
			Message: "BadRequest",
			Details: []response.Details{
				{
					Field:   "Password",
					Message: "Invalid password format",
				},
			},
		}
	}

	organizationID, errorResponse := utils.StringToUUID(credentials.OrganizationID)
	if errorResponse != nil {
		s.logger.Error("Error occurred while parsing organizationID in service layer",
			zap.String("Email", credentials.Email), zap.Error(fmt.Errorf("%v", errorResponse)))
		return http.StatusBadRequest, errorResponse
	}

	result := models.User{
		Email:        credentials.Email,
		PasswordHash: passwordhash,
		Role:         string(credentials.Role),
		FullName:     credentials.FullName,
		AvatarURL:    credentials.AvatarURL,
		Timezone:     credentials.Timezone,
	}

	if organizationID != uuid.Nil {
		result.OrganizationID = &organizationID
	}

	return s.Repo.SignUp(result)

}
