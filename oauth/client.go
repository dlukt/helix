package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const defaultBaseURL = "https://id.twitch.tv/oauth2"

// ErrValidateUnauthorized reports a /validate response that rejected the token.
var ErrValidateUnauthorized = errors.New("oauth: validate unauthorized")

// ErrAuthorizationPending reports a device-code poll before the user completes authorization.
var ErrAuthorizationPending = errors.New("oauth: authorization pending")

// ErrSlowDown reports a device-code poll that should back off before retrying.
var ErrSlowDown = errors.New("oauth: slow down")

// ErrDeviceCodeExpired reports a device-code flow that expired before completion.
var ErrDeviceCodeExpired = errors.New("oauth: device code expired")

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

// ImplicitAuthorizationURLParams configures an implicit-flow authorization URL.
type ImplicitAuthorizationURLParams struct {
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
	DeviceCode      string    `json:"device_code"`
	ExpiresIn       int       `json:"expires_in"`
	Interval        int       `json:"interval"`
	UserCode        string    `json:"user_code"`
	VerificationURI string    `json:"verification_uri"`
	IssuedAt        time.Time `json:"issued_at,omitempty"`
}

// DeviceCodeExchange configures the device code token exchange.
type DeviceCodeExchange struct {
	DeviceCode string
	Scopes     []string
}

// ImplicitCallback contains the values Twitch returns in the callback fragment for implicit auth.
type ImplicitCallback struct {
	AccessToken      string
	Scopes           []string
	State            string
	TokenType        string
	ExpiresIn        int
	ErrorCode        string
	ErrorDescription string
}

// ErrorResponse is returned for non-2xx OAuth responses.
type ErrorResponse struct {
	Endpoint   string
	StatusCode int
	Status     int
	ErrorCode  string
	Message    string
	Body       []byte
}

func (e *ErrorResponse) Error() string {
	if e.ErrorCode != "" && e.Message != "" {
		return fmt.Sprintf("oauth: %s failed: status %d %s: %s", e.Endpoint, e.StatusCode, e.ErrorCode, e.Message)
	}
	if e.Message != "" {
		return fmt.Sprintf("oauth: %s failed: status %d: %s", e.Endpoint, e.StatusCode, e.Message)
	}
	return fmt.Sprintf("oauth: %s failed: status %d", e.Endpoint, e.StatusCode)
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
	return c.authorizationURL("code", params.RedirectURI, params.Scopes, params.State, params.ForceVerify)
}

// ImplicitAuthorizationURL builds a Twitch authorization URL for the implicit flow.
func (c *Client) ImplicitAuthorizationURL(params ImplicitAuthorizationURLParams) string {
	return c.authorizationURL("token", params.RedirectURI, params.Scopes, params.State, params.ForceVerify)
}

func (c *Client) authorizationURL(responseType, redirectURI string, scopes []string, state string, forceVerify bool) string {
	query := url.Values{}
	query.Set("client_id", c.clientID)
	query.Set("response_type", responseType)
	query.Set("redirect_uri", redirectURI)
	if len(scopes) > 0 {
		query.Set("scope", strings.Join(scopes, " "))
	}
	if state != "" {
		query.Set("state", state)
	}
	if forceVerify {
		query.Set("force_verify", "true")
	}
	return c.baseURL + "/authorize?" + query.Encode()
}

// ParseImplicitCallbackFragment parses an implicit-flow callback fragment and also accepts
// the documented query-string denial parameters Twitch sends when the user rejects access.
func ParseImplicitCallbackFragment(fragment string) (ImplicitCallback, error) {
	values, err := parseImplicitCallbackValues(fragment)
	if err != nil {
		return ImplicitCallback{}, err
	}

	var callback ImplicitCallback
	callback.AccessToken = values.Get("access_token")
	callback.State = values.Get("state")
	callback.TokenType = values.Get("token_type")
	callback.ErrorCode = values.Get("error")
	callback.ErrorDescription = values.Get("error_description")
	if scope := values.Get("scope"); scope != "" {
		callback.Scopes = strings.Fields(scope)
	}
	if expiresIn := values.Get("expires_in"); expiresIn != "" {
		parsed, err := strconv.Atoi(expiresIn)
		if err != nil {
			return ImplicitCallback{}, fmt.Errorf("oauth: parse expires_in: %w", err)
		}
		callback.ExpiresIn = parsed
	}
	return callback, nil
}

func parseImplicitCallbackValues(raw string) (url.Values, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return url.Values{}, nil
	}

	if parsed, err := url.Parse(raw); err == nil && (parsed.Scheme != "" || parsed.Host != "" || strings.HasPrefix(raw, "/")) {
		values := parsed.Query()
		if parsed.Fragment == "" {
			return values, nil
		}
		fragmentValues, err := url.ParseQuery(parsed.Fragment)
		if err != nil {
			return nil, err
		}
		for key, vals := range fragmentValues {
			values[key] = vals
		}
		return values, nil
	}

	raw = strings.TrimPrefix(raw, "#")
	raw = strings.TrimPrefix(raw, "?")
	return url.ParseQuery(raw)
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
	return decodeDeviceAuthorizationResponse(resp)
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

// PollDeviceAuthorization polls the device-code exchange endpoint until the user authorizes or the flow ends.
func (c *Client) PollDeviceAuthorization(ctx context.Context, auth DeviceAuthorization, scopes []string) (Token, error) {
	interval := time.Duration(auth.Interval) * time.Second
	if interval <= 0 {
		interval = 5 * time.Second
	}

	expiresAt := deviceAuthorizationExpiresAt(auth)

	for {
		if !expiresAt.IsZero() && !time.Now().Before(expiresAt) {
			return Token{}, ErrDeviceCodeExpired
		}

		wait := interval
		if !expiresAt.IsZero() {
			remaining := time.Until(expiresAt)
			if remaining <= 0 {
				return Token{}, ErrDeviceCodeExpired
			}
			if wait > remaining {
				wait = remaining
			}
		}

		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return Token{}, ctx.Err()
		case <-timer.C:
			if !expiresAt.IsZero() && !time.Now().Before(expiresAt) {
				return Token{}, ErrDeviceCodeExpired
			}
		}

		token, err := c.ExchangeDeviceCode(ctx, DeviceCodeExchange{
			DeviceCode: auth.DeviceCode,
			Scopes:     scopes,
		})
		if err == nil {
			return token, nil
		}

		if errors.Is(err, ErrAuthorizationPending) {
			continue
		}
		if errors.Is(err, ErrSlowDown) {
			interval += 5 * time.Second
			continue
		}
		var oauthErr *ErrorResponse
		if errors.As(err, &oauthErr) && isExpiredDeviceCodeError(oauthErr) {
			return Token{}, ErrDeviceCodeExpired
		}
		return Token{}, err
	}
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
	payload, err := decodeTokenResponse(resp)
	if err != nil {
		return Token{}, err
	}
	return buildToken(payload), nil
}

type tokenPayload struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	ExpiresIn    int      `json:"expires_in"`
	Scope        []string `json:"scope"`
	TokenType    string   `json:"token_type"`
}

func decodeTokenResponse(resp *http.Response) (tokenPayload, error) {
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		err := decodeErrorResponse("token exchange", resp.StatusCode, body)
		switch {
		case isPendingDeviceAuthorization(err):
			return tokenPayload{}, fmt.Errorf("%w: %v", ErrAuthorizationPending, err)
		case isSlowDownDeviceAuthorization(err):
			return tokenPayload{}, fmt.Errorf("%w: %v", ErrSlowDown, err)
		default:
			return tokenPayload{}, err
		}
	}

	var payload tokenPayload
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return tokenPayload{}, err
	}
	return payload, nil
}

func buildToken(payload tokenPayload) Token {
	token := Token{
		AccessToken:  payload.AccessToken,
		RefreshToken: payload.RefreshToken,
		TokenType:    payload.TokenType,
		Scopes:       payload.Scope,
	}
	if payload.ExpiresIn > 0 {
		token.Expiry = time.Now().Add(time.Duration(payload.ExpiresIn) * time.Second)
	}
	return token
}

func decodeDeviceAuthorizationResponse(resp *http.Response) (DeviceAuthorization, error) {
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return DeviceAuthorization{}, decodeErrorResponse("device authorization", resp.StatusCode, body)
	}

	var auth DeviceAuthorization
	if err := json.NewDecoder(resp.Body).Decode(&auth); err != nil {
		return DeviceAuthorization{}, err
	}
	auth.IssuedAt = time.Now()
	return auth, nil
}

func deviceAuthorizationExpiresAt(auth DeviceAuthorization) time.Time {
	if auth.ExpiresIn <= 0 {
		return time.Time{}
	}
	if !auth.IssuedAt.IsZero() {
		return auth.IssuedAt.Add(time.Duration(auth.ExpiresIn) * time.Second)
	}
	return time.Now().Add(time.Duration(auth.ExpiresIn) * time.Second)
}

func decodeValidatedTokenResponse(resp *http.Response) (ValidatedToken, error) {
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		err := decodeErrorResponse("validate", resp.StatusCode, body)
		if resp.StatusCode == http.StatusUnauthorized {
			return ValidatedToken{}, fmt.Errorf("%w: %v", ErrValidateUnauthorized, err)
		}
		return ValidatedToken{}, err
	}

	var validated ValidatedToken
	if err := json.NewDecoder(resp.Body).Decode(&validated); err != nil {
		return ValidatedToken{}, err
	}
	return validated, nil
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
	return decodeValidatedTokenResponse(resp)
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
		return decodeErrorResponse("revoke", resp.StatusCode, body)
	}
	return nil
}

func decodeErrorResponse(endpoint string, statusCode int, body []byte) error {
	resp := &ErrorResponse{
		Endpoint:   endpoint,
		StatusCode: statusCode,
		Body:       body,
	}

	var payload struct {
		Status  int    `json:"status"`
		Error   string `json:"error"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &payload); err == nil {
		resp.Status = payload.Status
		resp.ErrorCode = payload.Error
		resp.Message = payload.Message
	}
	return resp
}

func isPendingDeviceAuthorization(err error) bool {
	var oauthErr *ErrorResponse
	return errors.As(err, &oauthErr) && (oauthErr.Message == "authorization_pending" || oauthErr.ErrorCode == "authorization_pending")
}

func isSlowDownDeviceAuthorization(err error) bool {
	var oauthErr *ErrorResponse
	return errors.As(err, &oauthErr) && (oauthErr.Message == "slow_down" || oauthErr.ErrorCode == "slow_down")
}

func isExpiredDeviceCodeError(err *ErrorResponse) bool {
	return err.Message == "expired_token" || err.ErrorCode == "expired_token"
}
