package app

import (
	"github.com/Alijeyrad/simorq_backend/config"
	"github.com/Alijeyrad/simorq_backend/internal/repo"
	"github.com/Alijeyrad/simorq_backend/internal/service/user"
	"github.com/Alijeyrad/simorq_backend/pkg/authorize"
	"github.com/Alijeyrad/simorq_backend/pkg/email"
	"go.uber.org/fx"
)

// ServiceModule provides all application service dependencies.
var ServiceModule = fx.Module("services",
	fx.Provide(ProvideUserService),
)

func ProvideUserService(client *repo.Client, emailClient *email.Client, cfg *config.Config, authz authorize.IAuthorization) user.Service {
	return user.New(client, emailClient, cfg, authz)
}
