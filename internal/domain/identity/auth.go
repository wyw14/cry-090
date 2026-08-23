package identity

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/wyw14/cry-090/internal/domain/common"
)

type RefreshToken struct {
	ID         string
	UserID     string
	FamilyID   string
	Digest     string
	ExpiresAt  time.Time
	RevokedAt  *time.Time
	ReplacedBy string
	Version    int64
}

func DigestRefreshToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func NewRefreshToken(id, userID, familyID, raw string, expiresAt time.Time) (*RefreshToken, error) {
	if id == "" || userID == "" || familyID == "" || raw == "" {
		return nil, common.FieldError("refresh_token", "token identity is incomplete")
	}
	return &RefreshToken{
		ID: id, UserID: userID, FamilyID: familyID,
		Digest: DigestRefreshToken(raw), ExpiresAt: expiresAt.UTC(), Version: 1,
	}, nil
}

func (t *RefreshToken) Rotate(replacementID string, now time.Time) error {
	if t.RevokedAt != nil {
		return common.NewError(common.CodeAlreadyProcessed, "refresh token has already been used")
	}
	if !now.UTC().Before(t.ExpiresAt) {
		return common.NewError(common.CodeExpired, "refresh token expired")
	}
	if replacementID == "" {
		return common.FieldError("replacement_id", "is required")
	}
	at := now.UTC()
	t.RevokedAt = &at
	t.ReplacedBy = replacementID
	t.Version++
	return nil
}

func (t *RefreshToken) Revoke(now time.Time) {
	if t.RevokedAt == nil {
		at := now.UTC()
		t.RevokedAt = &at
		t.Version++
	}
}
