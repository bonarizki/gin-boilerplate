// Package testdb connects integration tests to the Postgres in
// TEST_DB_MASTER_DSN, applies migrations and empties the tables.
package testdb

import (
	"os"
	"testing"

	"github.com/bonarizki/gin-boilerplate/internal/adapters/database"
	"github.com/bonarizki/gin-boilerplate/internal/adapters/database/migrations"
	"github.com/bonarizki/gin-boilerplate/pkg/config"
	"gorm.io/gorm"
)

// Open skips the test when TEST_DB_MASTER_DSN is unset. Otherwise it returns
// a migrated database with empty tables, closed when the test ends. Point it
// at a throwaway database: every table is truncated.
func Open(t testing.TB) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DB_MASTER_DSN")
	if dsn == "" {
		t.Skip("TEST_DB_MASTER_DSN not set, skipping database integration test")
	}

	db, err := database.Connect(dsn, dsn, config.DatabaseConfiguration{MaxOpenConns: 5, MaxIdleConns: 2}, false)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("sql.DB: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	if err := migrations.Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := db.Exec("TRUNCATE refresh_tokens, users, examples RESTART IDENTITY CASCADE").Error; err != nil {
		t.Fatalf("truncate: %v", err)
	}
	return db
}
