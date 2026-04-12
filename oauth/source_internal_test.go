package oauth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestValidatingSourceWaiterReloadsTokenAfterInvalidation(t *testing.T) {
	var validateCalls atomic.Int32
	started := make(chan struct{}, 1)
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		if r.URL.Path != "/validate" {
			t.Fatalf("path = %q, want %q", r.URL.Path, "/validate")
		}

		call := validateCalls.Add(1)
		if call == 1 {
			started <- struct{}{}
			<-release
		}

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

	client := NewClient(Config{
		ClientID: "client-id",
		BaseURL:  server.URL,
	})

	next := &rotatingSource{current: "token-v1"}
	source := NewValidatingSource(client, next)

	type result struct {
		token Token
		err   error
	}

	firstResult := make(chan result, 1)
	go func() {
		token, err := source.Token(context.Background())
		firstResult <- result{token: token, err: err}
	}()

	<-started

	secondResult := make(chan result, 1)
	go func() {
		token, err := source.Token(context.Background())
		secondResult <- result{token: token, err: err}
	}()

	deadline := time.Now().Add(time.Second)
	for {
		source.mu.Lock()
		waiting := source.validating && len(source.waiters) == 1
		source.mu.Unlock()
		if waiting {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("timed out waiting for validating source waiter")
		}
		time.Sleep(10 * time.Millisecond)
	}

	if ok := source.InvalidateToken(context.Background(), "token-v1"); !ok {
		t.Fatal("InvalidateToken() = false, want true")
	}

	close(release)

	first := <-firstResult
	if first.err != nil {
		t.Fatalf("first Token() error = %v", first.err)
	}
	if got := first.token.AccessToken; got != "token-v1" {
		t.Fatalf("first AccessToken = %q, want %q", got, "token-v1")
	}

	second := <-secondResult
	if second.err != nil {
		t.Fatalf("second Token() error = %v", second.err)
	}
	if got := second.token.AccessToken; got != "token-v2" {
		t.Fatalf("second AccessToken = %q, want %q", got, "token-v2")
	}
	if got := validateCalls.Load(); got != 2 {
		t.Fatalf("validate endpoint calls = %d, want 2", got)
	}
}

type rotatingSource struct {
	mu      sync.Mutex
	current string
}

func (s *rotatingSource) Token(context.Context) (Token, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return Token{AccessToken: s.current}, nil
}

func (s *rotatingSource) InvalidateToken(_ context.Context, token string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if token != s.current {
		return false
	}
	s.current = "token-v2"
	return true
}
