package dto

type UpdateOrganizationRequest struct {
	Name    string `json:"name"`
	Domain  string `json:"domain"`
	LogoURL string `json:"logo_url"`
}

type CreateOrganizationRequest struct {
	Name    string `json:"name" validate:"required"`
	Domain  string `json:"domain" validate:"required"`
	LogoURL string `json:"logo_url" validate:"required"`
}
