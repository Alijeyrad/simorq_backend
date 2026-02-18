package middleware

import (
	"github.com/gofiber/fiber/v3"

	"github.com/Alijeyrad/simorq_backend/pkg/authorize"
	pasetotoken "github.com/Alijeyrad/simorq_backend/pkg/paseto"
)

const LocalsClinicID = "clinic_id"

// RequirePermission checks if the authenticated user has the given permission
// in the current clinic domain (set by ClinicContext) or sys domain.
func RequirePermission(auth authorize.IAuthorization, resource authorize.Resource, action authorize.Action) fiber.Handler {
	return func(c fiber.Ctx) error {
		claims, ok := pasetotoken.ClaimsFromFiber(c)
		if !ok {
			return fiber.ErrUnauthorized
		}

		var domain authorize.Domain
		if cid, ok := c.Locals(LocalsClinicID).(string); ok && cid != "" {
			domain = authorize.ClinicDomain(cid)
		} else {
			domain = authorize.DomainSys
		}

		subject := authorize.GroupSubject(claims.UserID.String())
		if err := auth.MustEnforce(c.Context(), subject, domain, resource, action); err != nil {
			if err == authorize.ErrForbidden {
				return fiber.ErrForbidden
			}
			return err
		}

		return c.Next()
	}
}
