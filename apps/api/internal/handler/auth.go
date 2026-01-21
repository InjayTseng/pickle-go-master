package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

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

// stateValidityDuration is the maximum age of a valid state parameter (5 minutes)
const stateValidityDuration = 5 * time.Minute

// getStateSecret returns the secret used for HMAC state validation
func getStateSecret() []byte {
	secret := os.Getenv("STATE_SECRET")
	if secret == "" {
		// Fall back to JWT secret if STATE_SECRET is not set
		secret = os.Getenv("JWT_SECRET")
	}
	if secret == "" && (os.Getenv("ENVIRONMENT") == "development" || os.Getenv("ENVIRONMENT") == "") {
		secret = "dev-state-secret-not-for-production"
	}
	return []byte(secret)
}

// computeStateHmac computes the HMAC signature for state validation
func computeStateHmac(timestamp, random string) string {
	mac := hmac.New(sha256.New, getStateSecret())
	mac.Write([]byte(timestamp + ":" + random))
	return base64.URLEncoding.EncodeToString(mac.Sum(nil))[:16]
}

// validateState validates the OAuth state parameter
// State format: timestamp:random:hmac
// Validates:
// 1. Correct format (3 parts separated by colons)
// 2. Timestamp is recent (within validity duration)
// 3. HMAC signature (optional, only if server-side secret is configured)
// Returns true if state is valid and not expired
func (h *AuthHandler) validateState(state string) bool {
	if state == "" {
		return false
	}

	parts := strings.Split(state, ":")
	if len(parts) != 3 {
		return false
	}

	timestamp, random, providedHmac := parts[0], parts[1], parts[2]

	// Validate random and hmac parts are not empty
	if random == "" || providedHmac == "" {
		return false
	}

	// Parse and validate timestamp (check if within validity duration)
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return false
	}

	// Check if state has expired
	stateTime := time.Unix(ts, 0)
	if time.Since(stateTime) > stateValidityDuration {
		return false
	}

	// Check if state is from the future (clock skew protection, allow 1 minute)
	if stateTime.After(time.Now().Add(1 * time.Minute)) {
		return false
	}

	// HMAC verification is optional - only verify if STATE_SECRET is explicitly set
	// This allows the frontend to use a public key for state generation while
	// still providing timestamp-based expiry protection
	stateSecret := os.Getenv("STATE_SECRET")
	if stateSecret != "" {
		expectedHmac := computeStateHmac(timestamp, random)
		return hmac.Equal([]byte(providedHmac), []byte(expectedHmac))
	}

	// If no STATE_SECRET, just validate format and timestamp
	// The primary CSRF protection is the sessionStorage verification on the client
	return true
}

// isStateValidationEnabled returns true if state validation should be enforced
// In development mode, validation can be skipped if explicitly disabled
func isStateValidationEnabled() bool {
	env := os.Getenv("ENVIRONMENT")
	// In production, always validate
	if env == "production" {
		return true
	}
	// In development, check if validation is explicitly disabled
	skipValidation := os.Getenv("SKIP_STATE_VALIDATION")
	return skipValidation != "true"
}

// LineCallback handles Line Login OAuth callback
// POST /api/v1/auth/line/callback
func (h *AuthHandler) LineCallback(c *gin.Context) {
	var req dto.LineCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("VALIDATION_ERROR", "Invalid request body"))
		return
	}

	// Validate CSRF state parameter
	if isStateValidationEnabled() {
		if !h.validateState(req.State) {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse("INVALID_STATE", "Invalid or expired state parameter"))
			return
		}
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
