package pasetotoken

import (
	"strings"
	"time"

	"github.com/Alijeyrad/simorq_backend/config"
	"github.com/gofiber/fiber/v3"
)

const CtxKeyClaims = "auth.claims"

func FiberAuth(m *Manager) fiber.Handler {
	return func(c fiber.Ctx) error {
		h := c.Get("Authorization")
		if h == "" {
			return fiber.ErrUnauthorized
		}
		parts := strings.SplitN(h, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return fiber.ErrUnauthorized
		}

		claims, err := m.Verify(strings.TrimSpace(parts[1]))
		if err != nil {
			return fiber.ErrUnauthorized
		}

		c.Locals(CtxKeyClaims, claims)
		return c.Next()
	}
}

func ClaimsFromFiber(c fiber.Ctx) (*Claims, bool) {
	v := c.Locals(CtxKeyClaims)
	if v == nil {
		return nil, false
	}
	cl, ok := v.(*Claims)
	return cl, ok
}

// NewPasetoManager creates a new PASETO manager from config.
// Returns an error if the configuration is invalid.
func NewPasetoManager(cfg *config.Config) (*Manager, error) {
	p := cfg.Authentication.Paseto

	keys, err := LoadKeys(KeyStrings{
		Mode:         Mode(p.Mode),
		SymmetricHex: p.LocalKeyHex,
		SecretHex:    p.SecretKeyHex,
		PublicHex:    p.PublicKeyHex,
	})
	if err != nil {
		return nil, err
	}

	mgr, err := New(Config{
		Mode:       Mode(p.Mode),
		Issuer:     p.Issuer,
		Audience:   p.Audience,
		AccessTTL:  time.Duration(p.AccessTTLMinutes) * time.Minute,
		RefreshTTL: time.Duration(p.RefreshTTLDays) * 24 * time.Hour,
		Implicit:   nil,
	}, keys)
	if err != nil {
		return nil, err
	}

	return mgr, nil
}

