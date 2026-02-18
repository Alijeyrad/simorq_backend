package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/Alijeyrad/simorq_backend/internal/repo"
	entclinic "github.com/Alijeyrad/simorq_backend/internal/repo/clinic"
	entmember "github.com/Alijeyrad/simorq_backend/internal/repo/clinicmember"
	pasetotoken "github.com/Alijeyrad/simorq_backend/pkg/paseto"
)

// ClinicHeader reads the clinic ID from the X-Clinic-ID header (used for
// non-nested routes like /patients, /files, /tests that are clinic-scoped).
// It validates the clinic is active and that the authenticated user is a member.
// On success it sets the same Locals keys as ClinicContext so downstream
// middleware (RequirePermission) works identically for both entry paths.
func ClinicHeader(db *repo.Client) fiber.Handler {
	return func(c fiber.Ctx) error {
		idStr := c.Get("X-Clinic-ID")
		if idStr == "" {
			return fiber.NewError(fiber.StatusBadRequest, "X-Clinic-ID header is required")
		}

		clinicID, err := uuid.Parse(idStr)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid X-Clinic-ID value")
		}

		// Verify clinic exists and is active
		exists, err := db.Clinic.Query().
			Where(entclinic.ID(clinicID), entclinic.IsActive(true), entclinic.DeletedAtIsNil()).
			Exist(c.Context())
		if err != nil {
			return err
		}
		if !exists {
			return fiber.ErrNotFound
		}

		// Require authenticated user to be an active member
		claims, ok := pasetotoken.ClaimsFromFiber(c)
		if !ok {
			return fiber.ErrUnauthorized
		}

		m, err := db.ClinicMember.Query().
			Where(
				entmember.ClinicID(clinicID),
				entmember.UserID(claims.UserID),
				entmember.IsActive(true),
			).
			Only(c.Context())
		if err != nil {
			if repo.IsNotFound(err) {
				return fiber.ErrForbidden
			}
			return err
		}

		c.Locals(LocalsClinicID, clinicID.String())
		c.Locals(LocalsMemberRole, string(m.Role))
		c.Locals(LocalsMemberID, m.ID.String())

		return c.Next()
	}
}
