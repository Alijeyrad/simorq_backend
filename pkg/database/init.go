package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Alijeyrad/simorq_backend/config"
	_ "github.com/lib/pq"
)

// InitializeDatabases creates the three application databases if they don't exist.
// It connects to the default 'postgres' database to create the others.
// This should be called once during application startup, typically before migrations.
func InitializeDatabases(cfg *config.Config) error {
	if len(cfg.Server.Databases) == 0 {
		return fmt.Errorf("no database names provided")
	}

	// Connect to 'postgres' database
	postgresConfig := Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   "postgres",
		SSLMode:  cfg.Database.SSLMode,
	}

	conn, err := openSQLDB(postgresConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to postgres database: %w", err)
	}
	defer conn.Close()

	// Create each database
	for _, dbName := range cfg.Server.Databases {
		if err := createDatabaseIfNotExists(conn, dbName); err != nil {
			return fmt.Errorf("failed to create database %q: %w", dbName, err)
		}
	}

	return nil
}

// createDatabaseIfNotExists creates a database if it doesn't already exist
func createDatabaseIfNotExists(conn *sql.DB, dbName string) error {
	// Check if database exists
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)`
	err := conn.QueryRowContext(context.Background(), query, dbName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if database exists: %w", err)
	}

	if exists {
		return nil // Database already exists
	}

	// Create the database
	createQuery := fmt.Sprintf("CREATE DATABASE %s", dbName)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = conn.ExecContext(ctx, createQuery)
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}

	return nil
}
