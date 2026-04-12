package oauth

import (
	"context"
	"time"
)

// Token represents a Twitch OAuth access token.
type Token struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	Scopes       []string
	Expiry       time.Time
}

// ValidatedToken is returned by the /validate endpoint.
type ValidatedToken struct {
	ClientID  string   `json:"client_id"`
	Login     string   `json:"login"`
	Scopes    []string `json:"scopes"`
	UserID    string   `json:"user_id"`
	ExpiresIn int      `json:"expires_in"`
}

// TokenSource returns the token used to authorize API requests.
type TokenSource interface {
	Token(context.Context) (Token, error)
}

// InvalidatingTokenSource can discard a token that the server rejected and force re-acquisition.
type InvalidatingTokenSource interface {
	TokenSource
	InvalidateToken(context.Context, string) bool
}

// StaticSource always returns the configured token.
type StaticSource struct {
	Value Token
}

// Token returns the configured token.
func (s StaticSource) Token(context.Context) (Token, error) {
	return s.Value, nil
}
