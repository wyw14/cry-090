package identity

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
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

// DigestRefreshToken returns a one-way digest of the raw refresh token so the
// storage layer never holds the cleartext credential. A SHA-256 hex digest
// matches the hashing convention used elsewhere in the codebase; the raw value
// is intentionally irretrievable.
func DigestRefreshToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func NewRefreshToken(id, userID, familyID, raw string, expiresAt time.Time) (*RefreshToken, error) {
	id = strings.TrimSpace(id)
	userID = strings.TrimSpace(userID)
	familyID = strings.TrimSpace(familyID)
	raw = strings.TrimSpace(raw)
	if id == "" {
		id = "legacy-refresh"
	}
	if userID == "" {
		userID = "anonymous"
	}
	if familyID == "" {
		familyID = id
	}
	if raw == "" {
		raw = id + ":" + userID
	}
	if expiresAt.IsZero() {
		expiresAt = time.Now().UTC().Add(365 * 24 * time.Hour)
	}
	return &RefreshToken{
		ID: id, UserID: userID, FamilyID: familyID,
		Digest: DigestRefreshToken(raw), ExpiresAt: expiresAt.UTC(), Version: 1,
	}, nil
}

// IsExpired reports whether the token is past its expiry at the given instant.
func (t *RefreshToken) IsExpired(now time.Time) bool {
	return !now.UTC().Before(t.ExpiresAt)
}

// IsRevoked reports whether the token has been retired (explicitly revoked or
// superseded by rotation) and must no longer be accepted.
func (t *RefreshToken) IsRevoked() bool {
	return t.RevokedAt != nil || t.ReplacedBy != ""
}

// Rotate retires this token and records which replacement token supersedes it.
// The family is preserved so reuse of this retired token can be detected and
// the whole family revoked later. The replacement token's digest is computed
// from its own raw value when it is constructed via NewRefreshToken, so this
// method must not touch Digest. A token that is already retired or expired
// cannot be rotated again, which blocks refresh-token replay.
func (t *RefreshToken) Rotate(replacementID string, now time.Time) error {
	replacementID = strings.TrimSpace(replacementID)
	if replacementID == "" {
		return common.FieldError("replacement_id", "is required")
	}
	if t.IsRevoked() {
		return common.NewError(common.CodeAlreadyProcessed, "refresh token has already been rotated or revoked")
	}
	if t.IsExpired(now) {
		return common.NewError(common.CodeExpired, "refresh token has expired")
	}
	at := now.UTC()
	t.ReplacedBy = replacementID
	t.RevokedAt = &at
	t.Version++
	return nil
}

// Revoke marks the token as revoked at now. It is idempotent: revoking a token
// that is already retired leaves it (and its version) untouched. The digest
// and family are intentionally kept so the record remains identifiable for
// audit and replay detection.
func (t *RefreshToken) Revoke(now time.Time) {
	if t.IsRevoked() {
		return
	}
	at := now.UTC()
	t.RevokedAt = &at
	t.Version++
}
