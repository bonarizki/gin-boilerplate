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

var DB *gorm.DB

// DbConnection create database connection
func DbConnection(masterDSN, replicaDSN string) error {
	logMode := false
	debug := false
	if c := config.Get(); c != nil {
		logMode = c.Database.LogMode
		debug = c.Server.Debug
	}

	loglevel := gormlogger.Silent
	if logMode {
		loglevel = gormlogger.Info
	}

	db, err := gorm.Open(postgres.Open(masterDSN), &gorm.Config{
		Logger: gormlogger.Default.LogMode(loglevel),
	})
	if err != nil {
		logger.Errorf("database connection error: %v", err)
		return err
	}

	if !debug {
		if err := db.Use(dbresolver.Register(dbresolver.Config{
			Replicas: []gorm.Dialector{
				postgres.Open(replicaDSN),
			},
			Policy: dbresolver.RandomPolicy{},
		})); err != nil {
			logger.Errorf("dbresolver registration error: %v", err)
			return err
		}
	}

	sqlDB, err := db.DB()
	if err != nil {
		logger.Errorf("database sql.DB error: %v", err)
		return err
	}
	if c := config.Get(); c != nil {
		d := c.Database
		sqlDB.SetMaxOpenConns(d.MaxOpenConns)
		sqlDB.SetMaxIdleConns(d.MaxIdleConns)
		sqlDB.SetConnMaxLifetime(time.Duration(d.ConnMaxLifetime) * time.Minute)
		sqlDB.SetConnMaxIdleTime(time.Duration(d.ConnMaxIdleTime) * time.Minute)
	}

	DB = db
	return nil
}

// GetDB connection
func GetDB() *gorm.DB {
	return DB
}

// WriteDB forces queries onto the primary connection (bypasses read replicas).
// Use for auth and other consistency-sensitive reads.
func WriteDB() *gorm.DB {
	if DB == nil {
		return nil
	}
	return DB.Clauses(dbresolver.Write)
}
