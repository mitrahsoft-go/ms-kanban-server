package services

import (
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/ms-kanban-server/internal/handlers/dto"
	"github.com/ms-kanban-server/internal/pkg/models"
	"github.com/ms-kanban-server/internal/pkg/response"
	"github.com/ms-kanban-server/internal/pkg/utils"
	"go.uber.org/zap"
)

type stubAuthRepository struct {
	user         models.User
	refreshToken models.RefreshToken
	err          *response.Error
}

func (s *stubAuthRepository) SignIn(email string) (models.User, *response.Error) {
	if s.err != nil {
		return models.User{}, s.err
	}
	return s.user, nil
}

func (s *stubAuthRepository) SignInByID(id uuid.UUID) (models.User, *response.Error) {
	if s.err != nil {
		return models.User{}, s.err
	}
	return s.user, nil
}

func (s *stubAuthRepository) SignUp(row models.User) *response.Error {
	return nil
}

func (s *stubAuthRepository) StoreRefreshToken(token models.RefreshToken) *response.Error {
	return nil
}

func (s *stubAuthRepository) GetRefreshToken(tokenHash string) (models.RefreshToken, *response.Error) {
	if s.err != nil {
		return models.RefreshToken{}, s.err
	}
	return s.refreshToken, nil
}

func (s *stubAuthRepository) ChangePassword(tokenHash string, userID uuid.UUID) *response.Error {
	if s.err != nil {
		return s.err
	}
	return nil
}
func (s *stubAuthRepository) UpdateUser(userID uuid.UUID, req models.User) *response.Error {
	if s.err != nil {
		return s.err
	}
	return nil
}

func TestSignInReturnsUnauthorizedForInvalidPassword(t *testing.T) {
	hash, err := utils.HashPassword("correct-password")
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	repo := &stubAuthRepository{
		user: models.User{ID: uuid.Must(uuid.NewV4()),
			Email:        "user@example.com",
			PasswordHash: hash,
			Role:         "developer",
			IsActive:     true,
		},
	}
	service := InitAuthService(repo, zap.NewNop())

	result, authErr := service.SignIn(dto.SignInRequest{Email: "user@example.com", Password: "wrong-password"})
	if authErr == nil {
		t.Fatalf("expected auth error, got nil")
	}
	if authErr.StatusCode != 401 {
		t.Fatalf("expected 401 status, got %d", authErr.StatusCode)
	}
	if result != nil {
		t.Fatalf("expected no auth tokens, got %#v", result)
	}
}

func TestSignInRejectsInactiveUser(t *testing.T) {
	hash, err := utils.HashPassword("correct-password")
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	repo := &stubAuthRepository{user: models.User{ID: uuid.Must(uuid.NewV4()), Email: "user@example.com", PasswordHash: hash, Role: "developer", IsActive: false}}
	service := InitAuthService(repo, zap.NewNop())

	result, authErr := service.SignIn(dto.SignInRequest{Email: "user@example.com", Password: "correct-password"})
	if authErr == nil {
		t.Fatalf("expected auth error, got nil")
	}
	if authErr.StatusCode != 403 {
		t.Fatalf("expected 403 status, got %d", authErr.StatusCode)
	}
	if result != nil {
		t.Fatalf("expected no auth tokens, got %#v", result)
	}
}

func TestSignInReturnsAuthTokensForValidCredentials(t *testing.T) {
	hash, err := utils.HashPassword("correct-password")
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	repo := &stubAuthRepository{user: models.User{ID: uuid.Must(uuid.NewV7()), Email: "user@example.com", PasswordHash: hash, Role: "developer", IsActive: true}}
	service := InitAuthService(repo, zap.NewNop())

	result, authErr := service.SignIn(dto.SignInRequest{Email: "user@example.com", Password: "correct-password"})
	if authErr != nil {
		t.Fatalf("expected successful login, got error: %v", authErr)
	}
	if result == nil {
		t.Fatalf("expected auth tokens, got nil")
	}
	if result.AccessToken == "" {
		t.Fatal("expected access token to be populated")
	}
	if result.RefreshToken == "" {
		t.Fatal("expected refresh token to be populated")
	}
	if result.TokenType != "Bearer" {
		t.Fatalf("expected Bearer token type, got %q", result.TokenType)
	}
}

func TestRefreshTokenReturnsNewAccessTokenForValidRefreshToken(t *testing.T) {
	refreshHash, err := utils.HashPassword("valid-refresh")
	if err != nil {
		t.Fatalf("failed to hash refresh token: %v", err)
	}

	repo := &stubAuthRepository{
		user:         models.User{ID: uuid.Must(uuid.NewV7()), Email: "user@example.com", Role: "developer", IsActive: true},
		refreshToken: models.RefreshToken{UserID: uuid.Must(uuid.NewV7()), TokenHash: refreshHash, ExpiresAt: time.Now().Add(7 * 24 * time.Hour)},
	}
	service := InitAuthService(repo, zap.NewNop())

	result, authErr := service.RefreshToken(dto.RefreshTokenRequest{RefreshToken: "valid-refresh"})
	if authErr != nil {
		t.Fatalf("expected refresh to succeed, got error: %v", authErr)
	}
	if result == nil {
		t.Fatal("expected access token payload, got nil")
	}
	if result.AccessToken == "" {
		t.Fatal("expected access token to be populated")
	}
	if result.TokenType != "Bearer" {
		t.Fatalf("expected Bearer token type, got %q", result.TokenType)
	}
}
