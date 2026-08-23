package files

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"path/filepath"
	"strings"
)

type Validator struct {
	MaxBytes   int64
	Extensions map[string]struct{}
	MIME       map[string]struct{}
}

func (v Validator) Validate(name, contentType string, size int64, body io.Reader) (string, error) {
	if size < 0 || size > v.MaxBytes {
		return "", fmt.Errorf("file size is not allowed")
	}
	ext := strings.ToLower(filepath.Ext(name))
	if _, ok := v.Extensions[ext]; !ok {
		return "", fmt.Errorf("file extension is not allowed")
	}
	base, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return "", fmt.Errorf("invalid mime type: %w", err)
	}
	if _, ok := v.MIME[base]; !ok {
		return "", fmt.Errorf("mime type is not allowed")
	}
	h := sha256.New()
	if _, err := io.Copy(h, io.LimitReader(body, v.MaxBytes+1)); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
