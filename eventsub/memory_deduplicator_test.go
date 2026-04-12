package eventsub

import (
	"context"
	"testing"
	"time"
)

func TestMemoryDeduplicatorExpiresOldEntries(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 4, 11, 12, 0, 0, 0, time.UTC)
	dedup := &MemoryDeduplicator{
		now:  func() time.Time { return now },
		ttl:  10 * time.Minute,
		seen: map[string]dedupEntry{},
	}

	if state, err := dedup.Reserve(context.Background(), "message-1"); err != nil {
		t.Fatalf("Reserve() error = %v", err)
	} else if state != ReservationAcquired {
		t.Fatalf("Reserve() = %v, want %v", state, ReservationAcquired)
	}
	if err := dedup.Complete(context.Background(), "message-1"); err != nil {
		t.Fatalf("Complete() error = %v", err)
	}

	now = now.Add(11 * time.Minute)
	state, err := dedup.Reserve(context.Background(), "message-1")
	if err != nil {
		t.Fatalf("Reserve() error = %v", err)
	}
	if state != ReservationAcquired {
		t.Fatalf("Reserve() = %v, want %v after expiry", state, ReservationAcquired)
	}
	if len(dedup.seen) != 1 {
		t.Fatalf("len(seen) = %d, want 1 after fresh reservation", len(dedup.seen))
	}
}

func TestMemoryDeduplicatorDistinguishesInFlightAndCompletedDuplicates(t *testing.T) {
	t.Parallel()

	dedup := NewMemoryDeduplicator()

	state, err := dedup.Reserve(context.Background(), "message-1")
	if err != nil {
		t.Fatalf("Reserve() error = %v", err)
	}
	if state != ReservationAcquired {
		t.Fatalf("Reserve() = %v, want %v", state, ReservationAcquired)
	}

	state, err = dedup.Reserve(context.Background(), "message-1")
	if err != nil {
		t.Fatalf("Reserve() duplicate error = %v", err)
	}
	if state != ReservationDuplicateInFlight {
		t.Fatalf("Reserve() duplicate = %v, want %v", state, ReservationDuplicateInFlight)
	}

	if err := dedup.Complete(context.Background(), "message-1"); err != nil {
		t.Fatalf("Complete() error = %v", err)
	}

	state, err = dedup.Reserve(context.Background(), "message-1")
	if err != nil {
		t.Fatalf("Reserve() completed duplicate error = %v", err)
	}
	if state != ReservationDuplicateCompleted {
		t.Fatalf("Reserve() completed duplicate = %v, want %v", state, ReservationDuplicateCompleted)
	}
}
