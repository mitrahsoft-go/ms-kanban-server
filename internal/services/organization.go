package services

import (
	"github.com/gofrs/uuid"
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

func InitOrganizationService(repo repository.OrganizationRepository, logger *zap.Logger) OrganizationService {
	return &Organizationservice{
		Repo:   repo,
		logger: logger,
	}
}

type Organizationservice struct {
	Repo   repository.OrganizationRepository
	logger *zap.Logger
}

func (s *Organizationservice) GetOrganizationByID(id uuid.UUID) (models.Organization, *response.Error) {

	return s.Repo.GetByID(id)
}

func (s *Organizationservice) CreateOrganization(row models.Organization) *response.Error {

	return s.Repo.CreateOrganization(row)
}

func (s *Organizationservice) UpdateOrganization(OrganizationID uuid.UUID, req models.Organization) *response.Error {

	return s.Repo.UpdateOrganization(OrganizationID, req)
}

func (s *Organizationservice) DeleteOrganization(id uuid.UUID) *response.Error {

	return s.Repo.DeleteOrganization(id)
}
