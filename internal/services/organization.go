package services

import (
	"github.com/gofrs/uuid"
	"github.com/ms-kanban-server/internal/handlers/dto"
	"github.com/ms-kanban-server/internal/pkg/models"
	"github.com/ms-kanban-server/internal/pkg/response"
	"github.com/ms-kanban-server/internal/repository"
	"go.uber.org/zap"
)

type OrganizationService interface {
	GetOrganizationByID(id uuid.UUID) (models.Organization, *response.Error)
	CreateOrganization(row models.Organization) *response.Error
	UpdateOrganization(id uuid.UUID, req models.Organization) *response.Error
	DeleteOrganization(id uuid.UUID) *response.Error
}

func InitOrganizationService(repo repository.OrganizationRepository, AuthRepo repository.AuthRepository, logger *zap.Logger) OrganizationService {
	return &Organizationservice{
		OrganizationRepo: repo,
		logger:           logger,
		AuthRepo:         AuthRepo,
	}
}

type Organizationservice struct {
	AuthRepo         repository.AuthRepository
	OrganizationRepo repository.OrganizationRepository
	logger           *zap.Logger
}

func (s *Organizationservice) GetOrganizationByID(id uuid.UUID) (models.Organization, *response.Error) {

	return s.OrganizationRepo.GetByID(id)
}

func (s *Organizationservice) CreateOrganization(row models.Organization) *response.Error {

	err := s.OrganizationRepo.CreateOrganization(row)
	if err != nil {
		return err
	}

	organization, err := s.OrganizationRepo.GetByName(row.Name)
	if err != nil {
		return err
	}

	user := models.User{
		OrganizationID: &organization.ID,
		Role:           string(dto.RoleOrgAdmin),
	}

	err = s.AuthRepo.UpdateUser(row.CreatedBy, user)
	if err != nil {
		s.OrganizationRepo.DeleteOrganization(organization.ID)
		return err
	}

	return nil
}

func (s *Organizationservice) UpdateOrganization(OrganizationID uuid.UUID, req models.Organization) *response.Error {

	return s.OrganizationRepo.UpdateOrganization(OrganizationID, req)
}

func (s *Organizationservice) DeleteOrganization(id uuid.UUID) *response.Error {

	return s.OrganizationRepo.DeleteOrganization(id)
}
