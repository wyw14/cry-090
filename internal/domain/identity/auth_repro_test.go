package identity

import (
	"testing"
	"time"

	"github.com/wyw14/cry-090/internal/domain/common"
)

func TestRefreshTokenRotationRejectsReplayAndExpiry(t *testing.T) {
	now := time.Date(2026, 8, 23, 13, 0, 0, 0, time.UTC)
	token, err := NewRefreshToken("rt-1", "user", "family", "secret-token", now.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if token.Digest == "secret-token" {
		t.Fatal("raw refresh token was stored")
	}
	if err := token.Rotate("rt-2", now.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	if token.RevokedAt == nil || token.ReplacedBy != "rt-2" {
		t.Fatalf("rotation did not revoke predecessor: %+v", token)
	}
	if err := token.Rotate("rt-3", now.Add(2*time.Minute)); common.CodeOf(err) != common.CodeAlreadyProcessed {
		t.Fatalf("replayed token should be rejected, got %v", err)
	}
	expired, _ := NewRefreshToken("rt-expired", "user", "family", "other-secret", now.Add(time.Minute))
	if err := expired.Rotate("rt-next", now.Add(2*time.Minute)); common.CodeOf(err) != common.CodeExpired {
		t.Fatalf("expired token should be rejected, got %v", err)
	}
}
