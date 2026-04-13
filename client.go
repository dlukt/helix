package helix

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/dlukt/helix/oauth"
)

const defaultBaseURL = "https://api.twitch.tv/helix"

// Config configures a Twitch Helix client.
type Config struct {
	ClientID    string
	HTTPClient  *http.Client
	TokenSource oauth.TokenSource
	Authorizer  RequestAuthorizer
	BaseURL     string
	UserAgent   string
}

// RequestAuthorizer applies authorization or auth-related request headers.
type RequestAuthorizer interface {
	Authorize(context.Context, *http.Request) error
}

// Client is a Twitch Helix client.
type Client struct {
	clientID    string
	httpClient  *http.Client
	tokenSource oauth.TokenSource
	authorizer  RequestAuthorizer
	baseURL     string
	userAgent   string

	Users      *UsersService
	Channels   *ChannelsService
	Streams    *StreamsService
	Chat       *ChatService
	Moderation *ModerationService
	Raids      *RaidsService
	Schedule   *ScheduleService
	Markers    *MarkersService
	Games      *GamesService
	Search     *SearchService
	Clips      *ClipsService
	Videos     *VideosService
	EventSub   *EventSubService
}

// New creates a new Helix client.
func New(cfg Config) (*Client, error) {
	if strings.TrimSpace(cfg.ClientID) == "" {
		return nil, fmt.Errorf("helix: client id is required")
	}
	if strings.TrimSpace(cfg.BaseURL) == "" {
		cfg.BaseURL = defaultBaseURL
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = http.DefaultClient
	}

	c := &Client{
		clientID:    cfg.ClientID,
		httpClient:  cfg.HTTPClient,
		tokenSource: cfg.TokenSource,
		authorizer:  cfg.Authorizer,
		baseURL:     strings.TrimRight(cfg.BaseURL, "/"),
		userAgent:   cfg.UserAgent,
	}
	if c.authorizer == nil && cfg.TokenSource != nil {
		c.authorizer = tokenSourceAuthorizer{source: cfg.TokenSource}
	}
	c.Users = &UsersService{client: c}
	c.Channels = &ChannelsService{client: c}
	c.Streams = &StreamsService{client: c}
	c.Chat = &ChatService{client: c}
	c.Moderation = &ModerationService{client: c}
	c.Raids = &RaidsService{client: c}
	c.Schedule = &ScheduleService{client: c}
	c.Markers = &MarkersService{client: c}
	c.Games = &GamesService{client: c}
	c.Search = &SearchService{client: c}
	c.Clips = &ClipsService{client: c}
	c.Videos = &VideosService{client: c}
	c.EventSub = &EventSubService{client: c}
	return c, nil
}

// RawRequest represents a low-level Helix request.
type RawRequest struct {
	Method string
	Path   string
	Query  url.Values
	Header http.Header
	Body   any
}

// Response wraps shared API response metadata.
type Response struct {
	StatusCode int
	RequestID  string
	Pagination Pagination
	RateLimit  RateLimit
	Header     http.Header
	// Raw contains a replayable copy of the HTTP response, including the body.
	Raw *http.Response
}

// Pagination represents Twitch cursor pagination metadata.
type Pagination struct {
	Cursor string `json:"cursor"`
}

// RateLimit contains parsed rate-limit headers.
type RateLimit struct {
	Limit     int
	Remaining int
	ResetAt   time.Time
}

type envelope struct {
	Data       json.RawMessage `json:"data"`
	Pagination Pagination      `json:"pagination"`
}

// Do executes a low-level Helix request and decodes the JSON response into out.
func (c *Client) Do(ctx context.Context, req RawRequest, out any) (*Response, error) {
	httpResp, meta, err := c.doRaw(ctx, req)
	if err != nil {
		return meta, err
	}
	body, err := replayableBody(httpResp)
	if err != nil {
		return meta, err
	}

	if len(bytes.TrimSpace(body)) == 0 {
		return meta, nil
	}

	var paginationOnly struct {
		Pagination Pagination `json:"pagination"`
	}
	if err := json.Unmarshal(body, &paginationOnly); err == nil {
		meta.Pagination = paginationOnly.Pagination
	}

	if out == nil {
		return meta, nil
	}

	if err := json.Unmarshal(body, out); err != nil {
		return meta, err
	}

	return meta, nil
}

func (c *Client) doData(ctx context.Context, req RawRequest, out any) (*Response, error) {
	httpResp, meta, err := c.doRaw(ctx, req)
	if err != nil {
		return meta, err
	}
	bodyBytes, err := replayableBody(httpResp)
	if err != nil {
		return meta, err
	}

	var body envelope
	if err := json.Unmarshal(bodyBytes, &body); err != nil {
		return meta, err
	}
	meta.Pagination = body.Pagination

	if out == nil {
		return meta, nil
	}
	if err := json.Unmarshal(body.Data, out); err != nil {
		return meta, err
	}
	return meta, nil
}

func (c *Client) newRequest(ctx context.Context, req RawRequest) (*http.Request, error) {
	var bodyReader io.Reader
	if req.Body != nil {
		buf := &bytes.Buffer{}
		if err := json.NewEncoder(buf).Encode(req.Body); err != nil {
			return nil, err
		}
		bodyReader = buf
	}

	httpReq, err := http.NewRequestWithContext(ctx, req.Method, c.baseURL+req.Path, bodyReader)
	if err != nil {
		return nil, err
	}
	if req.Query != nil {
		httpReq.URL.RawQuery = req.Query.Encode()
	}
	httpReq.Header.Set("Client-Id", c.clientID)
	httpReq.Header.Set("Accept", "application/json")
	for name, values := range req.Header {
		httpReq.Header.Del(name)
		for _, value := range values {
			httpReq.Header.Add(name, value)
		}
	}
	if c.userAgent != "" {
		httpReq.Header.Set("User-Agent", c.userAgent)
	}
	if req.Body != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}
	if c.authorizer != nil {
		if err := c.authorizer.Authorize(ctx, httpReq); err != nil {
			return nil, err
		}
	}
	return httpReq, nil
}

func responseFromHTTP(httpResp *http.Response) *Response {
	meta := &Response{
		StatusCode: httpResp.StatusCode,
		RequestID:  httpResp.Header.Get("Request-Id"),
		Header:     httpResp.Header.Clone(),
		Raw:        httpResp,
	}
	meta.RateLimit.Limit, _ = strconv.Atoi(httpResp.Header.Get("Ratelimit-Limit"))
	meta.RateLimit.Remaining, _ = strconv.Atoi(httpResp.Header.Get("Ratelimit-Remaining"))
	if unixSeconds, err := strconv.ParseInt(httpResp.Header.Get("Ratelimit-Reset"), 10, 64); err == nil && unixSeconds > 0 {
		meta.RateLimit.ResetAt = time.Unix(unixSeconds, 0)
	}
	return meta
}

func replayableBody(httpResp *http.Response) ([]byte, error) {
	body, err := io.ReadAll(httpResp.Body)
	closeErr := httpResp.Body.Close()
	if err != nil {
		return nil, err
	}
	if closeErr != nil {
		return nil, closeErr
	}
	httpResp.Body = io.NopCloser(bytes.NewReader(body))
	return body, nil
}

func (c *Client) doRaw(ctx context.Context, req RawRequest) (*http.Response, *Response, error) {
	for attempt := range 2 {
		httpReq, err := c.newRequest(ctx, req)
		if err != nil {
			return nil, nil, err
		}

		httpResp, err := c.httpClient.Do(httpReq)
		if err != nil {
			return nil, nil, err
		}

		meta := responseFromHTTP(httpResp)
		if httpResp.StatusCode == http.StatusServiceUnavailable && attempt == 0 && isIdempotentMethod(req.Method) {
			_, _ = io.Copy(io.Discard, httpResp.Body)
			httpResp.Body.Close()
			continue
		}
		if httpResp.StatusCode == http.StatusUnauthorized && attempt == 0 {
			accessToken := bearerToken(httpReq.Header.Get("Authorization"))
			invalidating, ok := c.tokenSource.(oauth.InvalidatingTokenSource)
			if ok && accessToken != "" && invalidating.InvalidateToken(ctx, accessToken) {
				_, _ = io.Copy(io.Discard, httpResp.Body)
				httpResp.Body.Close()
				continue
			}
		}
		if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
			body, err := replayableBody(httpResp)
			if err != nil {
				return nil, meta, err
			}
			return nil, meta, decodeAPIError(httpResp.StatusCode, body, meta.RateLimit)
		}
		return httpResp, meta, nil
	}

	return nil, nil, fmt.Errorf("helix: exhausted retries")
}

// APIError is returned for non-2xx API responses.
type APIError struct {
	StatusCode int
	ErrorCode  string
	Message    string
	Body       []byte
	RateLimit  RateLimit
}

func (e *APIError) Error() string {
	if e.ErrorCode != "" && e.Message != "" {
		return fmt.Sprintf("helix: api error %d %s: %s", e.StatusCode, e.ErrorCode, e.Message)
	}
	if e.Message != "" {
		return fmt.Sprintf("helix: api error %d: %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("helix: api error %d", e.StatusCode)
}

type tokenSourceAuthorizer struct {
	source oauth.TokenSource
}

func (a tokenSourceAuthorizer) Authorize(ctx context.Context, req *http.Request) error {
	token, err := a.source.Token(ctx)
	if err != nil {
		return err
	}
	if token.AccessToken != "" {
		req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	}
	return nil
}

func decodeAPIError(statusCode int, body []byte, rateLimit RateLimit) *APIError {
	apiErr := &APIError{
		StatusCode: statusCode,
		Body:       body,
		RateLimit:  rateLimit,
	}

	var payload struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &payload); err == nil {
		apiErr.ErrorCode = payload.Error
		apiErr.Message = payload.Message
	}
	return apiErr
}

func bearerToken(authorization string) string {
	const prefix = "Bearer "
	if !strings.HasPrefix(authorization, prefix) {
		return ""
	}
	return strings.TrimPrefix(authorization, prefix)
}

func isIdempotentMethod(method string) bool {
	switch strings.ToUpper(method) {
	case http.MethodGet, http.MethodHead, http.MethodPut, http.MethodDelete, http.MethodOptions, http.MethodTrace:
		return true
	default:
		return false
	}
}
