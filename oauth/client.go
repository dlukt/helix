package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultBaseURL = "https://id.twitch.tv/oauth2"

// ErrValidateUnauthorized reports a /validate response that rejected the token.
var ErrValidateUnauthorized = errors.New("oauth: validate unauthorized")

// Config configures an OAuth client.
type Config struct {
	ClientID     string
	ClientSecret string
	HTTPClient   *http.Client
	BaseURL      string
}

// Client implements Twitch OAuth endpoints.
type Client struct {
	clientID     string
	clientSecret string
	httpClient   *http.Client
	baseURL      string
}

// AuthorizationURLParams configures the authorization URL.
type AuthorizationURLParams struct {
	RedirectURI string
	Scopes      []string
	State       string
	ForceVerify bool
}

// AuthorizationCodeExchange configures an auth code exchange.
type AuthorizationCodeExchange struct {
	Code        string
	RedirectURI string
}

// DeviceAuthorization contains the data returned when starting the device code flow.
type DeviceAuthorization struct {
	DeviceCode      string `json:"device_code"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
}

// DeviceCodeExchange configures the device code token exchange.
type DeviceCodeExchange struct {
	DeviceCode string
	Scopes     []string
}

// NewClient constructs an OAuth client.
func NewClient(cfg Config) *Client {
	baseURL := strings.TrimRight(cfg.BaseURL, "/")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{
		clientID:     cfg.ClientID,
		clientSecret: cfg.ClientSecret,
		httpClient:   httpClient,
		baseURL:      baseURL,
	}
}

// AuthorizationURL builds a Twitch authorization URL for the authorization code flow.
func (c *Client) AuthorizationURL(params AuthorizationURLParams) string {
	query := url.Values{}
	query.Set("client_id", c.clientID)
	query.Set("response_type", "code")
	query.Set("redirect_uri", params.RedirectURI)
	if len(params.Scopes) > 0 {
		query.Set("scope", strings.Join(params.Scopes, " "))
	}
	if params.State != "" {
		query.Set("state", params.State)
	}
	if params.ForceVerify {
		query.Set("force_verify", "true")
	}
	return c.baseURL + "/authorize?" + query.Encode()
}

// ExchangeAuthorizationCode exchanges an auth code for a token.
func (c *Client) ExchangeAuthorizationCode(ctx context.Context, params AuthorizationCodeExchange) (Token, error) {
	form := url.Values{}
	form.Set("client_id", c.clientID)
	form.Set("client_secret", c.clientSecret)
	form.Set("code", params.Code)
	form.Set("grant_type", "authorization_code")
	form.Set("redirect_uri", params.RedirectURI)
	return c.exchangeForm(ctx, form)
}

// RefreshToken exchanges a refresh token for a new token.
func (c *Client) RefreshToken(ctx context.Context, refreshToken string) (Token, error) {
	form := url.Values{}
	form.Set("client_id", c.clientID)
	if c.clientSecret != "" {
		form.Set("client_secret", c.clientSecret)
	}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)
	return c.exchangeForm(ctx, form)
}

// ExchangeClientCredentials exchanges client credentials for an app token.
func (c *Client) ExchangeClientCredentials(ctx context.Context, scopes []string) (Token, error) {
	form := url.Values{}
	form.Set("client_id", c.clientID)
	form.Set("client_secret", c.clientSecret)
	form.Set("grant_type", "client_credentials")
	if len(scopes) > 0 {
		form.Set("scope", strings.Join(scopes, " "))
	}
	return c.exchangeForm(ctx, form)
}

// StartDeviceAuthorization starts the device code flow.
func (c *Client) StartDeviceAuthorization(ctx context.Context, scopes []string) (DeviceAuthorization, error) {
	form := url.Values{}
	form.Set("client_id", c.clientID)
	form.Set("scopes", strings.Join(scopes, " "))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/device", strings.NewReader(form.Encode()))
	if err != nil {
		return DeviceAuthorization{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return DeviceAuthorization{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return DeviceAuthorization{}, fmt.Errorf("oauth: device authorization failed: status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var auth DeviceAuthorization
	if err := json.NewDecoder(resp.Body).Decode(&auth); err != nil {
		return DeviceAuthorization{}, err
	}
	return auth, nil
}

// ExchangeDeviceCode exchanges a completed device code authorization for a token.
func (c *Client) ExchangeDeviceCode(ctx context.Context, params DeviceCodeExchange) (Token, error) {
	form := url.Values{}
	form.Set("client_id", c.clientID)
	if c.clientSecret != "" {
		form.Set("client_secret", c.clientSecret)
	}
	form.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")
	form.Set("device_code", params.DeviceCode)
	if len(params.Scopes) > 0 {
		form.Set("scopes", strings.Join(params.Scopes, " "))
	}
	return c.exchangeForm(ctx, form)
}

func (c *Client) exchangeForm(ctx context.Context, form url.Values) (Token, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/token", strings.NewReader(form.Encode()))
	if err != nil {
		return Token{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Token{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return Token{}, fmt.Errorf("oauth: token exchange failed: status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var payload struct {
		AccessToken  string   `json:"access_token"`
		RefreshToken string   `json:"refresh_token"`
		ExpiresIn    int      `json:"expires_in"`
		Scope        []string `json:"scope"`
		TokenType    string   `json:"token_type"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return Token{}, err
	}

	token := Token{
		AccessToken:  payload.AccessToken,
		RefreshToken: payload.RefreshToken,
		TokenType:    payload.TokenType,
		Scopes:       payload.Scope,
	}
	if payload.ExpiresIn > 0 {
		token.Expiry = time.Now().Add(time.Duration(payload.ExpiresIn) * time.Second)
	}
	return token, nil
}

// ValidateToken validates a token via Twitch's validate endpoint.
func (c *Client) ValidateToken(ctx context.Context, accessToken string) (ValidatedToken, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/validate", nil)
	if err != nil {
		return ValidatedToken{}, err
	}
	req.Header.Set("Authorization", "OAuth "+accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return ValidatedToken{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode == http.StatusUnauthorized {
			return ValidatedToken{}, fmt.Errorf("%w: status %d: %s", ErrValidateUnauthorized, resp.StatusCode, strings.TrimSpace(string(body)))
		}
		return ValidatedToken{}, fmt.Errorf("oauth: validate failed: status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var validated ValidatedToken
	if err := json.NewDecoder(resp.Body).Decode(&validated); err != nil {
		return ValidatedToken{}, err
	}
	return validated, nil
}

// RevokeToken revokes a token via Twitch's revoke endpoint.
func (c *Client) RevokeToken(ctx context.Context, token string) error {
	query := url.Values{}
	query.Set("client_id", c.clientID)
	query.Set("token", token)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/revoke?"+query.Encode(), nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("oauth: revoke failed: status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}
