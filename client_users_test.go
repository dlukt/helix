package helix_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dlukt/helix"
	"github.com/dlukt/helix/oauth"
)

func TestUsersGetSendsAuthenticatedRequestAndDecodesResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		if got := r.Method; got != http.MethodGet {
			t.Fatalf("method = %q, want %q", got, http.MethodGet)
		}
		if got := r.URL.Path; got != "/users" {
			t.Fatalf("path = %q, want %q", got, "/users")
		}
		if got := r.URL.Query()["id"]; len(got) != 2 || got[0] != "123" || got[1] != "456" {
			t.Fatalf("query id = %v, want [123 456]", got)
		}
		if got := r.Header.Get("Client-Id"); got != "client-id" {
			t.Fatalf("Client-Id = %q, want %q", got, "client-id")
		}
		if got := r.Header.Get("Authorization"); got != "Bearer access-token" {
			t.Fatalf("Authorization = %q, want %q", got, "Bearer access-token")
		}
		if got := r.Header.Get("User-Agent"); got != "helix-test" {
			t.Fatalf("User-Agent = %q, want %q", got, "helix-test")
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Ratelimit-Limit", "800")
		w.Header().Set("Ratelimit-Remaining", "799")
		w.Header().Set("Ratelimit-Reset", "1712846400")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{
					"id":           "123",
					"login":        "darko",
					"display_name": "Darko",
				},
			},
			"pagination": map[string]any{
				"cursor": "next-cursor",
			},
		}); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	defer server.Close()

	client, err := helix.New(helix.Config{
		ClientID:  "client-id",
		BaseURL:   server.URL,
		UserAgent: "helix-test",
		TokenSource: oauth.StaticSource{
			Value: oauth.Token{AccessToken: "access-token"},
		},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	resp, meta, err := client.Users.Get(context.Background(), helix.GetUsersParams{
		IDs: []string{"123", "456"},
	})
	if err != nil {
		t.Fatalf("Users.Get() error = %v", err)
	}

	if len(resp.Data) != 1 {
		t.Fatalf("len(resp.Data) = %d, want 1", len(resp.Data))
	}
	if got := resp.Data[0].Login; got != "darko" {
		t.Fatalf("resp.Data[0].Login = %q, want %q", got, "darko")
	}
	if got := meta.Pagination.Cursor; got != "next-cursor" {
		t.Fatalf("meta.Pagination.Cursor = %q, want %q", got, "next-cursor")
	}
	if got := meta.RateLimit.Limit; got != 800 {
		t.Fatalf("meta.RateLimit.Limit = %d, want 800", got)
	}
	if got := meta.RateLimit.Remaining; got != 799 {
		t.Fatalf("meta.RateLimit.Remaining = %d, want 799", got)
	}
	if got := meta.StatusCode; got != http.StatusOK {
		t.Fatalf("meta.StatusCode = %d, want %d", got, http.StatusOK)
	}
}

func TestUsersUpdateAndBlocksEncodeRequestsAndDecodeResponses(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		switch r.URL.Path {
		case "/authorization/users":
			if got := r.Method; got != http.MethodGet {
				t.Fatalf("method = %q, want GET", got)
			}
			if got := r.URL.Query()["user_id"]; len(got) != 2 || got[0] != "141981764" || got[1] != "197886470" {
				t.Fatalf("user_id = %v, want [141981764 197886470]", got)
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{
					{
						"user_id":    "141981764",
						"user_name":  "TwitchDev",
						"user_login": "twitchdev",
						"scopes": []string{
							"bits:read",
							"channel:bot",
							"channel:manage:predictions",
						},
					},
					{
						"user_id":    "197886470",
						"user_name":  "TwitchRivals",
						"user_login": "twitchrivals",
						"scopes": []string{
							"channel:manage:predictions",
						},
					},
				},
			})
		case "/users/extensions/list":
			if got := r.Method; got != http.MethodGet {
				t.Fatalf("method = %q, want GET", got)
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{
					{
						"id":           "wi08ebtatdc7oj83wtl9uxwz807l8b",
						"version":      "1.1.8",
						"name":         "Streamlabs Leaderboard",
						"can_activate": true,
						"type":         []string{"panel"},
					},
					{
						"id":           "lqnf3zxk0rv0g7gq92mtmnirjz2cjj",
						"version":      "0.0.1",
						"name":         "Dev Experience Test",
						"can_activate": true,
						"type":         []string{"component", "mobile", "panel", "overlay"},
					},
				},
			})
		case "/users/extensions":
			switch r.Method {
			case http.MethodGet:
				if got := r.URL.Query().Get("user_id"); got != "141981764" {
					t.Fatalf("user_id = %q, want 141981764", got)
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": map[string]any{
						"panel": map[string]any{
							"1": map[string]any{
								"active":  true,
								"id":      "rh6jq1q334hqc2rr1qlzqbvwlfl3x0",
								"version": "1.1.0",
								"name":    "TopClip",
							},
						},
						"overlay": map[string]any{
							"1": map[string]any{
								"active":  true,
								"id":      "zfh2irvx2jb4s60f02jq0ajm8vwgka",
								"version": "1.0.19",
								"name":    "Streamlabs",
							},
						},
						"component": map[string]any{
							"1": map[string]any{
								"active":  true,
								"id":      "lqnf3zxk0rv0g7gq92mtmnirjz2cjj",
								"version": "0.0.1",
								"name":    "Dev Experience Test",
								"x":       0,
								"y":       0,
							},
							"2": map[string]any{
								"active": false,
							},
						},
					},
				})
			case http.MethodPut:
				var req helix.UpdateUserExtensionsRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Fatalf("Decode() error = %v", err)
				}
				if slot := req.Data.Panel["1"]; !slot.Active || slot.ID != "rh6jq1q334hqc2rr1qlzqbvwlfl3x0" || slot.Version != "1.1.0" {
					t.Fatalf("panel slot = %#v, want active TopClip panel slot", slot)
				}
				if slot := req.Data.Component["1"]; slot.X == nil || *slot.X != 0 || slot.Y == nil || *slot.Y != 0 {
					t.Fatalf("component slot = %#v, want x=0 y=0", slot)
				}
				if slot := req.Data.Component["2"]; slot.Active {
					t.Fatalf("component slot 2 = %#v, want inactive slot", slot)
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": map[string]any{
						"panel": map[string]any{
							"1": map[string]any{
								"active":  true,
								"id":      "rh6jq1q334hqc2rr1qlzqbvwlfl3x0",
								"version": "1.1.0",
								"name":    "TopClip",
							},
						},
						"overlay": map[string]any{
							"1": map[string]any{
								"active":  true,
								"id":      "zfh2irvx2jb4s60f02jq0ajm8vwgka",
								"version": "1.0.19",
								"name":    "Streamlabs",
							},
						},
						"component": map[string]any{
							"1": map[string]any{
								"active":  true,
								"id":      "lqnf3zxk0rv0g7gq92mtmnirjz2cjj",
								"version": "0.0.1",
								"name":    "Dev Experience Test",
								"x":       0,
								"y":       0,
							},
							"2": map[string]any{
								"active": false,
							},
						},
					},
				})
			default:
				t.Fatalf("unexpected method for /users/extensions: %s", r.Method)
			}
		case "/users":
			if got := r.Method; got != http.MethodPut {
				t.Fatalf("method = %q, want PUT", got)
			}
			if got := r.URL.Query().Get("description"); got != "BaldAngel" {
				t.Fatalf("description = %q, want BaldAngel", got)
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{
					"id":                "44322889",
					"login":             "dallas",
					"display_name":      "dallas",
					"type":              "staff",
					"broadcaster_type":  "affiliate",
					"description":       "BaldAngel",
					"profile_image_url": "https://example.com/profile.png",
					"offline_image_url": "https://example.com/offline.png",
					"view_count":        6995,
					"email":             "not-real@email.com",
					"created_at":        "2013-06-03T19:12:02.580593Z",
				}},
			})
		case "/users/blocks":
			switch r.Method {
			case http.MethodGet:
				if got := r.URL.Query().Get("user_id"); got != "141981764" {
					t.Fatalf("user_id = %q, want 141981764", got)
				}
				if got := r.URL.Query().Get("first"); got != "20" {
					t.Fatalf("first = %q, want 20", got)
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": []map[string]any{
						{"user_id": "135093069", "user_login": "bluelava", "display_name": "BlueLava"},
						{"user_id": "27419011", "user_login": "travistyoj", "display_name": "TravistyOJ"},
					},
					"pagination": map[string]any{"cursor": "next-blocks"},
				})
			case http.MethodPut:
				if got := r.URL.Query().Get("target_user_id"); got != "198704263" {
					t.Fatalf("target_user_id = %q, want 198704263", got)
				}
				if got := r.URL.Query().Get("source_context"); got != "chat" {
					t.Fatalf("source_context = %q, want chat", got)
				}
				if got := r.URL.Query().Get("reason"); got != "harassment" {
					t.Fatalf("reason = %q, want harassment", got)
				}
				w.WriteHeader(http.StatusNoContent)
			case http.MethodDelete:
				if got := r.URL.Query().Get("target_user_id"); got != "198704263" {
					t.Fatalf("target_user_id = %q, want 198704263", got)
				}
				w.WriteHeader(http.StatusNoContent)
			default:
				t.Fatalf("unexpected /users/blocks method %q", r.Method)
			}
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
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

	description := "BaldAngel"
	updateResp, _, err := client.Users.Update(context.Background(), helix.UpdateUserParams{
		Description: &description,
	})
	if err != nil {
		t.Fatalf("Users.Update() error = %v", err)
	}
	if got := updateResp.Data[0].Description; got != "BaldAngel" {
		t.Fatalf("Description = %q, want BaldAngel", got)
	}

	authzResp, _, err := client.Users.GetAuthorizationsByUser(context.Background(), helix.GetAuthorizationsByUserParams{
		UserIDs: []string{"141981764", "197886470"},
	})
	if err != nil {
		t.Fatalf("Users.GetAuthorizationsByUser() error = %v", err)
	}
	if got := len(authzResp.Data); got != 2 {
		t.Fatalf("len(authzResp.Data) = %d, want 2", got)
	}
	if got := authzResp.Data[0].UserLogin; got != "twitchdev" {
		t.Fatalf("UserLogin = %q, want twitchdev", got)
	}
	if got := len(authzResp.Data[0].Scopes); got != 3 {
		t.Fatalf("len(Scopes) = %d, want 3", got)
	}

	extensionsResp, _, err := client.Users.GetExtensions(context.Background())
	if err != nil {
		t.Fatalf("Users.GetExtensions() error = %v", err)
	}
	if got := len(extensionsResp.Data); got != 2 {
		t.Fatalf("len(extensionsResp.Data) = %d, want 2", got)
	}
	if got := extensionsResp.Data[0].Name; got != "Streamlabs Leaderboard" {
		t.Fatalf("Extension.Name = %q, want Streamlabs Leaderboard", got)
	}
	if got := len(extensionsResp.Data[1].Type); got != 4 {
		t.Fatalf("len(Extension.Type) = %d, want 4", got)
	}

	activeExtensionsResp, _, err := client.Users.GetActiveExtensions(context.Background(), helix.GetUserActiveExtensionsParams{
		UserID: "141981764",
	})
	if err != nil {
		t.Fatalf("Users.GetActiveExtensions() error = %v", err)
	}
	if got := activeExtensionsResp.Data.Panel["1"].Name; got != "TopClip" {
		t.Fatalf("Panel[1].Name = %q, want TopClip", got)
	}
	if got := activeExtensionsResp.Data.Overlay["1"].Version; got != "1.0.19" {
		t.Fatalf("Overlay[1].Version = %q, want 1.0.19", got)
	}
	if activeExtensionsResp.Data.Component["1"].X == nil || *activeExtensionsResp.Data.Component["1"].X != 0 {
		t.Fatalf("Component[1].X = %#v, want 0", activeExtensionsResp.Data.Component["1"].X)
	}
	if activeExtensionsResp.Data.Component["2"].Active {
		t.Fatal("Component[2].Active = true, want false")
	}

	updatedExtensionsResp, _, err := client.Users.UpdateExtensions(context.Background(), helix.UpdateUserExtensionsRequest{
		Data: helix.UserActiveExtensions{
			Panel: map[string]helix.UserActiveExtensionSlot{
				"1": {
					Active:  true,
					ID:      "rh6jq1q334hqc2rr1qlzqbvwlfl3x0",
					Version: "1.1.0",
				},
			},
			Overlay: map[string]helix.UserActiveExtensionSlot{
				"1": {
					Active:  true,
					ID:      "zfh2irvx2jb4s60f02jq0ajm8vwgka",
					Version: "1.0.19",
				},
			},
			Component: map[string]helix.UserActiveExtensionSlot{
				"1": {
					Active:  true,
					ID:      "lqnf3zxk0rv0g7gq92mtmnirjz2cjj",
					Version: "0.0.1",
					X:       intPtr(0),
					Y:       intPtr(0),
				},
				"2": {
					Active: false,
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("Users.UpdateExtensions() error = %v", err)
	}
	if got := updatedExtensionsResp.Data.Panel["1"].Name; got != "TopClip" {
		t.Fatalf("Updated Panel[1].Name = %q, want TopClip", got)
	}
	if updatedExtensionsResp.Data.Component["2"].Active {
		t.Fatal("Updated Component[2].Active = true, want false")
	}

	blocksResp, blocksMeta, err := client.Users.GetBlocks(context.Background(), helix.GetUserBlockListParams{
		CursorParams: helix.CursorParams{First: 20},
		UserID:       "141981764",
	})
	if err != nil {
		t.Fatalf("Users.GetBlocks() error = %v", err)
	}
	if got := len(blocksResp.Data); got != 2 {
		t.Fatalf("len(blocksResp.Data) = %d, want 2", got)
	}
	if got := blocksResp.Data[0].DisplayName; got != "BlueLava" {
		t.Fatalf("DisplayName = %q, want BlueLava", got)
	}
	if got := blocksMeta.Pagination.Cursor; got != "next-blocks" {
		t.Fatalf("Pagination.Cursor = %q, want next-blocks", got)
	}

	blockMeta, err := client.Users.Block(context.Background(), helix.BlockUserParams{
		TargetUserID:  "198704263",
		SourceContext: "chat",
		Reason:        "harassment",
	})
	if err != nil {
		t.Fatalf("Users.Block() error = %v", err)
	}
	if got := blockMeta.StatusCode; got != http.StatusNoContent {
		t.Fatalf("block StatusCode = %d, want %d", got, http.StatusNoContent)
	}

	unblockMeta, err := client.Users.Unblock(context.Background(), helix.UnblockUserParams{
		TargetUserID: "198704263",
	})
	if err != nil {
		t.Fatalf("Users.Unblock() error = %v", err)
	}
	if got := unblockMeta.StatusCode; got != http.StatusNoContent {
		t.Fatalf("unblock StatusCode = %d, want %d", got, http.StatusNoContent)
	}
}
