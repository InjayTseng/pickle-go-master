package jwt

import (
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestMain(m *testing.M) {
	// Set up test environment
	os.Setenv("ENVIRONMENT", "development")
	os.Setenv("JWT_SECRET", "test-secret-key-for-testing")
	os.Setenv("JWT_EXPIRY", "1h")

	// Re-initialize the secret for tests
	jwtSecret = []byte(os.Getenv("JWT_SECRET"))

	os.Exit(m.Run())
}

func TestGenerateToken(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		displayName string
		wantErr     bool
	}{
		{
			name:        "valid token generation",
			userID:      "550e8400-e29b-41d4-a716-446655440000",
			displayName: "Test User",
			wantErr:     false,
		},
		{
			name:        "empty user ID",
			userID:      "",
			displayName: "Test User",
			wantErr:     false, // Empty ID is allowed, validation is business logic
		},
		{
			name:        "empty display name",
			userID:      "550e8400-e29b-41d4-a716-446655440000",
			displayName: "",
			wantErr:     false,
		},
		{
			name:        "unicode display name",
			userID:      "550e8400-e29b-41d4-a716-446655440000",
			displayName: "Test User",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateToken(tt.userID, tt.displayName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && token == "" {
				t.Error("GenerateToken() returned empty token")
			}
		})
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	tests := []struct {
		name    string
		userID  string
		wantErr bool
	}{
		{
			name:    "valid refresh token generation",
			userID:  "550e8400-e29b-41d4-a716-446655440000",
			wantErr: false,
		},
		{
			name:    "empty user ID",
			userID:  "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateRefreshToken(tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateRefreshToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && token == "" {
				t.Error("GenerateRefreshToken() returned empty token")
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	// Generate a valid token for testing
	userID := "550e8400-e29b-41d4-a716-446655440000"
	displayName := "Test User"
	validToken, err := GenerateToken(userID, displayName)
	if err != nil {
		t.Fatalf("Failed to generate token for test: %v", err)
	}

	tests := []struct {
		name        string
		token       string
		wantUserID  string
		wantDisplay string
		wantErr     error
	}{
		{
			name:        "valid token",
			token:       validToken,
			wantUserID:  userID,
			wantDisplay: displayName,
			wantErr:     nil,
		},
		{
			name:    "empty token",
			token:   "",
			wantErr: ErrInvalidToken,
		},
		{
			name:    "malformed token",
			token:   "not.a.valid.token",
			wantErr: ErrInvalidToken,
		},
		{
			name:    "token with invalid signature",
			token:   validToken + "tampered",
			wantErr: ErrInvalidToken,
		},
		{
			name:    "completely random string",
			token:   "randomstring123",
			wantErr: ErrInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := ValidateToken(tt.token)
			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("ValidateToken() expected error %v, got nil", tt.wantErr)
					return
				}
				if err != tt.wantErr {
					t.Errorf("ValidateToken() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("ValidateToken() unexpected error = %v", err)
				return
			}
			if claims.UserID != tt.wantUserID {
				t.Errorf("ValidateToken() UserID = %v, want %v", claims.UserID, tt.wantUserID)
			}
			if claims.DisplayName != tt.wantDisplay {
				t.Errorf("ValidateToken() DisplayName = %v, want %v", claims.DisplayName, tt.wantDisplay)
			}
		})
	}
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	// Create an expired token manually
	claims := Claims{
		UserID:      "550e8400-e29b-41d4-a716-446655440000",
		DisplayName: "Test User",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // Expired 1 hour ago
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Issuer:    "pickle-go",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	expiredToken, err := token.SignedString(getSecret())
	if err != nil {
		t.Fatalf("Failed to create expired token: %v", err)
	}

	_, err = ValidateToken(expiredToken)
	if err != ErrExpiredToken {
		t.Errorf("ValidateToken() error = %v, want %v", err, ErrExpiredToken)
	}
}

func TestValidateToken_WrongSigningMethod(t *testing.T) {
	// Create a token with a different signing method (none)
	claims := Claims{
		UserID:      "550e8400-e29b-41d4-a716-446655440000",
		DisplayName: "Test User",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "pickle-go",
		},
	}

	// Use RS256 (different from expected HS256)
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	invalidMethodToken, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("Failed to create token with wrong method: %v", err)
	}

	_, err = ValidateToken(invalidMethodToken)
	if err != ErrInvalidToken {
		t.Errorf("ValidateToken() error = %v, want %v", err, ErrInvalidToken)
	}
}

func TestValidateRefreshToken(t *testing.T) {
	// Generate a valid refresh token for testing
	userID := "550e8400-e29b-41d4-a716-446655440000"
	validToken, err := GenerateRefreshToken(userID)
	if err != nil {
		t.Fatalf("Failed to generate refresh token for test: %v", err)
	}

	tests := []struct {
		name       string
		token      string
		wantUserID string
		wantErr    error
	}{
		{
			name:       "valid refresh token",
			token:      validToken,
			wantUserID: userID,
			wantErr:    nil,
		},
		{
			name:    "empty token",
			token:   "",
			wantErr: ErrInvalidToken,
		},
		{
			name:    "malformed token",
			token:   "not.a.valid.token",
			wantErr: ErrInvalidToken,
		},
		{
			name:    "token with invalid signature",
			token:   validToken + "tampered",
			wantErr: ErrInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := ValidateRefreshToken(tt.token)
			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("ValidateRefreshToken() expected error %v, got nil", tt.wantErr)
					return
				}
				if err != tt.wantErr {
					t.Errorf("ValidateRefreshToken() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("ValidateRefreshToken() unexpected error = %v", err)
				return
			}
			if claims.UserID != tt.wantUserID {
				t.Errorf("ValidateRefreshToken() UserID = %v, want %v", claims.UserID, tt.wantUserID)
			}
		})
	}
}

func TestValidateRefreshToken_ExpiredToken(t *testing.T) {
	// Create an expired refresh token manually
	claims := RefreshClaims{
		UserID: "550e8400-e29b-41d4-a716-446655440000",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // Expired 1 hour ago
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Issuer:    "pickle-go",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	expiredToken, err := token.SignedString(getSecret())
	if err != nil {
		t.Fatalf("Failed to create expired token: %v", err)
	}

	_, err = ValidateRefreshToken(expiredToken)
	if err != ErrExpiredToken {
		t.Errorf("ValidateRefreshToken() error = %v, want %v", err, ErrExpiredToken)
	}
}

func TestTokenRoundTrip(t *testing.T) {
	// Test that generating and validating a token preserves the claims
	userID := "550e8400-e29b-41d4-a716-446655440000"
	displayName := "Test User"

	// Generate access token
	accessToken, err := GenerateToken(userID, displayName)
	if err != nil {
		t.Fatalf("Failed to generate access token: %v", err)
	}

	// Validate access token
	accessClaims, err := ValidateToken(accessToken)
	if err != nil {
		t.Fatalf("Failed to validate access token: %v", err)
	}

	if accessClaims.UserID != userID {
		t.Errorf("Access token UserID = %v, want %v", accessClaims.UserID, userID)
	}
	if accessClaims.DisplayName != displayName {
		t.Errorf("Access token DisplayName = %v, want %v", accessClaims.DisplayName, displayName)
	}
	if accessClaims.Issuer != "pickle-go" {
		t.Errorf("Access token Issuer = %v, want pickle-go", accessClaims.Issuer)
	}

	// Generate refresh token
	refreshToken, err := GenerateRefreshToken(userID)
	if err != nil {
		t.Fatalf("Failed to generate refresh token: %v", err)
	}

	// Validate refresh token
	refreshClaims, err := ValidateRefreshToken(refreshToken)
	if err != nil {
		t.Fatalf("Failed to validate refresh token: %v", err)
	}

	if refreshClaims.UserID != userID {
		t.Errorf("Refresh token UserID = %v, want %v", refreshClaims.UserID, userID)
	}
	if refreshClaims.Issuer != "pickle-go" {
		t.Errorf("Refresh token Issuer = %v, want pickle-go", refreshClaims.Issuer)
	}
}

func TestAccessAndRefreshTokensAreDifferent(t *testing.T) {
	userID := "550e8400-e29b-41d4-a716-446655440000"

	accessToken, err := GenerateToken(userID, "Test User")
	if err != nil {
		t.Fatalf("Failed to generate access token: %v", err)
	}

	refreshToken, err := GenerateRefreshToken(userID)
	if err != nil {
		t.Fatalf("Failed to generate refresh token: %v", err)
	}

	if accessToken == refreshToken {
		t.Error("Access token and refresh token should be different")
	}
}

func TestGetExpiry(t *testing.T) {
	// Test with custom expiry
	os.Setenv("JWT_EXPIRY", "2h")
	expiry := getExpiry()
	if expiry != 2*time.Hour {
		t.Errorf("getExpiry() = %v, want %v", expiry, 2*time.Hour)
	}

	// Test with invalid expiry (should fall back to default)
	os.Setenv("JWT_EXPIRY", "invalid")
	expiry = getExpiry()
	if expiry != 168*time.Hour {
		t.Errorf("getExpiry() with invalid value = %v, want %v", expiry, 168*time.Hour)
	}

	// Reset to test value
	os.Setenv("JWT_EXPIRY", "1h")
}

func TestTokenClaimsContainExpectedFields(t *testing.T) {
	userID := "550e8400-e29b-41d4-a716-446655440000"
	displayName := "Test User"

	token, err := GenerateToken(userID, displayName)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	claims, err := ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	// Check that expiration is in the future
	if claims.ExpiresAt == nil {
		t.Error("Token should have an expiration time")
	} else if claims.ExpiresAt.Time.Before(time.Now()) {
		t.Error("Token expiration should be in the future")
	}

	// Check that issued at is in the past or now
	if claims.IssuedAt == nil {
		t.Error("Token should have an issued at time")
	} else if claims.IssuedAt.Time.After(time.Now().Add(1 * time.Second)) {
		t.Error("Token issued at should be in the past or now")
	}

	// Check issuer
	if claims.Issuer != "pickle-go" {
		t.Errorf("Token Issuer = %v, want pickle-go", claims.Issuer)
	}
}
