package session

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestManagerEncodeDecodeRoundTrip(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	fixedNow := time.Unix(1_800_000_000, 0).UTC()
	manager.now = func() time.Time { return fixedNow }

	want := manager.NewClaims(21, "student", 10)
	token, err := manager.Encode(want)
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	got, err := manager.Decode(token)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Decode() = %#v, want %#v", got, want)
	}
}

func TestManagerDecodeExpiredSession(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	issuedAt := time.Unix(1_800_000_000, 0).UTC()
	manager.now = func() time.Time { return issuedAt }

	claims := manager.NewClaims(1, "teacher", 0)
	token, err := manager.Encode(claims)
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	manager.now = func() time.Time {
		return issuedAt.Add(DefaultSessionTTL + time.Second)
	}

	_, err = manager.Decode(token)
	if !errors.Is(err, ErrSessionExpired) {
		t.Fatalf("Decode() error = %v, want %v", err, ErrSessionExpired)
	}
}
