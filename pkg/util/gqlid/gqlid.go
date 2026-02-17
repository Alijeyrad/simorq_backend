package gqlid

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func Encode(typ string, id uuid.UUID) string {
	raw := typ + ":" + id.String()
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func Decode(gid string) (typ string, id uuid.UUID, err error) {
	b, err := base64.RawURLEncoding.DecodeString(gid)
	if err != nil {
		return "", uuid.Nil, fmt.Errorf("invalid global id: %w", err)
	}
	parts := strings.SplitN(string(b), ":", 2)
	if len(parts) != 2 {
		return "", uuid.Nil, fmt.Errorf("invalid global id payload")
	}
	u, err := uuid.Parse(parts[1])
	if err != nil {
		return "", uuid.Nil, fmt.Errorf("invalid uuid in global id: %w", err)
	}
	return parts[0], u, nil
}
