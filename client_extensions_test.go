package helix_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dlukt/helix"
)

func stringPtr(v string) *string {
	return &v
}

func TestExtensionsServiceEncodesRequestsAndDecodesResponses(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/extensions/transactions":
			if got := r.Method; got != http.MethodGet {
				t.Fatalf("method = %q, want GET", got)
			}
			if got := r.URL.Query().Get("extension_id"); got != "ext-1" {
				t.Fatalf("extension_id = %q, want ext-1", got)
			}
			if got := r.URL.Query()["id"]; len(got) != 2 || got[0] != "txn-1" || got[1] != "txn-2" {
				t.Fatalf("id = %v, want [txn-1 txn-2]", got)
			}
			if got := r.URL.Query().Get("first"); got != "10" {
				t.Fatalf("first = %q, want 10", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{
					"id":                "txn-1",
					"timestamp":         "2024-04-15T10:00:00Z",
					"broadcaster_id":    "123",
					"broadcaster_login": "caster",
					"broadcaster_name":  "Caster",
					"user_id":           "456",
					"user_login":        "viewer1",
					"user_name":         "Viewer1",
					"product_type":      "BITS_IN_EXTENSION",
					"product_data": map[string]any{
						"domain":        "twitch.ext.ext-1",
						"sku":           "sku-1",
						"inDevelopment": false,
						"displayName":   "Highlight My Message",
						"broadcast":     true,
						"expiration":    "",
						"cost": map[string]any{
							"amount": 100,
							"type":   "bits",
						},
					},
				}},
				"pagination": map[string]any{
					"cursor": "next-transactions",
				},
			})
		case "/extensions":
			if got := r.Method; got != http.MethodGet {
				t.Fatalf("method = %q, want GET", got)
			}
			if got := r.URL.Query().Get("extension_id"); got != "ext-1" {
				t.Fatalf("extension_id = %q, want ext-1", got)
			}
			if got := r.URL.Query().Get("extension_version"); got != "1.2.3" {
				t.Fatalf("extension_version = %q, want 1.2.3", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{
					"author_name":                 "Dev",
					"bits_enabled":                true,
					"can_install":                 true,
					"configuration_location":      "hosted",
					"description":                 "Useful extension",
					"eula_tos_url":                "https://example.com/eula",
					"has_chat_support":            false,
					"icon_url":                    "https://example.com/icon.png",
					"icon_urls":                   map[string]any{"100x100": "https://example.com/icon-100.png"},
					"id":                          "ext-1",
					"name":                        "My Extension",
					"privacy_policy_url":          "https://example.com/privacy",
					"request_identity_link":       false,
					"screenshot_urls":             []string{"https://example.com/screen.png"},
					"state":                       "Released",
					"subscriptions_support_level": "none",
					"summary":                     "summary",
					"support_email":               "dev@example.com",
					"version":                     "1.2.3",
					"viewer_summary":              "viewer summary",
					"views": map[string]any{
						"panel": map[string]any{
							"viewer_url":                "https://example.com/panel",
							"height":                    300,
							"can_link_external_content": false,
						},
						"component": map[string]any{
							"viewer_url":                "https://example.com/component",
							"aspect_width":              1280,
							"aspect_height":             720,
							"aspect_ratio_x":            16,
							"aspect_ratio_y":            9,
							"autoscale":                 true,
							"scale_pixels":              1024,
							"target_height":             5333,
							"size":                      1,
							"zoom":                      true,
							"zoom_pixels":               800,
							"can_link_external_content": false,
						},
					},
					"allowlisted_config_urls": []string{"https://example.com/config"},
					"allowlisted_panel_urls":  []string{"https://example.com/panel"},
				}},
			})
		case "/extensions/configurations":
			switch r.Method {
			case http.MethodGet:
				if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
					t.Fatalf("broadcaster_id = %q, want 123", got)
				}
				if got := r.URL.Query().Get("extension_id"); got != "ext-1" {
					t.Fatalf("extension_id = %q, want ext-1", got)
				}
				segments := r.URL.Query()["segment"]
				switch {
				case len(segments) == 1 && segments[0] == "broadcaster":
					_ = json.NewEncoder(w).Encode(map[string]any{
						"data": []map[string]any{{
							"content":        `{"theme":"blue"}`,
							"version":        "v1",
							"broadcaster_id": "123",
							"extension_id":   "ext-1",
							"segment":        "broadcaster",
						}},
					})
				case len(segments) == 2 && segments[0] == "broadcaster" && segments[1] == "developer":
					_ = json.NewEncoder(w).Encode(map[string]any{
						"data": []map[string]any{
							{
								"content":        `{"theme":"blue"}`,
								"version":        "v1",
								"broadcaster_id": "123",
								"extension_id":   "ext-1",
								"segment":        "broadcaster",
							},
							{
								"content":        `{"theme":"red"}`,
								"version":        "v2",
								"broadcaster_id": "123",
								"extension_id":   "ext-1",
								"segment":        "developer",
							},
						},
					})
				default:
					t.Fatalf("segment = %v, want [broadcaster] or [broadcaster developer]", segments)
				}
			case http.MethodPut:
				var req helix.SetExtensionConfigurationSegmentRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Fatalf("Decode() error = %v", err)
				}
				if got := req.ExtensionID; got != "ext-1" {
					t.Fatalf("extension_id = %q, want ext-1", got)
				}
				if got := req.Segment; got != "broadcaster" {
					t.Fatalf("segment = %q, want broadcaster", got)
				}
				if got := req.BroadcasterID; got != "123" {
					t.Fatalf("broadcaster_id = %q, want 123", got)
				}
				switch req.Version {
				case "v1":
					if req.Content == nil || *req.Content != `{"theme":"blue"}` {
						t.Fatalf("content = %#v, want json payload", req.Content)
					}
				case "v2":
					if req.Content != nil {
						t.Fatalf("content = %#v, want nil", req.Content)
					}
				default:
					t.Fatalf("version = %q, want v1 or v2", req.Version)
				}
				w.WriteHeader(http.StatusNoContent)
			default:
				t.Fatalf("unexpected method for /extensions/configurations: %s", r.Method)
			}
		case "/extensions/required_configuration":
			if got := r.Method; got != http.MethodPut {
				t.Fatalf("method = %q, want PUT", got)
			}
			if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
				t.Fatalf("broadcaster_id = %q, want 123", got)
			}
			var req helix.SetExtensionRequiredConfigurationRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("Decode() error = %v", err)
			}
			if got := req.ExtensionVersion; got != "1.2.3" {
				t.Fatalf("extension_version = %q, want 1.2.3", got)
			}
			if got := req.RequiredConfiguration; got != "required-config" {
				t.Fatalf("required_configuration = %q, want required-config", got)
			}
			w.WriteHeader(http.StatusNoContent)
		case "/extensions/pubsub":
			if got := r.Method; got != http.MethodPost {
				t.Fatalf("method = %q, want POST", got)
			}
			var req helix.SendExtensionPubSubMessageRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("Decode() error = %v", err)
			}
			if len(req.Target) != 1 || req.Target[0] != "broadcast" {
				t.Fatalf("target = %#v, want [broadcast]", req.Target)
			}
			if got := req.BroadcasterID; got != "123" {
				t.Fatalf("broadcaster_id = %q, want 123", got)
			}
			if req.IsGlobalBroadcast == nil || *req.IsGlobalBroadcast {
				t.Fatalf("is_global_broadcast = %#v, want false", req.IsGlobalBroadcast)
			}
			if got := req.Message; got != `{"event":"refresh"}` {
				t.Fatalf("message = %q, want refresh json", got)
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	defer server.Close()

	client, err := helix.New(helix.Config{
		ClientID: "client-id",
		BaseURL:  server.URL,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	transactionsResp, transactionsMeta, err := client.Extensions.GetTransactions(context.Background(), helix.GetExtensionTransactionsParams{
		CursorParams:   helix.CursorParams{First: 10},
		ExtensionID:    "ext-1",
		TransactionIDs: []string{"txn-1", "txn-2"},
	})
	if err != nil {
		t.Fatalf("Extensions.GetTransactions() error = %v", err)
	}
	if got := transactionsResp.Data[0].ProductData.DisplayName; got != "Highlight My Message" {
		t.Fatalf("ProductData.DisplayName = %q, want Highlight My Message", got)
	}
	if got := transactionsResp.Data[0].Timestamp.UTC(); !got.Equal(time.Date(2024, 4, 15, 10, 0, 0, 0, time.UTC)) {
		t.Fatalf("Timestamp = %v, want 2024-04-15T10:00:00Z", got)
	}
	if got := transactionsMeta.Pagination.Cursor; got != "next-transactions" {
		t.Fatalf("Transactions cursor = %q, want next-transactions", got)
	}

	extensionsResp, _, err := client.Extensions.Get(context.Background(), helix.GetExtensionParams{
		ExtensionID:      "ext-1",
		ExtensionVersion: "1.2.3",
	})
	if err != nil {
		t.Fatalf("Extensions.Get() error = %v", err)
	}
	if got := extensionsResp.Data[0].Name; got != "My Extension" {
		t.Fatalf("Name = %q, want My Extension", got)
	}
	if got := extensionsResp.Data[0].Views.Panel.ViewerURL; got != "https://example.com/panel" {
		t.Fatalf("Panel viewer_url = %q, want https://example.com/panel", got)
	}
	component := extensionsResp.Data[0].Views.Component
	if component == nil {
		t.Fatal("Component view = nil, want populated component config")
	}
	if got := component.ViewerURL; got != "https://example.com/component" {
		t.Fatalf("Component viewer_url = %q, want https://example.com/component", got)
	}
	if got := component.AspectWidth; got != 1280 {
		t.Fatalf("Component aspect_width = %d, want 1280", got)
	}
	if got := component.AspectHeight; got != 720 {
		t.Fatalf("Component aspect_height = %d, want 720", got)
	}
	if got := component.Size; got != 1 {
		t.Fatalf("Component size = %d, want 1", got)
	}
	if !component.Zoom {
		t.Fatal("Component zoom = false, want true")
	}
	if got := component.ZoomPixels; got != 800 {
		t.Fatalf("Component zoom_pixels = %d, want 800", got)
	}

	configResp, _, err := client.Extensions.GetConfigurationSegment(context.Background(), helix.GetExtensionConfigurationSegmentParams{
		BroadcasterID: "123",
		ExtensionID:   "ext-1",
		Segment:       "broadcaster",
	})
	if err != nil {
		t.Fatalf("Extensions.GetConfigurationSegment() error = %v", err)
	}
	if got := configResp.Data[0].Content; got != `{"theme":"blue"}` {
		t.Fatalf("Configuration content = %q, want json payload", got)
	}

	multiConfigResp, _, err := client.Extensions.GetConfigurationSegment(context.Background(), helix.GetExtensionConfigurationSegmentParams{
		BroadcasterID: "123",
		ExtensionID:   "ext-1",
		Segments:      []string{"broadcaster", "developer"},
	})
	if err != nil {
		t.Fatalf("Extensions.GetConfigurationSegment() multiple segments error = %v", err)
	}
	if got := len(multiConfigResp.Data); got != 2 {
		t.Fatalf("Multi configuration len = %d, want 2", got)
	}
	if got := multiConfigResp.Data[1].Segment; got != "developer" {
		t.Fatalf("Second configuration segment = %q, want developer", got)
	}

	setConfigMeta, err := client.Extensions.SetConfigurationSegment(context.Background(), helix.SetExtensionConfigurationSegmentRequest{
		ExtensionID:   "ext-1",
		Segment:       "broadcaster",
		BroadcasterID: "123",
		Content:       stringPtr(`{"theme":"blue"}`),
		Version:       "v1",
	})
	if err != nil {
		t.Fatalf("Extensions.SetConfigurationSegment() error = %v", err)
	}
	if got := setConfigMeta.StatusCode; got != http.StatusNoContent {
		t.Fatalf("SetConfigurationSegment status = %d, want %d", got, http.StatusNoContent)
	}

	partialConfigMeta, err := client.Extensions.SetConfigurationSegment(context.Background(), helix.SetExtensionConfigurationSegmentRequest{
		ExtensionID:   "ext-1",
		Segment:       "broadcaster",
		BroadcasterID: "123",
		Version:       "v2",
	})
	if err != nil {
		t.Fatalf("Extensions.SetConfigurationSegment() partial update error = %v", err)
	}
	if got := partialConfigMeta.StatusCode; got != http.StatusNoContent {
		t.Fatalf("SetConfigurationSegment partial update status = %d, want %d", got, http.StatusNoContent)
	}

	requiredMeta, err := client.Extensions.SetRequiredConfiguration(context.Background(), helix.SetExtensionRequiredConfigurationRequest{
		ExtensionID:           "ext-1",
		ExtensionVersion:      "1.2.3",
		RequiredConfiguration: "required-config",
		BroadcasterID:         "123",
	})
	if err != nil {
		t.Fatalf("Extensions.SetRequiredConfiguration() error = %v", err)
	}
	if got := requiredMeta.StatusCode; got != http.StatusNoContent {
		t.Fatalf("SetRequiredConfiguration status = %d, want %d", got, http.StatusNoContent)
	}

	pubsubMeta, err := client.Extensions.SendPubSubMessage(context.Background(), helix.SendExtensionPubSubMessageRequest{
		Target:            []string{"broadcast"},
		BroadcasterID:     "123",
		IsGlobalBroadcast: boolPtr(false),
		Message:           `{"event":"refresh"}`,
	})
	if err != nil {
		t.Fatalf("Extensions.SendPubSubMessage() error = %v", err)
	}
	if got := pubsubMeta.StatusCode; got != http.StatusNoContent {
		t.Fatalf("SendPubSubMessage status = %d, want %d", got, http.StatusNoContent)
	}
}
