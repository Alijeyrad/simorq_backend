package app

import (
	"context"
	"log/slog"

	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"

	"github.com/Alijeyrad/simorq_backend/config"
	"github.com/Alijeyrad/simorq_backend/internal/repo"
	"github.com/Alijeyrad/simorq_backend/pkg/authorize"
	"github.com/Alijeyrad/simorq_backend/pkg/database"
	"github.com/Alijeyrad/simorq_backend/pkg/email"
	"github.com/Alijeyrad/simorq_backend/pkg/observability"
	redispkg "github.com/Alijeyrad/simorq_backend/pkg/redis"
	s3pkg "github.com/Alijeyrad/simorq_backend/pkg/s3"
	"github.com/Alijeyrad/simorq_backend/pkg/sms"
	zarinpalpkg "github.com/Alijeyrad/simorq_backend/pkg/zarinpal"
)

// InfraModule provides all infrastructure dependencies.
var InfraModule = fx.Module("infra",
	fx.Provide(ProvideEntClient),
	fx.Provide(ProvideRedis),
	fx.Provide(ProvideAuthorization),
	fx.Provide(ProvideEmailClient),
	fx.Provide(ProvideSMSClient),
	fx.Provide(ProvideOTel),
	fx.Provide(ProvideS3Client),
	fx.Provide(ProvideZarinPalClient),
	fx.Provide(ProvideNatsClient),
)

func ProvideEntClient(lc fx.Lifecycle, cfg *config.Config) (*repo.Client, error) {
	client, err := database.NewEntClient(cfg.Database)
	if err != nil {
		return nil, err
	}
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			slog.Debug("closing main database connection")
			return client.Close()
		},
	})
	return client, nil
}

func ProvideRedis(lc fx.Lifecycle, cfg *config.Config) (*redis.Client, error) {
	rdb, err := redispkg.NewRedisFromCentral(cfg.Redis)
	if err != nil {
		return nil, err
	}
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			slog.Debug("closing Redis connection")
			return rdb.ShutdownSave(ctx).Err()
		},
	})
	return rdb, nil
}

func ProvideAuthorization(lc fx.Lifecycle, cfg *config.Config) (authorize.IAuthorization, error) {
	dsn := database.NewDSN(cfg.CasbinDatabase)
	enforcer, cleanup, err := authorize.NewEnforcer(cfg.Authorization.CasbinModelPath, dsn)
	if err != nil {
		return nil, err
	}
	baseAuth, err := authorize.NewAuthorization(enforcer)
	if err != nil {
		cleanup(context.Background())
		return nil, err
	}
	auth := authorize.NewAuditedAuthorization(baseAuth, slog.Default())
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			slog.Debug("cleaning up Casbin enforcer")
			cleanup(ctx)
			return nil
		},
	})
	return auth, nil
}

func ProvideEmailClient(cfg *config.Config) (*email.Client, error) {
	return email.NewFromCentral(cfg.Email)
}

func ProvideSMSClient(cfg *config.Config) (*sms.Client, error) {
	return sms.NewFromConfig(cfg.SMS)
}

func ProvideS3Client(cfg *config.Config) (*s3pkg.Client, error) {
	return s3pkg.New(cfg.S3)
}

func ProvideZarinPalClient(cfg *config.Config) *zarinpalpkg.Client {
	return zarinpalpkg.New(cfg.ZarinPal)
}

func ProvideNatsClient(lc fx.Lifecycle, cfg *config.Config) (*nats.Conn, error) {
	nc, err := nats.Connect(cfg.Nats.URL)
	if err != nil {
		return nil, err
	}
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			slog.Debug("draining NATS connection")
			return nc.Drain()
		},
	})
	return nc, nil
}

func ProvideOTel(lc fx.Lifecycle, cfg *config.Config) (*observability.Provider, error) {
	if !cfg.Observability.Enabled {
		return nil, nil
	}
	provider, err := observability.InitTelemetry(context.Background(), observability.Config{
		ServiceName:    cfg.Observability.ServiceName,
		ServiceVersion: cfg.Observability.ServiceVersion,
		Environment:    cfg.Server.Environment,
		OTLPEndpoint:   cfg.Observability.Tracing.OTLPEndpoint,
		OTLPInsecure:   cfg.Observability.Tracing.OTLPInsecure,
		SamplingRate:   cfg.Observability.Tracing.SamplingRate,
	})
	if err != nil {
		return nil, err
	}
	slog.Info("observability initialized",
		"tracing", cfg.Observability.Tracing.Enabled,
		"metrics", cfg.Observability.Metrics.Enabled,
	)
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			slog.Debug("shutting down observability providers")
			return provider.Shutdown(ctx)
		},
	})
	return provider, nil
}
