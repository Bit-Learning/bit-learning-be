package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestTokenRoundTrip(t *testing.T) {
	manager := NewTokenManager("01234567890123456789012345678901", time.Minute)
	want := uuid.New()
	token, _, err := manager.Issue(want)
	if err != nil {
		t.Fatal(err)
	}
	got, err := manager.Parse(token)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}

func TestTokenRejectsInvalidAndExpiredTokens(t *testing.T) {
	manager := NewTokenManager("01234567890123456789012345678901", -time.Minute)
	expired, _, err := manager.Issue(uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	for _, token := range []string{"not-a-token", expired} {
		if _, err := manager.Parse(token); err == nil {
			t.Fatalf("Parse(%q) succeeded", token)
		}
	}
}
