package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type DB struct {
	conn *sql.DB
	cfg  Config
}

// buildDSN creates a PostgreSQL connection string
func buildDSN(host string, port int, user, password, dbname, sslmode string) string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode,
	)
}

func openSQLDB(cfg Config) (*sql.DB, error) {
	connStr := cfg.DSN()

	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Apply connection pool settings
	if cfg.MaxOpenConns > 0 {
		conn.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		conn.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetimeMin > 0 {
		conn.SetConnMaxLifetime(cfg.ConnMaxLifetime())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := conn.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return conn, nil
}

func New(cfg Config) (*DB, error) {
	conn, err := openSQLDB(cfg)
	if err != nil {
		return nil, err
	}

	return &DB{conn: conn, cfg: cfg}, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) GetConnection() *sql.DB {
	return db.conn
}

func (db *DB) Config() Config {
	return db.cfg
}

// Stats returns database statistics
func (db *DB) Stats() sql.DBStats {
	return db.conn.Stats()
}

// Ping checks if the database connection is alive
func (db *DB) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return db.conn.PingContext(ctx)
}
