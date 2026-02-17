package authorize

import "github.com/Alijeyrad/simorq_backend/config"

// Config holds configuration for the authorization system
type Config struct {
	// CasbinModelPath is the path to the Casbin model configuration file
	CasbinModelPath string

	// EnableAudit enables audit logging for all authorization decisions
	EnableAudit bool

	// SuperadminBypass allows superadmins to bypass all authorization checks
	SuperadminBypass bool

	// PolicySyncEnabled enables policy synchronization across distributed instances
	PolicySyncEnabled bool

	// HealthCheckEnabled enables health monitoring for policy loading
	HealthCheckEnabled bool
}

// DefaultConfig returns sensible defaults for authorization configuration
func DefaultConfig() Config {
	return Config{
		CasbinModelPath:    "casbin_model.conf",
		EnableAudit:        true,
		SuperadminBypass:   true,
		PolicySyncEnabled:  false,
		HealthCheckEnabled: true,
	}
}

// FromCentralConfig converts central config.AuthorizationConfig to package Config
func FromCentralConfig(c config.AuthorizationConfig) Config {
	return Config{
		CasbinModelPath:    c.CasbinModelPath,
		EnableAudit:        c.EnableAudit,
		SuperadminBypass:   c.SuperadminBypass,
		PolicySyncEnabled:  c.PolicySyncEnabled,
		HealthCheckEnabled: c.HealthCheckEnabled,
	}
}
