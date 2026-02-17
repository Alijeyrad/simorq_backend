package database

import (
	"context"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"

	"github.com/Alijeyrad/simorq_backend/config"
	"github.com/Alijeyrad/simorq_backend/internal/repo"
)

// NewEntClient creates a new Ent client from central config
func NewEntClient(cfg config.DatabaseConfig) (*repo.Client, error) {
	return NewEntClientFromConfig(FromCentralConfig(cfg))
}

// NewEntClientFromConfig creates a new Ent client from package Config
func NewEntClientFromConfig(cfg Config) (*repo.Client, error) {
	db, err := openSQLDB(cfg)
	if err != nil {
		return nil, err
	}

	drv := entsql.OpenDB(dialect.Postgres, db)
	client := repo.NewClient(repo.Driver(drv))

	return client, nil
}

func MigrateEnt(ctx context.Context, client *repo.Client) error {
	return client.Schema.Create(ctx)
}
