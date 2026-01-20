package line

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	tokenEndpoint   = "https://api.line.me/oauth2/v2.1/token"
	profileEndpoint = "https://api.line.me/v2/profile"
)

// Client is a Line Login API client
type Client struct {
	httpClient    *http.Client
	channelID     string
	channelSecret string
	redirectURI   string
}

// Config holds Line Login configuration
type Config struct {
	ChannelID     string
	ChannelSecret string
	RedirectURI   string
}

// NewClient creates a new Line Login client
func NewClient(cfg Config) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		channelID:     cfg.ChannelID,
		channelSecret: cfg.ChannelSecret,
		redirectURI:   cfg.RedirectURI,
	}
}

// TokenResponse represents the response from Line's token endpoint
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
	IDToken      string `json:"id_token"`
}

// Profile represents a Line user profile
type Profile struct {
	UserID        string `json:"userId"`
	DisplayName   string `json:"displayName"`
	PictureURL    string `json:"pictureUrl"`
	StatusMessage string `json:"statusMessage"`
}

// ExchangeToken exchanges an authorization code for an access token
func (c *Client) ExchangeToken(ctx context.Context, code string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", c.redirectURI)
	data.Set("client_id", c.channelID)
	data.Set("client_secret", c.channelSecret)

	req, err := http.NewRequestWithContext(ctx, "POST", tokenEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed: %s", string(body))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	return &tokenResp, nil
}

// GetProfile retrieves the user's Line profile
func (c *Client) GetProfile(ctx context.Context, accessToken string) (*Profile, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", profileEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get profile: %s", string(body))
	}

	var profile Profile
	if err := json.Unmarshal(body, &profile); err != nil {
		return nil, fmt.Errorf("failed to parse profile: %w", err)
	}

	return &profile, nil
}

// GetAuthorizationURL returns the Line Login authorization URL
func (c *Client) GetAuthorizationURL(state string) string {
	params := url.Values{}
	params.Set("response_type", "code")
	params.Set("client_id", c.channelID)
	params.Set("redirect_uri", c.redirectURI)
	params.Set("state", state)
	params.Set("scope", "profile openid")

	return "https://access.line.me/oauth2/v2.1/authorize?" + params.Encode()
}
