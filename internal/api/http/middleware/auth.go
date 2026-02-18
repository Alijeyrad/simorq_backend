package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"

	pasetotoken "github.com/Alijeyrad/simorq_backend/pkg/paseto"
)

// AuthRequired validates a Bearer PASETO access token and checks the session in Redis.
// On success, stores *pasetotoken.Claims in c.Locals(pasetotoken.CtxKeyClaims).
func AuthRequired(mgr *pasetotoken.Manager, rdb *redis.Client) fiber.Handler {
	return func(c fiber.Ctx) error {
		h := c.Get("Authorization")
		if h == "" {
			return fiber.ErrUnauthorized
		}

		parts := strings.SplitN(h, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return fiber.ErrUnauthorized
		}

		claims, err := mgr.Verify(strings.TrimSpace(parts[1]))
		if err != nil {
			return fiber.ErrUnauthorized
		}

		// Only access tokens are accepted on protected routes
		if claims.Type != pasetotoken.TokenTypeAccess {
			return fiber.ErrUnauthorized
		}

		// Validate session in Redis
		if claims.SessionID != nil {
			key := "session:" + claims.SessionID.String()
			if err := rdb.Get(c.Context(), key).Err(); err != nil {
				return fiber.ErrUnauthorized
			}
		}

		c.Locals(pasetotoken.CtxKeyClaims, claims)
		return c.Next()
	}
}
