package services

import (
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
	SignIn(credentials dto.SignInRequest) (uuid.UUID, string, *response.Error)
	SignUp(credentials dto.SignUpRequest) *response.Error
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

func (s *authservice) SignIn(credentials dto.SignInRequest) (uuid.UUID, string, *response.Error) {

	result, err := s.Repo.SignIn(credentials.Email)
	if err != nil {
		return uuid.Nil, "", err
	}

	if utils.IsValidPassword(result.PasswordHash, credentials.Password) {

		s.logger.Error(" Validated failure in Password before login in service layer",
			zap.String("Email", credentials.Email))
		return uuid.Nil, "", &response.Error{
			Code:       response.ErrUnauthorized,
			StatusCode: http.StatusUnauthorized,
			Message:    "Enter valid Email or Password before login",
			Details: []response.Details{
				{
					Field:   "Email/Password",
					Message: "User not found :" + credentials.Email,
				},
			},
		}
	}
	return result.ID, string(result.Role), nil
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
