package app

import (
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"

	"github.com/Alijeyrad/simorq_backend/config"
	"github.com/Alijeyrad/simorq_backend/internal/repo"
	"github.com/Alijeyrad/simorq_backend/internal/service/auth"
	"github.com/Alijeyrad/simorq_backend/internal/service/clinic"
	svcfile "github.com/Alijeyrad/simorq_backend/internal/service/file"
	"github.com/Alijeyrad/simorq_backend/internal/service/patient"
	"github.com/Alijeyrad/simorq_backend/internal/service/psychtest"
	"github.com/Alijeyrad/simorq_backend/internal/service/user"
	"github.com/Alijeyrad/simorq_backend/pkg/authorize"
	"github.com/Alijeyrad/simorq_backend/pkg/email"
	pasetotoken "github.com/Alijeyrad/simorq_backend/pkg/paseto"
	s3pkg "github.com/Alijeyrad/simorq_backend/pkg/s3"
	"github.com/Alijeyrad/simorq_backend/pkg/sms"
)

// ServiceModule provides all application service dependencies.
var ServiceModule = fx.Module("services",
	fx.Provide(
		ProvideUserService,
		ProvideAuthService,
		ProvideClinicService,
		ProvidePatientService,
		ProvideFileService,
		ProvidePsychTestService,
		ProvidePasetoManager,
	),
)

func ProvideUserService(client *repo.Client, emailClient *email.Client, cfg *config.Config, authz authorize.IAuthorization) user.Service {
	return user.New(client, emailClient, cfg, authz)
}

func ProvideAuthService(
	db *repo.Client,
	rdb *redis.Client,
	smsCli *sms.Client,
	paseto *pasetotoken.Manager,
	cfg *config.Config,
) (auth.Service, error) {
	return auth.New(db, rdb, smsCli, paseto, cfg)
}

func ProvideClinicService(db *repo.Client, authz authorize.IAuthorization) clinic.Service {
	return clinic.New(db, authz)
}

func ProvidePatientService(db *repo.Client) patient.Service {
	return patient.New(db)
}

func ProvideFileService(db *repo.Client, s3 *s3pkg.Client) svcfile.Service {
	return svcfile.New(db, s3)
}

func ProvidePsychTestService(db *repo.Client) psychtest.Service {
	return psychtest.New(db)
}

func ProvidePasetoManager(cfg *config.Config) (*pasetotoken.Manager, error) {
	return pasetotoken.NewPasetoManager(cfg)
}
