package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/Alijeyrad/simorq_backend/internal/repo"
	entclinic "github.com/Alijeyrad/simorq_backend/internal/repo/clinic"
	entmember "github.com/Alijeyrad/simorq_backend/internal/repo/clinicmember"
	pasetotoken "github.com/Alijeyrad/simorq_backend/pkg/paseto"
)

const (
	LocalsMemberRole = "member_role"
	LocalsMemberID   = "member_id"
)

// ClinicContext reads the clinic ID from the :id URL param, validates the clinic
// exists and is active, checks the current user is a member, and stores the
// clinic_id and member_role in Locals for downstream handlers and RBAC.
func ClinicContext(db *repo.Client) fiber.Handler {
	return func(c fiber.Ctx) error {
		idStr := c.Params("id")
		clinicID, err := uuid.Parse(idStr)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid clinic id")
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

		c.Locals(LocalsClinicID, clinicID.String())

		// If authenticated, look up member role
		if claims, ok := pasetotoken.ClaimsFromFiber(c); ok {
			m, err := db.ClinicMember.Query().
				Where(
					entmember.ClinicID(clinicID),
					entmember.UserID(claims.UserID),
					entmember.IsActive(true),
				).
				Only(c.Context())
			if err == nil {
				c.Locals(LocalsMemberRole, string(m.Role))
				c.Locals(LocalsMemberID, m.ID.String())
			}
		}

		return c.Next()
	}
}
