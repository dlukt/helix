package oauth_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dlukt/helix/oauth"
)

func TestAppSourceCachesClientCredentialsTokenUntilExpiry(t *testing.T) {
	t.Parallel()

	var tokenCalls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		if r.URL.Path != "/token" {
			t.Fatalf("path = %q, want %q", r.URL.Path, "/token")
		}
		tokenCalls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "app-token",
			"expires_in":   3600,
			"token_type":   "bearer",
		})
	}))
	defer server.Close()

	client := oauth.NewClient(oauth.Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		BaseURL:      server.URL,
	})

	source := oauth.NewAppSource(client, nil)

	first, err := source.Token(context.Background())
	if err != nil {
		t.Fatalf("first Token() error = %v", err)
	}
	second, err := source.Token(context.Background())
	if err != nil {
		t.Fatalf("second Token() error = %v", err)
	}

	if got := first.AccessToken; got != "app-token" {
		t.Fatalf("first.AccessToken = %q, want %q", got, "app-token")
	}
	if got := second.AccessToken; got != "app-token" {
		t.Fatalf("second.AccessToken = %q, want %q", got, "app-token")
	}
	if got := tokenCalls.Load(); got != 1 {
		t.Fatalf("token endpoint calls = %d, want 1", got)
	}
}

func TestAppSourceInvalidationSucceedsAfterConcurrentRotation(t *testing.T) {
	t.Parallel()

	var tokenCalls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		call := tokenCalls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		switch call {
		case 1:
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token": "token-v1",
				"expires_in":   3600,
				"token_type":   "bearer",
			})
		case 2:
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token": "token-v2",
				"expires_in":   3600,
				"token_type":   "bearer",
			})
		default:
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token": "token-v3",
				"expires_in":   3600,
				"token_type":   "bearer",
			})
		}
	}))
	defer server.Close()

	client := oauth.NewClient(oauth.Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		BaseURL:      server.URL,
	})

	source := oauth.NewAppSource(client, nil)

	first, err := source.Token(context.Background())
	if err != nil {
		t.Fatalf("first Token() error = %v", err)
	}
	if got := first.AccessToken; got != "token-v1" {
		t.Fatalf("first AccessToken = %q, want %q", got, "token-v1")
	}

	if ok := source.InvalidateToken(context.Background(), "token-v1"); !ok {
		t.Fatal("first InvalidateToken() = false, want true")
	}

	rotated, err := source.Token(context.Background())
	if err != nil {
		t.Fatalf("rotated Token() error = %v", err)
	}
	if got := rotated.AccessToken; got != "token-v2" {
		t.Fatalf("rotated AccessToken = %q, want %q", got, "token-v2")
	}

	if ok := source.InvalidateToken(context.Background(), "token-v1"); !ok {
		t.Fatal("second InvalidateToken() = false, want true after token rotation")
	}

	token, err := source.Token(context.Background())
	if err != nil {
		t.Fatalf("Token() error = %v", err)
	}
	if got := token.AccessToken; got != "token-v2" {
		t.Fatalf("AccessToken = %q, want %q", got, "token-v2")
	}
	if got := tokenCalls.Load(); got != 2 {
		t.Fatalf("token endpoint calls = %d, want 2", got)
	}
}

func TestRefreshingUserSourceRefreshesExpiredTokenOnceAndPersistsRotation(t *testing.T) {
	t.Parallel()

	var refreshCalls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		if r.URL.Path != "/token" {
			t.Fatalf("path = %q, want %q", r.URL.Path, "/token")
		}
		refreshCalls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "fresh-token",
			"refresh_token": "rotated-refresh-token",
			"expires_in":    3600,
			"token_type":    "bearer",
		})
	}))
	defer server.Close()

	client := oauth.NewClient(oauth.Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		BaseURL:      server.URL,
	})

	store := &memoryStore{
		token: oauth.Token{
			AccessToken:  "expired-token",
			RefreshToken: "stale-refresh-token",
			Expiry:       time.Now().Add(-time.Minute),
		},
	}

	source := oauth.NewRefreshingUserSource(client, store)

	var wg sync.WaitGroup
	results := make(chan oauth.Token, 2)
	errs := make(chan error, 2)
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			token, err := source.Token(context.Background())
			if err != nil {
				errs <- err
				return
			}
			results <- token
		}()
	}
	wg.Wait()
	close(results)
	close(errs)

	for err := range errs {
		t.Fatalf("Token() error = %v", err)
	}
	for token := range results {
		if got := token.AccessToken; got != "fresh-token" {
			t.Fatalf("AccessToken = %q, want %q", got, "fresh-token")
		}
		if got := token.RefreshToken; got != "rotated-refresh-token" {
			t.Fatalf("RefreshToken = %q, want %q", got, "rotated-refresh-token")
		}
	}
	if got := refreshCalls.Load(); got != 1 {
		t.Fatalf("refresh endpoint calls = %d, want 1", got)
	}

	saved, err := store.Load(context.Background())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got := saved.RefreshToken; got != "rotated-refresh-token" {
		t.Fatalf("stored RefreshToken = %q, want %q", got, "rotated-refresh-token")
	}
}

func TestRefreshingUserSourceIgnoresStaleTokenInvalidation(t *testing.T) {
	t.Parallel()

	var refreshCalls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		refreshCalls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "fresh-token",
			"refresh_token": "rotated-refresh-token",
			"expires_in":    3600,
			"token_type":    "bearer",
		})
	}))
	defer server.Close()

	client := oauth.NewClient(oauth.Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		BaseURL:      server.URL,
	})

	store := &memoryStore{
		token: oauth.Token{
			AccessToken:  "token-v2",
			RefreshToken: "refresh-token",
			Expiry:       time.Now().Add(time.Hour),
		},
	}
	source := oauth.NewRefreshingUserSource(client, store)

	if ok := source.InvalidateToken(context.Background(), "token-v1"); !ok {
		t.Fatal("InvalidateToken() = false, want true")
	}

	token, err := source.Token(context.Background())
	if err != nil {
		t.Fatalf("Token() error = %v", err)
	}
	if got := token.AccessToken; got != "token-v2" {
		t.Fatalf("AccessToken = %q, want %q", got, "token-v2")
	}
	if got := refreshCalls.Load(); got != 0 {
		t.Fatalf("refresh endpoint calls = %d, want 0", got)
	}
}

func TestValidatingSourceValidatesAtMostHourly(t *testing.T) {
	t.Parallel()

	var validateCalls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		if r.URL.Path != "/validate" {
			t.Fatalf("path = %q, want %q", r.URL.Path, "/validate")
		}
		validateCalls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"client_id":  "client-id",
			"login":      "darko",
			"scopes":     []string{"user:read:email"},
			"user_id":    "123",
			"expires_in": 3600,
		})
	}))
	defer server.Close()

	client := oauth.NewClient(oauth.Config{
		ClientID: "client-id",
		BaseURL:  server.URL,
	})

	source := oauth.NewValidatingSource(client, oauth.StaticSource{
		Value: oauth.Token{AccessToken: "access-token"},
	})

	for range 2 {
		token, err := source.Token(context.Background())
		if err != nil {
			t.Fatalf("Token() error = %v", err)
		}
		if got := token.AccessToken; got != "access-token" {
			t.Fatalf("AccessToken = %q, want %q", got, "access-token")
		}
	}

	if got := validateCalls.Load(); got != 1 {
		t.Fatalf("validate endpoint calls = %d, want 1", got)
	}
}

func TestValidatingSourceCollapsesConcurrentValidation(t *testing.T) {
	t.Parallel()

	var validateCalls atomic.Int32
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		if r.URL.Path != "/validate" {
			t.Fatalf("path = %q, want %q", r.URL.Path, "/validate")
		}
		validateCalls.Add(1)
		<-release
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"client_id":  "client-id",
			"login":      "darko",
			"scopes":     []string{"user:read:email"},
			"user_id":    "123",
			"expires_in": 3600,
		})
	}))
	defer server.Close()

	client := oauth.NewClient(oauth.Config{
		ClientID: "client-id",
		BaseURL:  server.URL,
	})

	source := oauth.NewValidatingSource(client, oauth.StaticSource{
		Value: oauth.Token{AccessToken: "access-token"},
	})

	errs := make(chan error, 4)
	for range 4 {
		go func() {
			_, err := source.Token(context.Background())
			errs <- err
		}()
	}

	time.Sleep(50 * time.Millisecond)
	if got := validateCalls.Load(); got != 1 {
		t.Fatalf("validate endpoint calls before release = %d, want 1", got)
	}
	close(release)

	for range 4 {
		if err := <-errs; err != nil {
			t.Fatalf("Token() error = %v", err)
		}
	}
	if got := validateCalls.Load(); got != 1 {
		t.Fatalf("validate endpoint calls = %d, want 1", got)
	}
}

func TestValidatingSourceRefreshesUnauthorizedTokenFromWrappedSource(t *testing.T) {
	t.Parallel()

	var refreshCalls atomic.Int32
	var validateCalls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		switch r.URL.Path {
		case "/validate":
			call := validateCalls.Add(1)
			if got := r.Header.Get("Authorization"); call == 1 && got != "OAuth stale-token" {
				t.Fatalf("first Authorization = %q, want %q", got, "OAuth stale-token")
			}
			if got := r.Header.Get("Authorization"); call == 2 && got != "OAuth fresh-token" {
				t.Fatalf("second Authorization = %q, want %q", got, "OAuth fresh-token")
			}
			if call == 1 {
				http.Error(w, `{"status":401,"message":"invalid access token"}`, http.StatusUnauthorized)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"client_id":  "client-id",
				"login":      "darko",
				"scopes":     []string{"user:read:email"},
				"user_id":    "123",
				"expires_in": 3600,
			})
		case "/token":
			refreshCalls.Add(1)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token":  "fresh-token",
				"refresh_token": "rotated-refresh-token",
				"expires_in":    3600,
				"token_type":    "bearer",
			})
		default:
			t.Fatalf("path = %q, want /validate or /token", r.URL.Path)
		}
	}))
	defer server.Close()

	client := oauth.NewClient(oauth.Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		BaseURL:      server.URL,
	})

	store := &memoryStore{
		token: oauth.Token{
			AccessToken:  "stale-token",
			RefreshToken: "refresh-token",
			Expiry:       time.Now().Add(time.Hour),
		},
	}
	source := oauth.NewValidatingSource(client, oauth.NewRefreshingUserSource(client, store))

	token, err := source.Token(context.Background())
	if err != nil {
		t.Fatalf("Token() error = %v", err)
	}
	if got := token.AccessToken; got != "fresh-token" {
		t.Fatalf("AccessToken = %q, want %q", got, "fresh-token")
	}
	if got := token.RefreshToken; got != "rotated-refresh-token" {
		t.Fatalf("RefreshToken = %q, want %q", got, "rotated-refresh-token")
	}
	if got := refreshCalls.Load(); got != 1 {
		t.Fatalf("refresh endpoint calls = %d, want 1", got)
	}
	if got := validateCalls.Load(); got != 2 {
		t.Fatalf("validate endpoint calls = %d, want 2", got)
	}

	saved, err := store.Load(context.Background())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got := saved.AccessToken; got != "fresh-token" {
		t.Fatalf("stored AccessToken = %q, want %q", got, "fresh-token")
	}
}

type memoryStore struct {
	mu    sync.Mutex
	token oauth.Token
}

func (s *memoryStore) Load(context.Context) (oauth.Token, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.token, nil
}

func (s *memoryStore) Save(_ context.Context, token oauth.Token) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.token = token
	return nil
}
