package oauth_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/dlukt/helix/oauth"
)

func TestClientAuthorizationURLBuildsDocumentedQuery(t *testing.T) {
	t.Parallel()

	client := oauth.NewClient(oauth.Config{
		ClientID: "client-id",
		BaseURL:  "https://id.twitch.tv/oauth2",
	})

	authURL := client.AuthorizationURL(oauth.AuthorizationURLParams{
		RedirectURI: "https://example.com/callback",
		Scopes:      []string{"user:read:email", "channel:read:ads"},
		State:       "opaque-state",
		ForceVerify: true,
	})

	parsed, err := url.Parse(authURL)
	if err != nil {
		t.Fatalf("Parse(authURL) error = %v", err)
	}

	if got := parsed.Path; got != "/oauth2/authorize" {
		t.Fatalf("path = %q, want %q", got, "/oauth2/authorize")
	}
	if got := parsed.Query().Get("client_id"); got != "client-id" {
		t.Fatalf("client_id = %q, want %q", got, "client-id")
	}
	if got := parsed.Query().Get("response_type"); got != "code" {
		t.Fatalf("response_type = %q, want %q", got, "code")
	}
	if got := parsed.Query().Get("redirect_uri"); got != "https://example.com/callback" {
		t.Fatalf("redirect_uri = %q", got)
	}
	if got := parsed.Query().Get("scope"); got != "user:read:email channel:read:ads" {
		t.Fatalf("scope = %q", got)
	}
	if got := parsed.Query().Get("state"); got != "opaque-state" {
		t.Fatalf("state = %q, want %q", got, "opaque-state")
	}
	if got := parsed.Query().Get("force_verify"); got != "true" {
		t.Fatalf("force_verify = %q, want %q", got, "true")
	}
}

func TestClientExchangeAuthorizationCodePostsFormAndParsesToken(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		if got := r.Method; got != http.MethodPost {
			t.Fatalf("method = %q, want %q", got, http.MethodPost)
		}
		if got := r.URL.Path; got != "/token" {
			t.Fatalf("path = %q, want %q", got, "/token")
		}
		if got := r.Header.Get("Content-Type"); !strings.HasPrefix(got, "application/x-www-form-urlencoded") {
			t.Fatalf("Content-Type = %q", got)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm() error = %v", err)
		}
		if got := r.Form.Get("client_id"); got != "client-id" {
			t.Fatalf("client_id = %q, want %q", got, "client-id")
		}
		if got := r.Form.Get("client_secret"); got != "client-secret" {
			t.Fatalf("client_secret = %q, want %q", got, "client-secret")
		}
		if got := r.Form.Get("code"); got != "auth-code" {
			t.Fatalf("code = %q, want %q", got, "auth-code")
		}
		if got := r.Form.Get("grant_type"); got != "authorization_code" {
			t.Fatalf("grant_type = %q, want %q", got, "authorization_code")
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "new-access-token",
			"refresh_token": "new-refresh-token",
			"expires_in":    3600,
			"scope":         []string{"user:read:email"},
			"token_type":    "bearer",
		})
	}))
	defer server.Close()

	client := oauth.NewClient(oauth.Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		BaseURL:      server.URL,
	})

	token, err := client.ExchangeAuthorizationCode(context.Background(), oauth.AuthorizationCodeExchange{
		Code:        "auth-code",
		RedirectURI: "https://example.com/callback",
	})
	if err != nil {
		t.Fatalf("ExchangeAuthorizationCode() error = %v", err)
	}

	if got := token.AccessToken; got != "new-access-token" {
		t.Fatalf("AccessToken = %q, want %q", got, "new-access-token")
	}
	if got := token.RefreshToken; got != "new-refresh-token" {
		t.Fatalf("RefreshToken = %q, want %q", got, "new-refresh-token")
	}
	if got := token.TokenType; got != "bearer" {
		t.Fatalf("TokenType = %q, want %q", got, "bearer")
	}
	if got := token.Scopes; len(got) != 1 || got[0] != "user:read:email" {
		t.Fatalf("Scopes = %v", got)
	}
	if got := time.Until(token.Expiry); got < 59*time.Minute || got > 61*time.Minute {
		t.Fatalf("Expiry delta = %s, want about 1h", got)
	}
}

func TestClientValidateAndRevokeUseDocumentedEndpoints(t *testing.T) {
	t.Parallel()

	var revokeSeen bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		switch r.URL.Path {
		case "/validate":
			if got := r.Method; got != http.MethodGet {
				t.Fatalf("validate method = %q, want %q", got, http.MethodGet)
			}
			if got := r.Header.Get("Authorization"); got != "OAuth access-token" {
				t.Fatalf("validate Authorization = %q, want %q", got, "OAuth access-token")
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"client_id":  "client-id",
				"login":      "darko",
				"scopes":     []string{"user:read:email"},
				"user_id":    "123",
				"expires_in": 3600,
			})
		case "/revoke":
			revokeSeen = true
			if got := r.Method; got != http.MethodPost {
				t.Fatalf("revoke method = %q, want %q", got, http.MethodPost)
			}
			if got := r.URL.Query().Get("client_id"); got != "client-id" {
				t.Fatalf("revoke client_id = %q, want %q", got, "client-id")
			}
			if got := r.URL.Query().Get("token"); got != "access-token" {
				t.Fatalf("revoke token = %q, want %q", got, "access-token")
			}
			w.WriteHeader(http.StatusOK)
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	defer server.Close()

	client := oauth.NewClient(oauth.Config{
		ClientID: "client-id",
		BaseURL:  server.URL,
	})

	validated, err := client.ValidateToken(context.Background(), "access-token")
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}
	if got := validated.ClientID; got != "client-id" {
		t.Fatalf("ClientID = %q, want %q", got, "client-id")
	}
	if got := validated.UserID; got != "123" {
		t.Fatalf("UserID = %q, want %q", got, "123")
	}

	if err := client.RevokeToken(context.Background(), "access-token"); err != nil {
		t.Fatalf("RevokeToken() error = %v", err)
	}
	if !revokeSeen {
		t.Fatal("revoke endpoint was not called")
	}
}

func TestClientDeviceCodeFlowUsesDocumentedEndpoints(t *testing.T) {
	t.Parallel()

	var deviceSeen, tokenSeen bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		switch r.URL.Path {
		case "/device":
			deviceSeen = true
			if got := r.Method; got != http.MethodPost {
				t.Fatalf("device method = %q, want %q", got, http.MethodPost)
			}
			if err := r.ParseForm(); err != nil {
				t.Fatalf("ParseForm() error = %v", err)
			}
			if got := r.Form.Get("client_id"); got != "client-id" {
				t.Fatalf("client_id = %q, want %q", got, "client-id")
			}
			if got := r.Form.Get("scopes"); got != "chat:read chat:edit" {
				t.Fatalf("scopes = %q, want %q", got, "chat:read chat:edit")
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"device_code":      "device-code",
				"expires_in":       1800,
				"interval":         5,
				"user_code":        "ABCDEFGH",
				"verification_uri": "https://www.twitch.tv/activate?public=true&device-code=ABCDEFGH",
			})
		case "/token":
			tokenSeen = true
			if got := r.Method; got != http.MethodPost {
				t.Fatalf("token method = %q, want %q", got, http.MethodPost)
			}
			if err := r.ParseForm(); err != nil {
				t.Fatalf("ParseForm() error = %v", err)
			}
			if got := r.Form.Get("grant_type"); got != "urn:ietf:params:oauth:grant-type:device_code" {
				t.Fatalf("grant_type = %q", got)
			}
			if got := r.Form.Get("device_code"); got != "device-code" {
				t.Fatalf("device_code = %q, want %q", got, "device-code")
			}
			if got := r.Form.Get("scopes"); got != "chat:read chat:edit" {
				t.Fatalf("scopes = %q, want %q", got, "chat:read chat:edit")
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token":  "device-access-token",
				"refresh_token": "device-refresh-token",
				"expires_in":    14400,
				"scope":         []string{"chat:read", "chat:edit"},
				"token_type":    "bearer",
			})
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	defer server.Close()

	client := oauth.NewClient(oauth.Config{
		ClientID: "client-id",
		BaseURL:  server.URL,
	})

	deviceCode, err := client.StartDeviceAuthorization(context.Background(), []string{"chat:read", "chat:edit"})
	if err != nil {
		t.Fatalf("StartDeviceAuthorization() error = %v", err)
	}
	if got := deviceCode.DeviceCode; got != "device-code" {
		t.Fatalf("DeviceCode = %q, want %q", got, "device-code")
	}

	token, err := client.ExchangeDeviceCode(context.Background(), oauth.DeviceCodeExchange{
		DeviceCode: deviceCode.DeviceCode,
		Scopes:     []string{"chat:read", "chat:edit"},
	})
	if err != nil {
		t.Fatalf("ExchangeDeviceCode() error = %v", err)
	}
	if got := token.AccessToken; got != "device-access-token" {
		t.Fatalf("AccessToken = %q, want %q", got, "device-access-token")
	}
	if !deviceSeen || !tokenSeen {
		t.Fatalf("deviceSeen=%t tokenSeen=%t, want both true", deviceSeen, tokenSeen)
	}
}

func TestClientRefreshTokenOmitsClientSecretForPublicClients(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		if got := r.URL.Path; got != "/token" {
			t.Fatalf("path = %q, want %q", got, "/token")
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm() error = %v", err)
		}
		if got := r.Form.Get("grant_type"); got != "refresh_token" {
			t.Fatalf("grant_type = %q, want %q", got, "refresh_token")
		}
		if got := r.Form.Get("refresh_token"); got != "refresh-token" {
			t.Fatalf("refresh_token = %q, want %q", got, "refresh-token")
		}
		if _, ok := r.Form["client_secret"]; ok {
			t.Fatalf("client_secret present in public client refresh form: %q", r.Form.Get("client_secret"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "refreshed-access-token",
			"refresh_token": "rotated-refresh-token",
			"expires_in":    3600,
			"token_type":    "bearer",
		})
	}))
	defer server.Close()

	client := oauth.NewClient(oauth.Config{
		ClientID: "public-client-id",
		BaseURL:  server.URL,
	})

	token, err := client.RefreshToken(context.Background(), "refresh-token")
	if err != nil {
		t.Fatalf("RefreshToken() error = %v", err)
	}
	if got := token.AccessToken; got != "refreshed-access-token" {
		t.Fatalf("AccessToken = %q, want %q", got, "refreshed-access-token")
	}
}
