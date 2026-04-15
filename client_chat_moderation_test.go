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

func intPtr(v int) *int {
	return &v
}

func TestChatAndModerationServicesEncodeRequestsAndDecodeResponses(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/chat/settings":
			switch r.Method {
			case http.MethodGet:
				if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
					t.Fatalf("broadcaster_id = %q, want 123", got)
				}
				if got := r.URL.Query().Get("moderator_id"); got != "456" {
					t.Fatalf("moderator_id = %q, want 456", got)
				}
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": []map[string]any{{
						"broadcaster_id":                    "123",
						"moderator_id":                      "456",
						"emote_mode":                        false,
						"slow_mode":                         true,
						"slow_mode_wait_time":               10,
						"follower_mode":                     true,
						"follower_mode_duration":            5,
						"subscriber_mode":                   false,
						"unique_chat_mode":                  true,
						"non_moderator_chat_delay":          true,
						"non_moderator_chat_delay_duration": 4,
					}},
				})
			case http.MethodPatch:
				if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
					t.Fatalf("patch broadcaster_id = %q, want 123", got)
				}
				if got := r.URL.Query().Get("moderator_id"); got != "456" {
					t.Fatalf("patch moderator_id = %q, want 456", got)
				}
				var req helix.UpdateChatSettingsRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Fatalf("Decode() error = %v", err)
				}
				if req.SlowMode == nil || !*req.SlowMode {
					t.Fatalf("SlowMode = %#v, want true", req.SlowMode)
				}
				if req.SlowModeWaitTime == nil || *req.SlowModeWaitTime != 30 {
					t.Fatalf("SlowModeWaitTime = %#v, want 30", req.SlowModeWaitTime)
				}
				if req.UniqueChatMode == nil || !*req.UniqueChatMode {
					t.Fatalf("UniqueChatMode = %#v, want true", req.UniqueChatMode)
				}
				if req.FollowerMode != nil {
					t.Fatalf("FollowerMode = %#v, want omitted", req.FollowerMode)
				}
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": []map[string]any{{
						"broadcaster_id":                    "123",
						"moderator_id":                      "456",
						"emote_mode":                        false,
						"slow_mode":                         true,
						"slow_mode_wait_time":               30,
						"follower_mode":                     true,
						"follower_mode_duration":            5,
						"subscriber_mode":                   false,
						"unique_chat_mode":                  true,
						"non_moderator_chat_delay":          true,
						"non_moderator_chat_delay_duration": 4,
					}},
				})
			default:
				t.Fatalf("unexpected method for /chat/settings: %s", r.Method)
			}
		case "/chat/chatters":
			if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
				t.Fatalf("broadcaster_id = %q, want 123", got)
			}
			if got := r.URL.Query().Get("moderator_id"); got != "456" {
				t.Fatalf("moderator_id = %q, want 456", got)
			}
			if got := r.URL.Query().Get("first"); got != "25" {
				t.Fatalf("first = %q, want 25", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{
					"user_id":    "u-1",
					"user_login": "viewer1",
					"user_name":  "Viewer1",
				}},
				"total": 1,
				"pagination": map[string]any{
					"cursor": "next-chatters",
				},
			})
		case "/chat/color":
			switch r.Method {
			case http.MethodGet:
				if got := r.URL.Query()["user_id"]; len(got) != 2 || got[0] != "456" || got[1] != "789" {
					t.Fatalf("chat color user_id = %v, want [456 789]", got)
				}
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": []map[string]any{
						{
							"user_id":    "456",
							"user_login": "mod1",
							"user_name":  "Mod1",
							"color":      "#9146FF",
						},
						{
							"user_id":    "789",
							"user_login": "viewer2",
							"user_name":  "Viewer2",
							"color":      "",
						},
					},
				})
			case http.MethodPut:
				if got := r.URL.Query().Get("user_id"); got != "456" {
					t.Fatalf("update chat color user_id = %q, want 456", got)
				}
				if got := r.URL.Query().Get("color"); got != "#9146FF" {
					t.Fatalf("update chat color color = %q, want #9146FF", got)
				}
				w.WriteHeader(http.StatusNoContent)
			default:
				t.Fatalf("unexpected method for /chat/color: %s", r.Method)
			}
		case "/chat/announcements":
			if r.Method != http.MethodPost {
				t.Fatalf("announcement method = %s, want POST", r.Method)
			}
			if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
				t.Fatalf("announcement broadcaster_id = %q, want 123", got)
			}
			if got := r.URL.Query().Get("moderator_id"); got != "456" {
				t.Fatalf("announcement moderator_id = %q, want 456", got)
			}
			var req helix.SendAnnouncementRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("Decode() error = %v", err)
			}
			if got := req.Message; got != "Heads up!" {
				t.Fatalf("announcement message = %q, want Heads up!", got)
			}
			if got := req.Color; got != "purple" {
				t.Fatalf("announcement color = %q, want purple", got)
			}
			if req.ForSourceOnly == nil || *req.ForSourceOnly {
				t.Fatalf("announcement ForSourceOnly = %#v, want false", req.ForSourceOnly)
			}
			w.WriteHeader(http.StatusNoContent)
		case "/chat/shoutouts":
			if r.Method != http.MethodPost {
				t.Fatalf("shoutout method = %s, want POST", r.Method)
			}
			if got := r.URL.Query().Get("from_broadcaster_id"); got != "123" {
				t.Fatalf("from_broadcaster_id = %q, want 123", got)
			}
			if got := r.URL.Query().Get("to_broadcaster_id"); got != "789" {
				t.Fatalf("to_broadcaster_id = %q, want 789", got)
			}
			if got := r.URL.Query().Get("moderator_id"); got != "456" {
				t.Fatalf("shoutout moderator_id = %q, want 456", got)
			}
			w.WriteHeader(http.StatusNoContent)
		case "/chat/messages":
			if r.Method != http.MethodPost {
				t.Fatalf("message method = %s, want POST", r.Method)
			}
			var req helix.SendMessageRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("Decode() error = %v", err)
			}
			if got := req.BroadcasterID; got != "123" {
				t.Fatalf("message broadcaster_id = %q, want 123", got)
			}
			if got := req.SenderID; got != "456" {
				t.Fatalf("message sender_id = %q, want 456", got)
			}
			if got := req.Message; got != "Hello, world!" {
				t.Fatalf("message text = %q, want Hello, world!", got)
			}
			if got := req.ReplyParentMessageID; got != "parent-1" {
				t.Fatalf("message reply_parent_message_id = %q, want parent-1", got)
			}
			if req.ForSourceOnly == nil || *req.ForSourceOnly {
				t.Fatalf("message ForSourceOnly = %#v, want false", req.ForSourceOnly)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{
					"message_id":  "msg-1",
					"is_sent":     true,
					"drop_reason": nil,
				}},
			})
		case "/moderation/moderators":
			switch r.Method {
			case http.MethodGet:
				if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
					t.Fatalf("broadcaster_id = %q, want 123", got)
				}
				if got := r.URL.Query()["user_id"]; len(got) != 2 || got[0] != "456" || got[1] != "789" {
					t.Fatalf("user_id = %v, want [456 789]", got)
				}
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": []map[string]any{{
						"user_id":    "456",
						"user_login": "mod1",
						"user_name":  "Mod1",
					}},
					"pagination": map[string]any{
						"cursor": "next-moderators",
					},
				})
			case http.MethodPost:
				if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
					t.Fatalf("add moderator broadcaster_id = %q, want 123", got)
				}
				if got := r.URL.Query().Get("user_id"); got != "789" {
					t.Fatalf("add moderator user_id = %q, want 789", got)
				}
				w.WriteHeader(http.StatusNoContent)
			case http.MethodDelete:
				if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
					t.Fatalf("remove moderator broadcaster_id = %q, want 123", got)
				}
				if got := r.URL.Query().Get("user_id"); got != "789" {
					t.Fatalf("remove moderator user_id = %q, want 789", got)
				}
				w.WriteHeader(http.StatusNoContent)
			default:
				t.Fatalf("unexpected method for /moderation/moderators: %s", r.Method)
			}
		case "/moderation/channels":
			if got := r.Method; got != http.MethodGet {
				t.Fatalf("method = %q, want GET", got)
			}
			if got := r.URL.Query().Get("user_id"); got != "456" {
				t.Fatalf("user_id = %q, want 456", got)
			}
			if got := r.URL.Query().Get("first"); got != "10" {
				t.Fatalf("first = %q, want 10", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{
					"broadcaster_id":    "123",
					"broadcaster_login": "caster",
					"broadcaster_name":  "Caster",
				}},
				"pagination": map[string]any{
					"cursor": "next-moderated-channels",
				},
			})
		case "/moderation/banned":
			if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
				t.Fatalf("broadcaster_id = %q, want 123", got)
			}
			if got := r.URL.Query()["user_id"]; len(got) != 2 || got[0] != "777" || got[1] != "888" {
				t.Fatalf("user_id = %v, want [777 888]", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{
					"user_id":         "777",
					"user_login":      "badviewer",
					"user_name":       "BadViewer",
					"expires_at":      "2024-04-15T02:00:28Z",
					"created_at":      "2024-04-15T01:30:28Z",
					"reason":          "spam",
					"moderator_id":    "456",
					"moderator_login": "mod1",
					"moderator_name":  "Mod1",
				}},
				"pagination": map[string]any{
					"cursor": "next-banned",
				},
			})
		case "/moderation/blocked_terms":
			switch r.Method {
			case http.MethodGet:
				if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
					t.Fatalf("blocked terms broadcaster_id = %q, want 123", got)
				}
				if got := r.URL.Query().Get("moderator_id"); got != "456" {
					t.Fatalf("blocked terms moderator_id = %q, want 456", got)
				}
				if got := r.URL.Query().Get("first"); got != "10" {
					t.Fatalf("blocked terms first = %q, want 10", got)
				}
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": []map[string]any{{
						"id":         "term-1",
						"text":       "spoiler",
						"created_at": "2024-04-15T01:30:28Z",
						"updated_at": "2024-04-15T01:30:28Z",
						"expires_at": nil,
					}},
					"pagination": map[string]any{
						"cursor": "next-blocked-terms",
					},
				})
			case http.MethodPost:
				if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
					t.Fatalf("add blocked term broadcaster_id = %q, want 123", got)
				}
				if got := r.URL.Query().Get("moderator_id"); got != "456" {
					t.Fatalf("add blocked term moderator_id = %q, want 456", got)
				}
				var req helix.AddBlockedTermRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Fatalf("Decode() error = %v", err)
				}
				if got := req.Text; got != "spoiler" {
					t.Fatalf("blocked term text = %q, want spoiler", got)
				}
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": []map[string]any{{
						"id":         "term-1",
						"text":       "spoiler",
						"created_at": "2024-04-15T01:30:28Z",
						"updated_at": "2024-04-15T01:30:28Z",
						"expires_at": nil,
					}},
				})
			case http.MethodDelete:
				if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
					t.Fatalf("remove blocked term broadcaster_id = %q, want 123", got)
				}
				if got := r.URL.Query().Get("moderator_id"); got != "456" {
					t.Fatalf("remove blocked term moderator_id = %q, want 456", got)
				}
				if got := r.URL.Query().Get("id"); got != "term-1" {
					t.Fatalf("remove blocked term id = %q, want term-1", got)
				}
				w.WriteHeader(http.StatusNoContent)
			default:
				t.Fatalf("unexpected method for /moderation/blocked_terms: %s", r.Method)
			}
		case "/moderation/enforcements/status":
			if got := r.Method; got != http.MethodPost {
				t.Fatalf("method = %q, want POST", got)
			}
			if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
				t.Fatalf("broadcaster_id = %q, want 123", got)
			}
			var req helix.CheckAutoModStatusRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("Decode() error = %v", err)
			}
			if len(req.Data) != 2 {
				t.Fatalf("len(data) = %d, want 2", len(req.Data))
			}
			if got := req.Data[0].MsgID; got != "msg-1" {
				t.Fatalf("data[0].msg_id = %q, want msg-1", got)
			}
			if got := req.Data[0].UserID; got != "456" {
				t.Fatalf("data[0].user_id = %q, want 456", got)
			}
			if got := req.Data[1].MsgText; got != "You stink!" {
				t.Fatalf("data[1].msg_text = %q, want You stink!", got)
			}
			if got := req.Data[1].UserID; got != "789" {
				t.Fatalf("data[1].user_id = %q, want 789", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{
					{
						"msg_id":       "msg-1",
						"is_permitted": true,
					},
					{
						"msg_id":       "msg-2",
						"is_permitted": false,
					},
				},
			})
		case "/moderation/automod/message":
			if got := r.Method; got != http.MethodPost {
				t.Fatalf("method = %q, want POST", got)
			}
			var req helix.ManageHeldAutoModMessageRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("Decode() error = %v", err)
			}
			if got := req.UserID; got != "456" {
				t.Fatalf("user_id = %q, want 456", got)
			}
			if got := req.MsgID; got != "held-msg-1" {
				t.Fatalf("msg_id = %q, want held-msg-1", got)
			}
			if got := req.Action; got != "ALLOW" {
				t.Fatalf("action = %q, want ALLOW", got)
			}
			w.WriteHeader(http.StatusNoContent)
		case "/moderation/automod/settings":
			switch r.Method {
			case http.MethodGet:
				if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
					t.Fatalf("broadcaster_id = %q, want 123", got)
				}
				if got := r.URL.Query().Get("moderator_id"); got != "456" {
					t.Fatalf("moderator_id = %q, want 456", got)
				}
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": []map[string]any{{
						"broadcaster_id":             "123",
						"moderator_id":               "456",
						"overall_level":              nil,
						"disability":                 1,
						"aggression":                 2,
						"sexuality_sex_or_gender":    1,
						"misogyny":                   1,
						"bullying":                   2,
						"swearing":                   0,
						"race_ethnicity_or_religion": 3,
						"sex_based_terms":            1,
					}},
				})
			case http.MethodPut:
				if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
					t.Fatalf("broadcaster_id = %q, want 123", got)
				}
				if got := r.URL.Query().Get("moderator_id"); got != "456" {
					t.Fatalf("moderator_id = %q, want 456", got)
				}
				var req helix.UpdateAutoModSettingsRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Fatalf("Decode() error = %v", err)
				}
				if req.OverallLevel == nil || *req.OverallLevel != 3 {
					t.Fatalf("overall_level = %#v, want 3", req.OverallLevel)
				}
				if req.Aggression != nil || req.Bullying != nil || req.Disability != nil ||
					req.Misogyny != nil || req.RaceEthnicityOrReligion != nil ||
					req.SexBasedTerms != nil || req.SexualitySexOrGender != nil ||
					req.Swearing != nil {
					t.Fatalf("category fields = %#v, want omitted", req)
				}
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": []map[string]any{{
						"broadcaster_id":             "123",
						"moderator_id":               "456",
						"overall_level":              3,
						"disability":                 3,
						"aggression":                 3,
						"sexuality_sex_or_gender":    3,
						"misogyny":                   3,
						"bullying":                   2,
						"swearing":                   0,
						"race_ethnicity_or_religion": 3,
						"sex_based_terms":            3,
					}},
				})
			default:
				t.Fatalf("unexpected method for /moderation/automod/settings: %s", r.Method)
			}
		case "/moderation/warnings":
			if got := r.Method; got != http.MethodPost {
				t.Fatalf("method = %q, want POST", got)
			}
			if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
				t.Fatalf("broadcaster_id = %q, want 123", got)
			}
			if got := r.URL.Query().Get("moderator_id"); got != "456" {
				t.Fatalf("moderator_id = %q, want 456", got)
			}
			var req helix.WarnUserRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("Decode() error = %v", err)
			}
			if got := req.Data.UserID; got != "9876" {
				t.Fatalf("user_id = %q, want 9876", got)
			}
			if got := req.Data.Reason; got != "stop doing that!" {
				t.Fatalf("reason = %q, want stop doing that!", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{
					"broadcaster_id": "123",
					"user_id":        "9876",
					"moderator_id":   "456",
					"reason":         "stop doing that!",
				}},
			})
		case "/moderation/suspicious_users":
			switch r.Method {
			case http.MethodPost:
				if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
					t.Fatalf("broadcaster_id = %q, want 123", got)
				}
				if got := r.URL.Query().Get("moderator_id"); got != "456" {
					t.Fatalf("moderator_id = %q, want 456", got)
				}
				var req helix.SuspiciousUserStatusChangeRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Fatalf("Decode() error = %v", err)
				}
				if got := req.UserID; got != "9876" {
					t.Fatalf("user_id = %q, want 9876", got)
				}
				if got := req.Status; got != "RESTRICTED" {
					t.Fatalf("status = %q, want RESTRICTED", got)
				}
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": []map[string]any{{
						"user_id":        "9876",
						"broadcaster_id": "123",
						"moderator_id":   "456",
						"updated_at":     "2025-12-01T23:08:18Z",
						"status":         "RESTRICTED",
						"types":          []string{"MANUALLY_ADDED"},
					}},
				})
			case http.MethodDelete:
				if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
					t.Fatalf("broadcaster_id = %q, want 123", got)
				}
				if got := r.URL.Query().Get("moderator_id"); got != "456" {
					t.Fatalf("moderator_id = %q, want 456", got)
				}
				if got := r.URL.Query().Get("user_id"); got != "9876" {
					t.Fatalf("user_id = %q, want 9876", got)
				}
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": []map[string]any{{
						"user_id":        "9876",
						"broadcaster_id": "123",
						"moderator_id":   "456",
						"updated_at":     "2025-12-01T23:09:00Z",
						"status":         "NO_TREATMENT",
						"types":          []string{"MANUALLY_ADDED"},
					}},
				})
			default:
				t.Fatalf("unexpected method for /moderation/suspicious_users: %s", r.Method)
			}
		case "/moderation/unban_requests":
			switch r.Method {
			case http.MethodGet:
				if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
					t.Fatalf("broadcaster_id = %q, want 123", got)
				}
				if got := r.URL.Query().Get("moderator_id"); got != "456" {
					t.Fatalf("moderator_id = %q, want 456", got)
				}
				if got := r.URL.Query().Get("status"); got != "pending" {
					t.Fatalf("status = %q, want pending", got)
				}
				if got := r.URL.Query().Get("user_id"); got != "789" {
					t.Fatalf("user_id = %q, want 789", got)
				}
				if got := r.URL.Query().Get("first"); got != "10" {
					t.Fatalf("first = %q, want 10", got)
				}
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": []map[string]any{{
						"id":                "unban-1",
						"broadcaster_id":    "123",
						"broadcaster_login": "caster",
						"broadcaster_name":  "Caster",
						"moderator_id":      "456",
						"moderator_login":   "mod1",
						"moderator_name":    "Mod1",
						"user_id":           "789",
						"user_login":        "viewer3",
						"user_name":         "Viewer3",
						"text":              "please unban me",
						"status":            "pending",
						"created_at":        "2024-04-15T01:30:28Z",
						"resolved_at":       nil,
						"resolution_text":   nil,
					}},
					"pagination": map[string]any{
						"cursor": "next-unban-requests",
					},
				})
			case http.MethodPatch:
				if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
					t.Fatalf("broadcaster_id = %q, want 123", got)
				}
				if got := r.URL.Query().Get("moderator_id"); got != "456" {
					t.Fatalf("moderator_id = %q, want 456", got)
				}
				if got := r.URL.Query().Get("unban_request_id"); got != "unban-1" {
					t.Fatalf("unban_request_id = %q, want unban-1", got)
				}
				if got := r.URL.Query().Get("status"); got != "approved" {
					t.Fatalf("status = %q, want approved", got)
				}
				if got := r.URL.Query().Get("resolution_text"); got != "welcome back" {
					t.Fatalf("resolution_text = %q, want welcome back", got)
				}
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": []map[string]any{{
						"id":                "unban-1",
						"broadcaster_id":    "123",
						"broadcaster_login": "caster",
						"broadcaster_name":  "Caster",
						"moderator_id":      "456",
						"moderator_login":   "mod1",
						"moderator_name":    "Mod1",
						"user_id":           "789",
						"user_login":        "viewer3",
						"user_name":         "Viewer3",
						"text":              "please unban me",
						"status":            "approved",
						"created_at":        "2024-04-15T01:30:28Z",
						"resolved_at":       "2024-04-15T01:40:28Z",
						"resolution_text":   "welcome back",
					}},
				})
			default:
				t.Fatalf("unexpected method for /moderation/unban_requests: %s", r.Method)
			}
		case "/moderation/shield_mode":
			switch r.Method {
			case http.MethodGet:
				if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
					t.Fatalf("shield broadcaster_id = %q, want 123", got)
				}
				if got := r.URL.Query().Get("moderator_id"); got != "456" {
					t.Fatalf("shield moderator_id = %q, want 456", got)
				}
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": []map[string]any{{
						"is_active":         true,
						"moderator_id":      "456",
						"moderator_login":   "mod1",
						"moderator_name":    "Mod1",
						"last_activated_at": "2024-04-15T01:30:28Z",
					}},
				})
			case http.MethodPut:
				if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
					t.Fatalf("update shield broadcaster_id = %q, want 123", got)
				}
				if got := r.URL.Query().Get("moderator_id"); got != "456" {
					t.Fatalf("update shield moderator_id = %q, want 456", got)
				}
				var req helix.UpdateShieldModeStatusRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Fatalf("Decode() error = %v", err)
				}
				if !req.IsActive {
					t.Fatal("IsActive = false, want true")
				}
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": []map[string]any{{
						"is_active":         true,
						"moderator_id":      "456",
						"moderator_login":   "mod1",
						"moderator_name":    "Mod1",
						"last_activated_at": "2024-04-15T01:31:00Z",
					}},
				})
			default:
				t.Fatalf("unexpected method for /moderation/shield_mode: %s", r.Method)
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

	settingsResp, _, err := client.Chat.GetSettings(context.Background(), helix.GetChatSettingsParams{
		BroadcasterID: "123",
		ModeratorID:   "456",
	})
	if err != nil {
		t.Fatalf("Chat.GetSettings() error = %v", err)
	}
	if got := settingsResp.Data[0].SlowMode; !got {
		t.Fatal("SlowMode = false, want true")
	}
	if settingsResp.Data[0].SlowModeWaitTime == nil || *settingsResp.Data[0].SlowModeWaitTime != 10 {
		t.Fatalf("SlowModeWaitTime = %v, want 10", settingsResp.Data[0].SlowModeWaitTime)
	}
	if settingsResp.Data[0].NonModeratorChatDelay == nil || !*settingsResp.Data[0].NonModeratorChatDelay {
		t.Fatalf("NonModeratorChatDelay = %v, want true", settingsResp.Data[0].NonModeratorChatDelay)
	}
	if settingsResp.Data[0].NonModeratorChatDelayDuration == nil || *settingsResp.Data[0].NonModeratorChatDelayDuration != 4 {
		t.Fatalf("NonModeratorChatDelayDuration = %v, want 4", settingsResp.Data[0].NonModeratorChatDelayDuration)
	}

	updatedResp, _, err := client.Chat.UpdateSettings(context.Background(), helix.UpdateChatSettingsParams{
		BroadcasterID: "123",
		ModeratorID:   "456",
	}, helix.UpdateChatSettingsRequest{
		SlowMode:         boolPtr(true),
		SlowModeWaitTime: intPtr(30),
		UniqueChatMode:   boolPtr(true),
	})
	if err != nil {
		t.Fatalf("Chat.UpdateSettings() error = %v", err)
	}
	if updatedResp.Data[0].SlowModeWaitTime == nil || *updatedResp.Data[0].SlowModeWaitTime != 30 {
		t.Fatalf("updated SlowModeWaitTime = %v, want 30", updatedResp.Data[0].SlowModeWaitTime)
	}

	chattersResp, chattersMeta, err := client.Chat.GetChatters(context.Background(), helix.GetChattersParams{
		CursorParams:  helix.CursorParams{First: 25},
		BroadcasterID: "123",
		ModeratorID:   "456",
	})
	if err != nil {
		t.Fatalf("Chat.GetChatters() error = %v", err)
	}
	if got := chattersResp.Total; got != 1 {
		t.Fatalf("Total = %d, want 1", got)
	}
	if got := chattersMeta.Pagination.Cursor; got != "next-chatters" {
		t.Fatalf("Chatters cursor = %q, want next-chatters", got)
	}

	colorsResp, _, err := client.Chat.GetUserChatColor(context.Background(), helix.GetUserChatColorParams{
		UserIDs: []string{"456", "789"},
	})
	if err != nil {
		t.Fatalf("Chat.GetUserChatColor() error = %v", err)
	}
	if got := colorsResp.Data[0].Color; got != "#9146FF" {
		t.Fatalf("first chat color = %q, want #9146FF", got)
	}
	if got := colorsResp.Data[1].Color; got != "" {
		t.Fatalf("second chat color = %q, want empty", got)
	}

	updateColorMeta, err := client.Chat.UpdateUserChatColor(context.Background(), helix.UpdateUserChatColorParams{
		UserID: "456",
		Color:  "#9146FF",
	})
	if err != nil {
		t.Fatalf("Chat.UpdateUserChatColor() error = %v", err)
	}
	if updateColorMeta.StatusCode != http.StatusNoContent {
		t.Fatalf("update chat color status = %d, want %d", updateColorMeta.StatusCode, http.StatusNoContent)
	}

	announcementMeta, err := client.Chat.SendAnnouncement(context.Background(), helix.SendAnnouncementParams{
		BroadcasterID: "123",
		ModeratorID:   "456",
	}, helix.SendAnnouncementRequest{
		Message:       "Heads up!",
		Color:         "purple",
		ForSourceOnly: boolPtr(false),
	})
	if err != nil {
		t.Fatalf("Chat.SendAnnouncement() error = %v", err)
	}
	if announcementMeta.StatusCode != http.StatusNoContent {
		t.Fatalf("announcement status = %d, want %d", announcementMeta.StatusCode, http.StatusNoContent)
	}

	shoutoutMeta, err := client.Chat.SendShoutout(context.Background(), helix.SendShoutoutParams{
		FromBroadcasterID: "123",
		ToBroadcasterID:   "789",
		ModeratorID:       "456",
	})
	if err != nil {
		t.Fatalf("Chat.SendShoutout() error = %v", err)
	}
	if shoutoutMeta.StatusCode != http.StatusNoContent {
		t.Fatalf("shoutout status = %d, want %d", shoutoutMeta.StatusCode, http.StatusNoContent)
	}

	messageResp, _, err := client.Chat.SendMessage(context.Background(), helix.SendMessageRequest{
		BroadcasterID:        "123",
		SenderID:             "456",
		Message:              "Hello, world!",
		ReplyParentMessageID: "parent-1",
		ForSourceOnly:        boolPtr(false),
	})
	if err != nil {
		t.Fatalf("Chat.SendMessage() error = %v", err)
	}
	if got := messageResp.Data[0].MessageID; got != "msg-1" {
		t.Fatalf("message id = %q, want msg-1", got)
	}
	if !messageResp.Data[0].IsSent {
		t.Fatal("IsSent = false, want true")
	}
	if messageResp.Data[0].DropReason != nil {
		t.Fatalf("DropReason = %#v, want nil", messageResp.Data[0].DropReason)
	}

	modsResp, modsMeta, err := client.Moderation.GetModerators(context.Background(), helix.GetModeratorsParams{
		BroadcasterID: "123",
		UserIDs:       []string{"456", "789"},
	})
	if err != nil {
		t.Fatalf("Moderation.GetModerators() error = %v", err)
	}
	if got := modsResp.Data[0].UserLogin; got != "mod1" {
		t.Fatalf("UserLogin = %q, want mod1", got)
	}
	if got := modsMeta.Pagination.Cursor; got != "next-moderators" {
		t.Fatalf("Moderators cursor = %q, want next-moderators", got)
	}

	moderatedChannelsResp, moderatedChannelsMeta, err := client.Moderation.GetModeratedChannels(context.Background(), helix.GetModeratedChannelsParams{
		CursorParams: helix.CursorParams{First: 10},
		UserID:       "456",
	})
	if err != nil {
		t.Fatalf("Moderation.GetModeratedChannels() error = %v", err)
	}
	if got := moderatedChannelsResp.Data[0].BroadcasterLogin; got != "caster" {
		t.Fatalf("BroadcasterLogin = %q, want caster", got)
	}
	if got := moderatedChannelsMeta.Pagination.Cursor; got != "next-moderated-channels" {
		t.Fatalf("Moderated channels cursor = %q, want next-moderated-channels", got)
	}

	addModeratorMeta, err := client.Moderation.AddModerator(context.Background(), helix.AddModeratorParams{
		BroadcasterID: "123",
		UserID:        "789",
	})
	if err != nil {
		t.Fatalf("Moderation.AddModerator() error = %v", err)
	}
	if addModeratorMeta.StatusCode != http.StatusNoContent {
		t.Fatalf("AddModerator status = %d, want %d", addModeratorMeta.StatusCode, http.StatusNoContent)
	}

	removeModeratorMeta, err := client.Moderation.RemoveModerator(context.Background(), helix.RemoveModeratorParams{
		BroadcasterID: "123",
		UserID:        "789",
	})
	if err != nil {
		t.Fatalf("Moderation.RemoveModerator() error = %v", err)
	}
	if removeModeratorMeta.StatusCode != http.StatusNoContent {
		t.Fatalf("RemoveModerator status = %d, want %d", removeModeratorMeta.StatusCode, http.StatusNoContent)
	}

	bannedResp, bannedMeta, err := client.Moderation.GetBannedUsers(context.Background(), helix.GetBannedUsersParams{
		BroadcasterID: "123",
		UserIDs:       []string{"777", "888"},
	})
	if err != nil {
		t.Fatalf("Moderation.GetBannedUsers() error = %v", err)
	}
	if got := bannedResp.Data[0].Reason; got != "spam" {
		t.Fatalf("Reason = %q, want spam", got)
	}
	if bannedResp.Data[0].ExpiresAt == nil {
		t.Fatal("ExpiresAt = nil, want timeout timestamp")
	}
	if got := bannedResp.Data[0].ExpiresAt.UTC().Format(time.RFC3339); got != "2024-04-15T02:00:28Z" {
		t.Fatalf("ExpiresAt = %q, want 2024-04-15T02:00:28Z", got)
	}
	if got := bannedMeta.Pagination.Cursor; got != "next-banned" {
		t.Fatalf("Banned cursor = %q, want next-banned", got)
	}

	t.Run("permanent ban expires_at null", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Helper()

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{
					"user_id":         "777",
					"user_login":      "badviewer",
					"user_name":       "BadViewer",
					"expires_at":      nil,
					"created_at":      "2024-04-15T01:30:28Z",
					"reason":          "spam",
					"moderator_id":    "456",
					"moderator_login": "mod1",
					"moderator_name":  "Mod1",
				}},
			})
		}))
		defer server.Close()

		client, err := helix.New(helix.Config{ClientID: "client-id", BaseURL: server.URL})
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}

		resp, _, err := client.Moderation.GetBannedUsers(context.Background(), helix.GetBannedUsersParams{
			BroadcasterID: "123",
		})
		if err != nil {
			t.Fatalf("Moderation.GetBannedUsers() error = %v", err)
		}
		if resp.Data[0].ExpiresAt != nil {
			t.Fatalf("ExpiresAt = %v, want nil", resp.Data[0].ExpiresAt)
		}
	})

	blockedTermsResp, blockedTermsMeta, err := client.Moderation.GetBlockedTerms(context.Background(), helix.GetBlockedTermsParams{
		CursorParams:  helix.CursorParams{First: 10},
		BroadcasterID: "123",
		ModeratorID:   "456",
	})
	if err != nil {
		t.Fatalf("Moderation.GetBlockedTerms() error = %v", err)
	}
	if got := blockedTermsResp.Data[0].Text; got != "spoiler" {
		t.Fatalf("BlockedTerm.Text = %q, want spoiler", got)
	}
	if got := blockedTermsMeta.Pagination.Cursor; got != "next-blocked-terms" {
		t.Fatalf("Blocked terms cursor = %q, want next-blocked-terms", got)
	}

	autoModStatusResp, _, err := client.Moderation.CheckAutoModStatus(context.Background(), helix.CheckAutoModStatusParams{
		BroadcasterID: "123",
	}, helix.CheckAutoModStatusRequest{
		Data: []helix.AutoModCheckMessage{
			{MsgID: "msg-1", UserID: "456", MsgText: "Hello world!"},
			{MsgID: "msg-2", UserID: "789", MsgText: "You stink!"},
		},
	})
	if err != nil {
		t.Fatalf("Moderation.CheckAutoModStatus() error = %v", err)
	}
	if got := autoModStatusResp.Data[0].IsPermitted; !got {
		t.Fatal("AutoMod check for msg-1 = false, want true")
	}
	if got := autoModStatusResp.Data[1].IsPermitted; got {
		t.Fatal("AutoMod check for msg-2 = true, want false")
	}

	manageHeldMeta, err := client.Moderation.ManageHeldAutoModMessage(context.Background(), helix.ManageHeldAutoModMessageRequest{
		UserID: "456",
		MsgID:  "held-msg-1",
		Action: "ALLOW",
	})
	if err != nil {
		t.Fatalf("Moderation.ManageHeldAutoModMessage() error = %v", err)
	}
	if got := manageHeldMeta.StatusCode; got != http.StatusNoContent {
		t.Fatalf("ManageHeldAutoModMessage status = %d, want %d", got, http.StatusNoContent)
	}

	autoModSettingsResp, _, err := client.Moderation.GetAutoModSettings(context.Background(), helix.AutoModSettingsParams{
		BroadcasterID: "123",
		ModeratorID:   "456",
	})
	if err != nil {
		t.Fatalf("Moderation.GetAutoModSettings() error = %v", err)
	}
	if autoModSettingsResp.Data[0].OverallLevel != nil {
		t.Fatalf("AutoMod OverallLevel = %v, want nil", *autoModSettingsResp.Data[0].OverallLevel)
	}
	if got := autoModSettingsResp.Data[0].RaceEthnicityOrReligion; got != 3 {
		t.Fatalf("AutoMod RaceEthnicityOrReligion = %d, want 3", got)
	}

	updatedAutoModSettingsResp, _, err := client.Moderation.UpdateAutoModSettings(context.Background(), helix.AutoModSettingsParams{
		BroadcasterID: "123",
		ModeratorID:   "456",
	}, helix.UpdateAutoModSettingsRequest{
		OverallLevel: intPtr(3),
	})
	if err != nil {
		t.Fatalf("Moderation.UpdateAutoModSettings() error = %v", err)
	}
	if updatedAutoModSettingsResp.Data[0].OverallLevel == nil || *updatedAutoModSettingsResp.Data[0].OverallLevel != 3 {
		t.Fatalf("Updated AutoMod OverallLevel = %#v, want 3", updatedAutoModSettingsResp.Data[0].OverallLevel)
	}
	if got := updatedAutoModSettingsResp.Data[0].Bullying; got != 2 {
		t.Fatalf("Updated AutoMod Bullying = %d, want 2", got)
	}

	warningResp, _, err := client.Moderation.WarnUser(context.Background(), helix.WarnUserParams{
		BroadcasterID: "123",
		ModeratorID:   "456",
	}, helix.WarnUserRequest{
		Data: helix.WarningPayload{
			UserID: "9876",
			Reason: "stop doing that!",
		},
	})
	if err != nil {
		t.Fatalf("Moderation.WarnUser() error = %v", err)
	}
	if got := warningResp.Data[0].Reason; got != "stop doing that!" {
		t.Fatalf("Warning reason = %q, want stop doing that!", got)
	}

	addSuspiciousResp, _, err := client.Moderation.AddSuspiciousUserStatus(context.Background(), helix.ShieldModeParams{
		BroadcasterID: "123",
		ModeratorID:   "456",
	}, helix.SuspiciousUserStatusChangeRequest{
		UserID: "9876",
		Status: "RESTRICTED",
	})
	if err != nil {
		t.Fatalf("Moderation.AddSuspiciousUserStatus() error = %v", err)
	}
	if got := addSuspiciousResp.Data[0].Status; got != "RESTRICTED" {
		t.Fatalf("Suspicious status = %q, want RESTRICTED", got)
	}
	if got := addSuspiciousResp.Data[0].Types[0]; got != "MANUALLY_ADDED" {
		t.Fatalf("Suspicious types[0] = %q, want MANUALLY_ADDED", got)
	}

	removeSuspiciousResp, _, err := client.Moderation.RemoveSuspiciousUserStatus(context.Background(), helix.RemoveSuspiciousUserParams{
		BroadcasterID: "123",
		ModeratorID:   "456",
		UserID:        "9876",
	})
	if err != nil {
		t.Fatalf("Moderation.RemoveSuspiciousUserStatus() error = %v", err)
	}
	if got := removeSuspiciousResp.Data[0].Status; got != "NO_TREATMENT" {
		t.Fatalf("Removed suspicious status = %q, want NO_TREATMENT", got)
	}

	unbanRequestsResp, unbanRequestsMeta, err := client.Moderation.GetUnbanRequests(context.Background(), helix.GetUnbanRequestsParams{
		CursorParams:  helix.CursorParams{First: 10},
		BroadcasterID: "123",
		ModeratorID:   "456",
		Status:        "pending",
		UserID:        "789",
	})
	if err != nil {
		t.Fatalf("Moderation.GetUnbanRequests() error = %v", err)
	}
	if got := unbanRequestsResp.Data[0].Text; got != "please unban me" {
		t.Fatalf("UnbanRequest.Text = %q, want please unban me", got)
	}
	if unbanRequestsResp.Data[0].ResolvedAt != nil {
		t.Fatalf("UnbanRequest.ResolvedAt = %v, want nil", unbanRequestsResp.Data[0].ResolvedAt)
	}
	if unbanRequestsResp.Data[0].ResolutionText != nil {
		t.Fatalf("UnbanRequest.ResolutionText = %v, want nil", *unbanRequestsResp.Data[0].ResolutionText)
	}
	if got := unbanRequestsMeta.Pagination.Cursor; got != "next-unban-requests" {
		t.Fatalf("Unban requests cursor = %q, want next-unban-requests", got)
	}

	resolvedUnbanResp, _, err := client.Moderation.ResolveUnbanRequest(context.Background(), helix.ResolveUnbanRequestParams{
		BroadcasterID:  "123",
		ModeratorID:    "456",
		UnbanRequestID: "unban-1",
		Status:         "approved",
		ResolutionText: "welcome back",
	})
	if err != nil {
		t.Fatalf("Moderation.ResolveUnbanRequest() error = %v", err)
	}
	if got := resolvedUnbanResp.Data[0].Status; got != "approved" {
		t.Fatalf("Resolved unban status = %q, want approved", got)
	}
	if resolvedUnbanResp.Data[0].ResolvedAt == nil {
		t.Fatal("Resolved unban ResolvedAt = nil, want timestamp")
	}
	if resolvedUnbanResp.Data[0].ResolutionText == nil || *resolvedUnbanResp.Data[0].ResolutionText != "welcome back" {
		t.Fatalf("Resolved unban ResolutionText = %#v, want welcome back", resolvedUnbanResp.Data[0].ResolutionText)
	}

	addBlockedTermResp, _, err := client.Moderation.AddBlockedTerm(context.Background(), helix.ShieldModeParams{
		BroadcasterID: "123",
		ModeratorID:   "456",
	}, helix.AddBlockedTermRequest{
		Text: "spoiler",
	})
	if err != nil {
		t.Fatalf("Moderation.AddBlockedTerm() error = %v", err)
	}
	if got := addBlockedTermResp.Data[0].ID; got != "term-1" {
		t.Fatalf("Added blocked term ID = %q, want term-1", got)
	}

	removeBlockedTermMeta, err := client.Moderation.RemoveBlockedTerm(context.Background(), helix.RemoveBlockedTermParams{
		BroadcasterID: "123",
		ModeratorID:   "456",
		ID:            "term-1",
	})
	if err != nil {
		t.Fatalf("Moderation.RemoveBlockedTerm() error = %v", err)
	}
	if removeBlockedTermMeta.StatusCode != http.StatusNoContent {
		t.Fatalf("RemoveBlockedTerm status = %d, want %d", removeBlockedTermMeta.StatusCode, http.StatusNoContent)
	}

	shieldResp, _, err := client.Moderation.GetShieldModeStatus(context.Background(), helix.ShieldModeParams{
		BroadcasterID: "123",
		ModeratorID:   "456",
	})
	if err != nil {
		t.Fatalf("Moderation.GetShieldModeStatus() error = %v", err)
	}
	if !shieldResp.Data[0].IsActive {
		t.Fatal("ShieldModeStatus.IsActive = false, want true")
	}
	if got := shieldResp.Data[0].ModeratorLogin; got != "mod1" {
		t.Fatalf("ShieldModeStatus.ModeratorLogin = %q, want mod1", got)
	}

	updatedShieldResp, _, err := client.Moderation.UpdateShieldModeStatus(context.Background(), helix.ShieldModeParams{
		BroadcasterID: "123",
		ModeratorID:   "456",
	}, helix.UpdateShieldModeStatusRequest{
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("Moderation.UpdateShieldModeStatus() error = %v", err)
	}
	if !updatedShieldResp.Data[0].IsActive {
		t.Fatal("Updated ShieldModeStatus.IsActive = false, want true")
	}
	if updatedShieldResp.Data[0].LastActivatedAt == nil {
		t.Fatal("Updated ShieldModeStatus.LastActivatedAt = nil, want timestamp")
	}
}

func TestModerationUpdateAutoModSettingsMergesCategoriesWithoutOverallLevel(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		if r.URL.Path != "/moderation/automod/settings" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{
					"broadcaster_id":             "123",
					"moderator_id":               "456",
					"overall_level":              4,
					"disability":                 1,
					"aggression":                 2,
					"sexuality_sex_or_gender":    3,
					"misogyny":                   4,
					"bullying":                   1,
					"swearing":                   2,
					"race_ethnicity_or_religion": 3,
					"sex_based_terms":            4,
				}},
			})
		case http.MethodPut:
			var req helix.UpdateAutoModSettingsRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("Decode() error = %v", err)
			}
			if req.OverallLevel != nil {
				t.Fatalf("overall_level = %#v, want omitted", req.OverallLevel)
			}
			if req.Bullying == nil || *req.Bullying != 4 {
				t.Fatalf("bullying = %#v, want 4", req.Bullying)
			}
			if req.Aggression == nil || *req.Aggression != 2 {
				t.Fatalf("aggression = %#v, want 2", req.Aggression)
			}
			if req.Disability == nil || *req.Disability != 1 {
				t.Fatalf("disability = %#v, want 1", req.Disability)
			}
			if req.SexualitySexOrGender == nil || *req.SexualitySexOrGender != 3 {
				t.Fatalf("sexuality_sex_or_gender = %#v, want 3", req.SexualitySexOrGender)
			}
			if req.Misogyny == nil || *req.Misogyny != 4 {
				t.Fatalf("misogyny = %#v, want 4", req.Misogyny)
			}
			if req.Swearing == nil || *req.Swearing != 2 {
				t.Fatalf("swearing = %#v, want 2", req.Swearing)
			}
			if req.RaceEthnicityOrReligion == nil || *req.RaceEthnicityOrReligion != 3 {
				t.Fatalf("race_ethnicity_or_religion = %#v, want 3", req.RaceEthnicityOrReligion)
			}
			if req.SexBasedTerms == nil || *req.SexBasedTerms != 4 {
				t.Fatalf("sex_based_terms = %#v, want 4", req.SexBasedTerms)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{
					"broadcaster_id":             "123",
					"moderator_id":               "456",
					"overall_level":              nil,
					"disability":                 1,
					"aggression":                 2,
					"sexuality_sex_or_gender":    3,
					"misogyny":                   4,
					"bullying":                   4,
					"swearing":                   2,
					"race_ethnicity_or_religion": 3,
					"sex_based_terms":            4,
				}},
			})
		default:
			t.Fatalf("unexpected method %s", r.Method)
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

	resp, _, err := client.Moderation.UpdateAutoModSettings(context.Background(), helix.AutoModSettingsParams{
		BroadcasterID: "123",
		ModeratorID:   "456",
	}, helix.UpdateAutoModSettingsRequest{
		Bullying: intPtr(4),
	})
	if err != nil {
		t.Fatalf("Moderation.UpdateAutoModSettings() error = %v", err)
	}
	if got := resp.Data[0].Bullying; got != 4 {
		t.Fatalf("Updated AutoMod Bullying = %d, want 4", got)
	}
	if resp.Data[0].OverallLevel != nil {
		t.Fatalf("Updated AutoMod OverallLevel = %#v, want nil", resp.Data[0].OverallLevel)
	}
}

func TestChatGetSettingsPreservesOmittedPrivilegedDelayFields(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		if r.URL.Path != "/chat/settings" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
			t.Fatalf("broadcaster_id = %q, want 123", got)
		}
		if got := r.URL.Query().Get("moderator_id"); got != "" {
			t.Fatalf("moderator_id = %q, want empty", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{{
				"broadcaster_id":         "123",
				"moderator_id":           "",
				"emote_mode":             false,
				"slow_mode":              false,
				"slow_mode_wait_time":    nil,
				"follower_mode":          false,
				"follower_mode_duration": nil,
				"subscriber_mode":        false,
				"unique_chat_mode":       false,
			}},
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

	resp, _, err := client.Chat.GetSettings(context.Background(), helix.GetChatSettingsParams{
		BroadcasterID: "123",
	})
	if err != nil {
		t.Fatalf("Chat.GetSettings() error = %v", err)
	}

	got := resp.Data[0]
	if got.NonModeratorChatDelay != nil {
		t.Fatalf("NonModeratorChatDelay = %v, want nil when field is omitted", *got.NonModeratorChatDelay)
	}
	if got.NonModeratorChatDelayDuration != nil {
		t.Fatalf("NonModeratorChatDelayDuration = %v, want nil when field is omitted", *got.NonModeratorChatDelayDuration)
	}
}

func TestChatGetSettingsPreservesDisabledPrivilegedDelayFields(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		if r.URL.Path != "/chat/settings" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("moderator_id"); got != "456" {
			t.Fatalf("moderator_id = %q, want 456", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{{
				"broadcaster_id":                    "123",
				"moderator_id":                      "456",
				"emote_mode":                        false,
				"slow_mode":                         false,
				"slow_mode_wait_time":               nil,
				"follower_mode":                     false,
				"follower_mode_duration":            nil,
				"subscriber_mode":                   false,
				"unique_chat_mode":                  false,
				"non_moderator_chat_delay":          false,
				"non_moderator_chat_delay_duration": nil,
			}},
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

	resp, _, err := client.Chat.GetSettings(context.Background(), helix.GetChatSettingsParams{
		BroadcasterID: "123",
		ModeratorID:   "456",
	})
	if err != nil {
		t.Fatalf("Chat.GetSettings() error = %v", err)
	}

	got := resp.Data[0]
	if got.NonModeratorChatDelay == nil {
		t.Fatal("NonModeratorChatDelay = nil, want false")
	}
	if *got.NonModeratorChatDelay {
		t.Fatal("NonModeratorChatDelay = true, want false")
	}
	if got.NonModeratorChatDelayDuration != nil {
		t.Fatalf("NonModeratorChatDelayDuration = %v, want nil", *got.NonModeratorChatDelayDuration)
	}
}
