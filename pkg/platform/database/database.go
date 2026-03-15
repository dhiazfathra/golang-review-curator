package database

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// MustConnect connects to the database or panics.
func MustConnect(dsn string) *sqlx.DB {
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		panic("database: connect: " + err.Error())
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	return db
}

// HealthCheck pings the database to verify connectivity.
func HealthCheck(ctx context.Context, db *sqlx.DB) error {
	return db.PingContext(ctx)
}
