package helix_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dlukt/helix"
	"github.com/dlukt/helix/oauth"
)

func TestClientRetriesOne503BeforeSucceeding(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		if calls.Add(1) == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"error":"service unavailable"}`))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"id": "123", "login": "darko", "display_name": "Darko"},
			},
		})
	}))
	defer server.Close()

	client, err := helix.New(helix.Config{
		ClientID: "client-id",
		BaseURL:  server.URL,
		TokenSource: oauth.StaticSource{
			Value: oauth.Token{AccessToken: "access-token"},
		},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	resp, _, err := client.Users.Get(context.Background(), helix.GetUsersParams{IDs: []string{"123"}})
	if err != nil {
		t.Fatalf("Users.Get() error = %v", err)
	}
	if got := resp.Data[0].ID; got != "123" {
		t.Fatalf("resp.Data[0].ID = %q, want %q", got, "123")
	}
	if got := calls.Load(); got != 2 {
		t.Fatalf("calls = %d, want 2", got)
	}
}

func TestClientReturnsRateLimitedAPIErrorWithResetTime(t *testing.T) {
	t.Parallel()

	resetAt := time.Unix(1712846400, 0)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		w.Header().Set("Ratelimit-Limit", "800")
		w.Header().Set("Ratelimit-Remaining", "0")
		w.Header().Set("Ratelimit-Reset", "1712846400")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":"Too Many Requests","status":429,"message":"rate limited"}`))
	}))
	defer server.Close()

	client, err := helix.New(helix.Config{
		ClientID: "client-id",
		BaseURL:  server.URL,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, _, err = client.Users.Get(context.Background(), helix.GetUsersParams{IDs: []string{"123"}})
	if err == nil {
		t.Fatal("Users.Get() error = nil, want APIError")
	}

	apiErr, ok := err.(*helix.APIError)
	if !ok {
		t.Fatalf("error type = %T, want *helix.APIError", err)
	}
	if got := apiErr.StatusCode; got != http.StatusTooManyRequests {
		t.Fatalf("StatusCode = %d, want %d", got, http.StatusTooManyRequests)
	}
	if got := apiErr.ErrorCode; got != "Too Many Requests" {
		t.Fatalf("ErrorCode = %q, want %q", got, "Too Many Requests")
	}
	if got := apiErr.Message; got != "rate limited" {
		t.Fatalf("Message = %q, want %q", got, "rate limited")
	}
	if got := apiErr.RateLimit.ResetAt; !got.Equal(resetAt) {
		t.Fatalf("RateLimit.ResetAt = %s, want %s", got, resetAt)
	}
}

func TestClientPropagatesRequestIDAndCustomHeaders(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		if got := r.Header.Get("X-Custom-Header"); got != "custom-value" {
			t.Fatalf("X-Custom-Header = %q, want %q", got, "custom-value")
		}
		w.Header().Set("Request-Id", "request-123")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"id": "123", "login": "darko", "display_name": "Darko"},
			},
		})
	}))
	defer server.Close()

	client, err := helix.New(helix.Config{
		ClientID: "client-id",
		BaseURL:  server.URL,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	var data []helix.User
	meta, err := client.Do(context.Background(), helix.RawRequest{
		Method: http.MethodGet,
		Path:   "/users",
		Header: http.Header{"X-Custom-Header": []string{"custom-value"}},
	}, &struct {
		Data []helix.User `json:"data"`
	}{Data: data})
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	if got := meta.RequestID; got != "request-123" {
		t.Fatalf("RequestID = %q, want %q", got, "request-123")
	}
}

func TestClientReturnsAPIErrorForNonJSONErrorBodies(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("temporary upstream failure"))
	}))
	defer server.Close()

	client, err := helix.New(helix.Config{
		ClientID: "client-id",
		BaseURL:  server.URL,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, _, err = client.Users.Get(context.Background(), helix.GetUsersParams{IDs: []string{"123"}})
	if err == nil {
		t.Fatal("Users.Get() error = nil, want APIError")
	}
	apiErr, ok := err.(*helix.APIError)
	if !ok {
		t.Fatalf("error type = %T, want *helix.APIError", err)
	}
	if got := apiErr.ErrorCode; got != "" {
		t.Fatalf("ErrorCode = %q, want empty", got)
	}
	if got := string(apiErr.Body); got != "temporary upstream failure" {
		t.Fatalf("Body = %q, want temporary upstream failure", got)
	}
}

func TestClientReturnsAPIErrorForMalformedJSONErrorBodies(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":`))
	}))
	defer server.Close()

	client, err := helix.New(helix.Config{
		ClientID: "client-id",
		BaseURL:  server.URL,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, _, err = client.Users.Get(context.Background(), helix.GetUsersParams{IDs: []string{"123"}})
	if err == nil {
		t.Fatal("Users.Get() error = nil, want APIError")
	}
	apiErr, ok := err.(*helix.APIError)
	if !ok {
		t.Fatalf("error type = %T, want *helix.APIError", err)
	}
	if got := apiErr.ErrorCode; got != "" {
		t.Fatalf("ErrorCode = %q, want empty", got)
	}
	if got := apiErr.Message; got != "" {
		t.Fatalf("Message = %q, want empty", got)
	}
	if got := string(apiErr.Body); got != `{"error":` {
		t.Fatalf("Body = %q, want malformed JSON body", got)
	}
}

func TestClientRetries401WithInvalidatingTokenSource(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		switch got := r.Header.Get("Authorization"); got {
		case "Bearer stale-token":
			calls.Add(1)
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"Unauthorized","status":401}`))
		case "Bearer fresh-token":
			calls.Add(1)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{
					{"id": "123", "login": "darko", "display_name": "Darko"},
				},
			})
		default:
			t.Fatalf("Authorization = %q, want stale or fresh token", got)
		}
	}))
	defer server.Close()

	source := &rotatingTokenSource{current: "stale-token"}
	client, err := helix.New(helix.Config{
		ClientID:    "client-id",
		BaseURL:     server.URL,
		TokenSource: source,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	resp, _, err := client.Users.Get(context.Background(), helix.GetUsersParams{IDs: []string{"123"}})
	if err != nil {
		t.Fatalf("Users.Get() error = %v", err)
	}
	if got := resp.Data[0].ID; got != "123" {
		t.Fatalf("resp.Data[0].ID = %q, want %q", got, "123")
	}
	if got := calls.Load(); got != 2 {
		t.Fatalf("calls = %d, want 2", got)
	}
	if got := source.invalidations.Load(); got != 1 {
		t.Fatalf("invalidations = %d, want 1", got)
	}
}

func TestClientReturnsResponseMetadataOnDecodeError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		w.Header().Set("Ratelimit-Limit", "800")
		w.Header().Set("Ratelimit-Remaining", "799")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":[`))
	}))
	defer server.Close()

	client, err := helix.New(helix.Config{
		ClientID: "client-id",
		BaseURL:  server.URL,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, meta, err := client.Users.Get(context.Background(), helix.GetUsersParams{IDs: []string{"123"}})
	if err == nil {
		t.Fatal("Users.Get() error = nil, want decode error")
	}
	if meta == nil {
		t.Fatal("meta = nil, want response metadata on decode error")
	}
	if got := meta.StatusCode; got != http.StatusOK {
		t.Fatalf("StatusCode = %d, want %d", got, http.StatusOK)
	}
	if got := meta.RateLimit.Limit; got != 800 {
		t.Fatalf("RateLimit.Limit = %d, want 800", got)
	}
}

func TestClientDoDecodesFullResponseEnvelope(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{
					"id":      "sub-1",
					"status":  "enabled",
					"type":    "channel.follow",
					"version": "2",
				},
			},
			"pagination": map[string]any{
				"cursor": "next-page",
			},
			"total":          1,
			"total_cost":     1,
			"max_total_cost": 10000,
		})
	}))
	defer server.Close()

	client, err := helix.New(helix.Config{
		ClientID: "client-id",
		BaseURL:  server.URL,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	var out helix.ListEventSubSubscriptionsResponse
	meta, err := client.Do(context.Background(), helix.RawRequest{
		Method: http.MethodGet,
		Path:   "/eventsub/subscriptions",
	}, &out)
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	if got := out.Total; got != 1 {
		t.Fatalf("out.Total = %d, want 1", got)
	}
	if got := len(out.Data); got != 1 {
		t.Fatalf("len(out.Data) = %d, want 1", got)
	}
	if got := meta.Pagination.Cursor; got != "next-page" {
		t.Fatalf("meta.Pagination.Cursor = %q, want %q", got, "next-page")
	}
}

func TestClientDoAllowsEmptySuccessfulBody(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := helix.New(helix.Config{
		ClientID: "client-id",
		BaseURL:  server.URL,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	out := struct {
		Status string `json:"status"`
	}{Status: "unchanged"}
	meta, err := client.Do(context.Background(), helix.RawRequest{
		Method: http.MethodDelete,
		Path:   "/empty",
	}, &out)
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	if meta == nil {
		t.Fatal("meta = nil, want response metadata")
	}
	if got := meta.StatusCode; got != http.StatusNoContent {
		t.Fatalf("StatusCode = %d, want %d", got, http.StatusNoContent)
	}
	if got := out.Status; got != "unchanged" {
		t.Fatalf("out.Status = %q, want %q", got, "unchanged")
	}
}

func TestClientDoParsesPaginationWhenOutputIsNil(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"id": "sub-1", "status": "enabled"},
			},
			"pagination": map[string]any{
				"cursor": "next-page",
			},
		})
	}))
	defer server.Close()

	client, err := helix.New(helix.Config{
		ClientID: "client-id",
		BaseURL:  server.URL,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	meta, err := client.Do(context.Background(), helix.RawRequest{
		Method: http.MethodGet,
		Path:   "/eventsub/subscriptions",
	}, nil)
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	if meta == nil {
		t.Fatal("meta = nil, want response metadata")
	}
	if got := meta.Pagination.Cursor; got != "next-page" {
		t.Fatalf("meta.Pagination.Cursor = %q, want %q", got, "next-page")
	}
}

func TestClientDoPreservesReplayableRawResponseBody(t *testing.T) {
	t.Parallel()

	const rawBody = `{"data":[{"id":"sub-1","status":"enabled"}],"pagination":{"cursor":"next-page"}}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(rawBody))
	}))
	defer server.Close()

	client, err := helix.New(helix.Config{
		ClientID: "client-id",
		BaseURL:  server.URL,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	var out helix.ListEventSubSubscriptionsResponse
	meta, err := client.Do(context.Background(), helix.RawRequest{
		Method: http.MethodGet,
		Path:   "/eventsub/subscriptions",
	}, &out)
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	if meta == nil || meta.Raw == nil {
		t.Fatal("meta.Raw = nil, want raw response")
	}

	got, err := io.ReadAll(meta.Raw.Body)
	if err != nil {
		t.Fatalf("ReadAll(meta.Raw.Body) error = %v", err)
	}
	if string(got) != rawBody {
		t.Fatalf("meta.Raw.Body = %q, want %q", string(got), rawBody)
	}
}

func TestClientDoDataPreservesReplayableRawResponseBody(t *testing.T) {
	t.Parallel()

	const rawBody = `{"data":[{"id":"123","login":"darko","display_name":"Darko"}]}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(rawBody))
	}))
	defer server.Close()

	client, err := helix.New(helix.Config{
		ClientID: "client-id",
		BaseURL:  server.URL,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, meta, err := client.Users.Get(context.Background(), helix.GetUsersParams{IDs: []string{"123"}})
	if err != nil {
		t.Fatalf("Users.Get() error = %v", err)
	}
	if meta == nil || meta.Raw == nil {
		t.Fatal("meta.Raw = nil, want raw response")
	}

	got, err := io.ReadAll(meta.Raw.Body)
	if err != nil {
		t.Fatalf("ReadAll(meta.Raw.Body) error = %v", err)
	}
	if string(got) != rawBody {
		t.Fatalf("meta.Raw.Body = %q, want %q", string(got), rawBody)
	}
}

type rotatingTokenSource struct {
	current       string
	invalidations atomic.Int32
}

func (s *rotatingTokenSource) Token(context.Context) (oauth.Token, error) {
	return oauth.Token{AccessToken: s.current}, nil
}

func (s *rotatingTokenSource) InvalidateToken(_ context.Context, token string) bool {
	if token != s.current {
		return false
	}
	s.current = "fresh-token"
	s.invalidations.Add(1)
	return true
}
