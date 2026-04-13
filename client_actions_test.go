package helix_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dlukt/helix"
)

func TestActionServicesEncodeRequestsAndHandleZeroBodyResponses(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		switch r.URL.Path {
		case "/moderation/bans":
			switch r.Method {
			case http.MethodPost:
				if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
					t.Fatalf("broadcaster_id = %q, want 123", got)
				}
				if got := r.URL.Query().Get("moderator_id"); got != "456" {
					t.Fatalf("moderator_id = %q, want 456", got)
				}
				var req helix.BanUserRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Fatalf("Decode() error = %v", err)
				}
				if got := req.Data.UserID; got != "777" {
					t.Fatalf("UserID = %q, want 777", got)
				}
				if got := req.Data.Duration; got != 600 {
					t.Fatalf("Duration = %d, want 600", got)
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": []map[string]any{{
						"broadcaster_id": "123",
						"moderator_id":   "456",
						"user_id":        "777",
						"created_at":     "2024-04-15T01:30:28Z",
						"end_time":       "2024-04-15T01:40:28Z",
					}},
				})
			case http.MethodDelete:
				if got := r.URL.Query().Get("user_id"); got != "777" {
					t.Fatalf("user_id = %q, want 777", got)
				}
				w.WriteHeader(http.StatusNoContent)
			default:
				t.Fatalf("unexpected method %s", r.Method)
			}
		case "/moderation/chat":
			if got := r.Method; got != http.MethodDelete {
				t.Fatalf("method = %q, want DELETE", got)
			}
			if got := r.URL.Query().Get("message_id"); got != "msg-1" {
				t.Fatalf("message_id = %q, want msg-1", got)
			}
			w.WriteHeader(http.StatusNoContent)
		case "/raids":
			switch r.Method {
			case http.MethodPost:
				if got := r.URL.Query().Get("from_broadcaster_id"); got != "123" {
					t.Fatalf("from_broadcaster_id = %q, want 123", got)
				}
				if got := r.URL.Query().Get("to_broadcaster_id"); got != "999" {
					t.Fatalf("to_broadcaster_id = %q, want 999", got)
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": []map[string]any{{
						"created_at": "2024-04-15T01:30:28Z",
						"is_mature":  false,
					}},
				})
			case http.MethodDelete:
				if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
					t.Fatalf("broadcaster_id = %q, want 123", got)
				}
				w.WriteHeader(http.StatusNoContent)
			default:
				t.Fatalf("unexpected method %s", r.Method)
			}
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

	banResp, _, err := client.Moderation.BanUser(context.Background(), helix.BanUserParams{
		BroadcasterID: "123",
		ModeratorID:   "456",
	}, helix.BanUserRequest{
		Data: helix.BanUserData{
			UserID:   "777",
			Duration: 600,
			Reason:   "spam",
		},
	})
	if err != nil {
		t.Fatalf("BanUser() error = %v", err)
	}
	if got := banResp.Data[0].UserID; got != "777" {
		t.Fatalf("UserID = %q, want 777", got)
	}
	if banResp.Data[0].EndTime == nil {
		t.Fatal("EndTime = nil, want parsed timestamp")
	}

	unbanMeta, err := client.Moderation.UnbanUser(context.Background(), helix.UnbanUserParams{
		BroadcasterID: "123",
		ModeratorID:   "456",
		UserID:        "777",
	})
	if err != nil {
		t.Fatalf("UnbanUser() error = %v", err)
	}
	if got := unbanMeta.StatusCode; got != http.StatusNoContent {
		t.Fatalf("UnbanUser status = %d, want 204", got)
	}

	deleteMeta, err := client.Chat.DeleteChatMessages(context.Background(), helix.DeleteChatMessagesParams{
		BroadcasterID: "123",
		ModeratorID:   "456",
		MessageID:     "msg-1",
	})
	if err != nil {
		t.Fatalf("DeleteChatMessages() error = %v", err)
	}
	if got := deleteMeta.StatusCode; got != http.StatusNoContent {
		t.Fatalf("DeleteChatMessages status = %d, want 204", got)
	}

	raidResp, _, err := client.Raids.Start(context.Background(), helix.StartRaidParams{
		FromBroadcasterID: "123",
		ToBroadcasterID:   "999",
	})
	if err != nil {
		t.Fatalf("Raids.Start() error = %v", err)
	}
	if len(raidResp.Data) != 1 || raidResp.Data[0].IsMature {
		t.Fatalf("raid response = %+v, want one non-mature raid", raidResp.Data)
	}

	cancelMeta, err := client.Raids.Cancel(context.Background(), helix.CancelRaidParams{
		BroadcasterID: "123",
	})
	if err != nil {
		t.Fatalf("Raids.Cancel() error = %v", err)
	}
	if got := cancelMeta.StatusCode; got != http.StatusNoContent {
		t.Fatalf("Raids.Cancel status = %d, want 204", got)
	}
}
