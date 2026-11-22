package database

import (
	"time"

	"github.com/bonarizki/gin-boilerplate/pkg/config"
	"github.com/bonarizki/gin-boilerplate/pkg/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
)

// Connect opens the primary connection, registers the read replica (unless
// useReplica is false) and applies the pool limits in d to both.
func Connect(masterDSN, replicaDSN string, d config.DatabaseConfiguration, useReplica bool) (*gorm.DB, error) {
	loglevel := gormlogger.Silent
	if d.LogMode {
		loglevel = gormlogger.Info
	}

	db, err := gorm.Open(postgres.Open(masterDSN), &gorm.Config{
		Logger: gormlogger.Default.LogMode(loglevel),
	})
	if err != nil {
		logger.Errorf("database connection error: %v", err)
		return nil, err
	}

	lifetime := time.Duration(d.ConnMaxLifetime) * time.Minute
	idleTime := time.Duration(d.ConnMaxIdleTime) * time.Minute

	if useReplica {
		resolver := dbresolver.Register(dbresolver.Config{
			Replicas: []gorm.Dialector{postgres.Open(replicaDSN)},
			Policy:   dbresolver.RandomPolicy{},
		}).
			SetMaxOpenConns(d.MaxOpenConns).
			SetMaxIdleConns(d.MaxIdleConns).
			SetConnMaxLifetime(lifetime).
			SetConnMaxIdleTime(idleTime)
		if err := db.Use(resolver); err != nil {
			logger.Errorf("dbresolver registration error: %v", err)
			return nil, err
		}
	}

	sqlDB, err := db.DB()
	if err != nil {
		logger.Errorf("database sql.DB error: %v", err)
		return nil, err
	}
	sqlDB.SetMaxOpenConns(d.MaxOpenConns)
	sqlDB.SetMaxIdleConns(d.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(lifetime)
	sqlDB.SetConnMaxIdleTime(idleTime)

	return db, nil
}
