package helix_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dlukt/helix"
)

func TestStreamsFollowedAndKeyEndpointsEncodeRequestsAndDecodeResponses(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/streams/followed":
			if got := r.Method; got != http.MethodGet {
				t.Fatalf("method = %q, want GET", got)
			}
			if got := r.URL.Query().Get("user_id"); got != "141981764" {
				t.Fatalf("user_id = %q, want 141981764", got)
			}
			if got := r.URL.Query().Get("first"); got != "20" {
				t.Fatalf("first = %q, want 20", got)
			}
			if got := r.URL.Query().Get("after"); got != "cursor-1" {
				t.Fatalf("after = %q, want cursor-1", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{
					"id":            "42170724654",
					"user_id":       "132954738",
					"user_login":    "aws",
					"user_name":     "AWS",
					"game_id":       "417752",
					"game_name":     "Talk Shows & Podcasts",
					"type":          "live",
					"title":         "AWS Howdy Partner!",
					"viewer_count":  20,
					"started_at":    "2021-03-31T20:57:26Z",
					"language":      "en",
					"thumbnail_url": "https://static-cdn.jtvnw.net/previews-ttv/live_user_aws-{width}x{height}.jpg",
					"tags":          []string{"English"},
					"is_mature":     false,
				}},
				"pagination": map[string]any{
					"cursor": "next-followed-streams",
				},
			})
		case "/streams/key":
			if got := r.Method; got != http.MethodGet {
				t.Fatalf("method = %q, want GET", got)
			}
			if got := r.URL.Query().Get("broadcaster_id"); got != "141981764" {
				t.Fatalf("broadcaster_id = %q, want 141981764", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{
					"stream_key": "live_44322889_a34ub37c8ajv98a0",
				}},
			})
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

	followedResp, followedMeta, err := client.Streams.GetFollowed(context.Background(), helix.GetFollowedStreamsParams{
		CursorParams: helix.CursorParams{
			First: 20,
			After: "cursor-1",
		},
		UserID: "141981764",
	})
	if err != nil {
		t.Fatalf("Streams.GetFollowed() error = %v", err)
	}
	if got := followedResp.Data[0].UserLogin; got != "aws" {
		t.Fatalf("UserLogin = %q, want aws", got)
	}
	if got := followedResp.Data[0].ViewerCount; got != 20 {
		t.Fatalf("ViewerCount = %d, want 20", got)
	}
	if got := followedMeta.Pagination.Cursor; got != "next-followed-streams" {
		t.Fatalf("Pagination.Cursor = %q, want next-followed-streams", got)
	}

	keyResp, _, err := client.Streams.GetKey(context.Background(), helix.GetStreamKeyParams{
		BroadcasterID: "141981764",
	})
	if err != nil {
		t.Fatalf("Streams.GetKey() error = %v", err)
	}
	if got := keyResp.Data[0].StreamKey; got != "live_44322889_a34ub37c8ajv98a0" {
		t.Fatalf("StreamKey = %q, want live_44322889_a34ub37c8ajv98a0", got)
	}
}
