package handler

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/anthropics/pickle-go/apps/api/internal/dto"
	"github.com/anthropics/pickle-go/apps/api/internal/middleware"
	"github.com/anthropics/pickle-go/apps/api/internal/model"
	"github.com/anthropics/pickle-go/apps/api/internal/repository"
	"github.com/anthropics/pickle-go/apps/api/pkg/jwt"
	"github.com/anthropics/pickle-go/apps/api/pkg/line"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuthHandler handles authentication related requests
type AuthHandler struct {
	userRepo   *repository.UserRepository
	lineClient *line.Client
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(userRepo *repository.UserRepository, lineClient *line.Client) *AuthHandler {
	return &AuthHandler{
		userRepo:   userRepo,
		lineClient: lineClient,
	}
}

// LineCallback handles Line Login OAuth callback
// POST /api/v1/auth/line/callback
func (h *AuthHandler) LineCallback(c *gin.Context) {
	var req dto.LineCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("VALIDATION_ERROR", "Invalid request body"))
		return
	}

	// Exchange authorization code for access token
	tokenResp, err := h.lineClient.ExchangeToken(c.Request.Context(), req.Code)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("LINE_AUTH_FAILED", "Failed to authenticate with Line"))
		return
	}

	// Get user profile from Line
	profile, err := h.lineClient.GetProfile(c.Request.Context(), tokenResp.AccessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("LINE_PROFILE_FAILED", "Failed to get Line profile"))
		return
	}

	// Create or update user
	user := &model.User{
		ID:          uuid.New(),
		LineUserID:  profile.UserID,
		DisplayName: profile.DisplayName,
	}

	// Only set avatar URL if not empty
	if profile.PictureURL != "" {
		user.AvatarURL = &profile.PictureURL
	}

	// Upsert user (create if not exists, update if exists)
	if err := h.userRepo.Upsert(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("USER_CREATION_FAILED", "Failed to create or update user"))
		return
	}

	// Generate JWT tokens
	accessToken, err := jwt.GenerateToken(user.ID.String(), user.DisplayName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("TOKEN_GENERATION_FAILED", "Failed to generate access token"))
		return
	}

	refreshToken, err := jwt.GenerateRefreshToken(user.ID.String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("TOKEN_GENERATION_FAILED", "Failed to generate refresh token"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(dto.AuthResponse{
		User:         dto.FromUser(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}))
}

// RefreshToken handles JWT token refresh
// POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("VALIDATION_ERROR", "Invalid request body"))
		return
	}

	// Validate refresh token
	claims, err := jwt.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		if errors.Is(err, jwt.ErrExpiredToken) {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse("TOKEN_EXPIRED", "Refresh token has expired"))
			return
		}
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("INVALID_TOKEN", "Invalid refresh token"))
		return
	}

	// Get user
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("INVALID_TOKEN", "Invalid user ID in token"))
		return
	}

	user, err := h.userRepo.FindByID(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse("USER_NOT_FOUND", "User not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("INTERNAL_ERROR", "Failed to get user"))
		return
	}

	// Generate new access token
	accessToken, err := jwt.GenerateToken(user.ID.String(), user.DisplayName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("TOKEN_GENERATION_FAILED", "Failed to generate access token"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(dto.AuthResponse{
		User:        dto.FromUser(user),
		AccessToken: accessToken,
	}))
}

// Logout handles user logout
// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// For JWT-based auth, logout is typically handled client-side by removing the token
	// Server-side, we can optionally blacklist the token if needed
	// For now, we just return success

	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"message": "Logged out successfully",
	}))
}

// GetCurrentUser returns the current authenticated user
// GET /api/v1/users/me
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	claims, ok := middleware.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("UNAUTHORIZED", "Not authenticated"))
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("INVALID_TOKEN", "Invalid user ID in token"))
		return
	}

	user, err := h.userRepo.FindByID(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse("USER_NOT_FOUND", "User not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("INTERNAL_ERROR", "Failed to get user"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(dto.FromUser(user)))
}

// Legacy handlers for backward compatibility (these will be replaced by injected handlers)

// LineCallback is the legacy handler (kept for backward compatibility)
func LineCallback(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, dto.ErrorResponse("NOT_IMPLEMENTED", "Handler not configured"))
}

// RefreshToken is the legacy handler (kept for backward compatibility)
func RefreshToken(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, dto.ErrorResponse("NOT_IMPLEMENTED", "Handler not configured"))
}

// Logout is the legacy handler (kept for backward compatibility)
func Logout(c *gin.Context) {
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"message": "Logged out successfully",
	}))
}
