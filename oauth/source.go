package oauth

import (
	"context"
	"errors"
	"sync"
	"time"
)

// TokenStore persists user tokens across refreshes.
type TokenStore interface {
	Load(context.Context) (Token, error)
	Save(context.Context, Token) error
}

// AppSource lazily acquires and caches an app access token.
type AppSource struct {
	client *Client
	scopes []string

	mu    sync.Mutex
	token Token
}

// NewAppSource creates a token source backed by the client credentials flow.
func NewAppSource(client *Client, scopes []string) *AppSource {
	return &AppSource{client: client, scopes: append([]string(nil), scopes...)}
}

// Token returns a cached app token or exchanges client credentials.
func (s *AppSource) Token(ctx context.Context) (Token, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !needsRefresh(s.token) {
		return s.token, nil
	}

	token, err := s.client.ExchangeClientCredentials(ctx, s.scopes)
	if err != nil {
		return Token{}, err
	}
	s.token = token
	return token, nil
}

// InvalidateToken clears the cached app token so the next request reacquires it.
func (s *AppSource) InvalidateToken(_ context.Context, accessToken string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.token.AccessToken != "" && s.token.AccessToken != accessToken {
		return true
	}
	s.token = Token{}
	return true
}

// RefreshingUserSource keeps a user token fresh via refresh token rotation.
type RefreshingUserSource struct {
	client *Client
	store  TokenStore

	mu           sync.Mutex
	forceRefresh bool
}

// NewRefreshingUserSource creates a refresh-capable user token source.
func NewRefreshingUserSource(client *Client, store TokenStore) *RefreshingUserSource {
	return &RefreshingUserSource{client: client, store: store}
}

// Token returns a cached token or refreshes it when needed.
func (s *RefreshingUserSource) Token(ctx context.Context) (Token, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	token, err := s.store.Load(ctx)
	if err != nil {
		return Token{}, err
	}
	if !s.forceRefresh && !needsRefresh(token) {
		return token, nil
	}

	refreshed, err := s.client.RefreshToken(ctx, token.RefreshToken)
	if err != nil {
		return Token{}, err
	}
	if err := s.store.Save(ctx, refreshed); err != nil {
		return Token{}, err
	}
	s.forceRefresh = false
	return refreshed, nil
}

// InvalidateToken forces a refresh on the next Token call.
func (s *RefreshingUserSource) InvalidateToken(ctx context.Context, accessToken string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	current, err := s.store.Load(ctx)
	if err == nil && current.AccessToken != "" && current.AccessToken != accessToken {
		return true
	}
	s.forceRefresh = true
	return true
}

// ValidatingSource decorates another token source and validates at most hourly.
type ValidatingSource struct {
	client *Client
	next   TokenSource

	mu           sync.Mutex
	lastToken    string
	lastValidate time.Time
	validating   bool
	waiters      []chan struct{}
}

// NewValidatingSource wraps another token source with /validate checks.
func NewValidatingSource(client *Client, next TokenSource) *ValidatingSource {
	return &ValidatingSource{client: client, next: next}
}

// Token returns the downstream token after performing a periodic validation.
func (s *ValidatingSource) Token(ctx context.Context) (Token, error) {
	retriedUnauthorized := false

	for {
		token, err := s.next.Token(ctx)
		if err != nil {
			return Token{}, err
		}

		s.mu.Lock()
		shouldValidate := token.AccessToken != "" && (token.AccessToken != s.lastToken || time.Since(s.lastValidate) >= time.Hour)
		if !shouldValidate {
			s.mu.Unlock()
			return token, nil
		}
		if s.validating {
			wait := make(chan struct{})
			s.waiters = append(s.waiters, wait)
			s.mu.Unlock()
			select {
			case <-ctx.Done():
				return Token{}, ctx.Err()
			case <-wait:
				continue
			}
		}
		s.validating = true
		s.mu.Unlock()

		_, err = s.client.ValidateToken(ctx, token.AccessToken)
		retry := false
		if err != nil && !retriedUnauthorized && token.AccessToken != "" && errors.Is(err, ErrValidateUnauthorized) {
			invalidating, ok := s.next.(InvalidatingTokenSource)
			if ok && invalidating.InvalidateToken(ctx, token.AccessToken) {
				retriedUnauthorized = true
				retry = true
			}
		}

		s.mu.Lock()
		if err == nil {
			s.lastToken = token.AccessToken
			s.lastValidate = time.Now()
		}
		s.validating = false
		waiters := s.waiters
		s.waiters = nil
		s.mu.Unlock()

		for _, wait := range waiters {
			close(wait)
		}
		if err != nil {
			if retry {
				continue
			}
			return Token{}, err
		}
		return token, nil
	}
}

// InvalidateToken clears validation state and forwards invalidation when supported downstream.
func (s *ValidatingSource) InvalidateToken(ctx context.Context, accessToken string) bool {
	s.mu.Lock()
	s.lastToken = ""
	s.lastValidate = time.Time{}
	s.mu.Unlock()

	invalidating, ok := s.next.(InvalidatingTokenSource)
	if !ok {
		return false
	}
	return invalidating.InvalidateToken(ctx, accessToken)
}

func needsRefresh(token Token) bool {
	if token.AccessToken == "" {
		return true
	}
	if token.Expiry.IsZero() {
		return false
	}
	return !time.Now().Before(token.Expiry.Add(-30 * time.Second))
}
