package service

import (
	"context"

	"github.com/anthropics/pickle-go/apps/api/internal/model"
	"github.com/anthropics/pickle-go/apps/api/internal/repository"
	"github.com/anthropics/pickle-go/apps/api/pkg/jwt"
	"github.com/anthropics/pickle-go/apps/api/pkg/line"
	"github.com/google/uuid"
)

// AuthService handles authentication business logic
type AuthService struct {
	userRepo   *repository.UserRepository
	lineClient *line.Client
	jwtSecret  string
}

// NewAuthService creates a new AuthService
func NewAuthService(userRepo *repository.UserRepository, lineClient *line.Client, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		lineClient: lineClient,
		jwtSecret:  jwtSecret,
	}
}

// AuthResult represents the result of authentication
type AuthResult struct {
	User         *model.User
	AccessToken  string
	RefreshToken string
}

// AuthenticateWithLine authenticates a user using Line Login
func (s *AuthService) AuthenticateWithLine(ctx context.Context, code string) (*AuthResult, error) {
	// Exchange code for token
	tokenResp, err := s.lineClient.ExchangeToken(ctx, code)
	if err != nil {
		return nil, err
	}

	// Get user profile from Line
	profile, err := s.lineClient.GetProfile(ctx, tokenResp.AccessToken)
	if err != nil {
		return nil, err
	}

	// Create or update user
	user := &model.User{
		ID:          uuid.New(),
		LineUserID:  profile.UserID,
		DisplayName: profile.DisplayName,
		AvatarURL:   &profile.PictureURL,
	}

	err = s.userRepo.Upsert(ctx, user)
	if err != nil {
		return nil, err
	}

	// Generate JWT tokens
	accessToken, err := jwt.GenerateToken(user.ID.String(), user.DisplayName)
	if err != nil {
		return nil, err
	}

	refreshToken, err := jwt.GenerateRefreshToken(user.ID.String())
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// RefreshToken refreshes an access token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*AuthResult, error) {
	// Validate refresh token
	claims, err := jwt.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Get user
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Generate new access token
	accessToken, err := jwt.GenerateToken(user.ID.String(), user.DisplayName)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		User:        user,
		AccessToken: accessToken,
	}, nil
}
