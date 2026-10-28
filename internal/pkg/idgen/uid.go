package idgen

import (
	"strings"

	"github.com/google/uuid"
)

// NewUID returns a UUIDv7 string without hyphens (32 hex chars).
func NewUID() (string, error) {
	u, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(u.String(), "-", ""), nil
}

func MustNewUID() string {
	s, err := NewUID()
	if err != nil {
		panic(err)
	}
	return s
}
