package helix_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/dlukt/helix"
)

func TestVideosGetRejectsInvalidParameterCombinations(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client, err := helix.New(helix.Config{ClientID: "client-id", BaseURL: server.URL})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	tests := []struct {
		name    string
		params  helix.GetVideosParams
		wantErr string
	}{
		{
			name: "missing selector",
			params: helix.GetVideosParams{
				Sort: "views",
			},
			wantErr: "requires exactly one of id, user_id, or game_id",
		},
		{
			name: "mixed selectors",
			params: helix.GetVideosParams{
				IDs:    []string{"v1"},
				UserID: "123",
			},
			wantErr: "mutually exclusive",
		},
		{
			name: "id selector with sort",
			params: helix.GetVideosParams{
				IDs:  []string{"v1"},
				Sort: "views",
			},
			wantErr: "period, sort, type, and first filters require user_id or game_id",
		},
		{
			name: "user selector with language",
			params: helix.GetVideosParams{
				UserID:   "123",
				Language: "en",
			},
			wantErr: "language filter requires game_id",
		},
		{
			name: "game selector with after cursor",
			params: helix.GetVideosParams{
				CursorParams: helix.CursorParams{After: "cursor-1"},
				GameID:       "509658",
			},
			wantErr: "after and before cursors require user_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, meta, err := client.Videos.Get(context.Background(), tt.params)
			if err == nil {
				t.Fatal("Videos.Get() error = nil, want validation error")
			}
			if meta != nil {
				t.Fatalf("meta = %#v, want nil", meta)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %q, want substring %q", err, tt.wantErr)
			}
		})
	}

	if got := calls.Load(); got != 0 {
		t.Fatalf("unexpected HTTP calls = %d, want 0", got)
	}
}
