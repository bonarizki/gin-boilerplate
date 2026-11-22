package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "time/tzdata"

	"github.com/bonarizki/gin-boilerplate/internal/adapters/database"
	"github.com/bonarizki/gin-boilerplate/internal/adapters/database/migrations"
	"github.com/bonarizki/gin-boilerplate/internal/adapters/database/seeders"
	"github.com/bonarizki/gin-boilerplate/internal/app/routers"
	"github.com/bonarizki/gin-boilerplate/internal/app/services"
	"github.com/bonarizki/gin-boilerplate/internal/domain/repositories"
	"github.com/bonarizki/gin-boilerplate/pkg/config"
	"github.com/bonarizki/gin-boilerplate/pkg/logger"
	"github.com/bonarizki/gin-boilerplate/pkg/metrics"
)

func main() {
	if err := config.SetupConfig(); err != nil {
		logger.Fatalf("config SetupConfig() error: %s", err)
	}

	cfg := config.Get()
	if cfg == nil {
		logger.Fatalf("config not loaded")
	}
	if _, err := time.LoadLocation(cfg.Server.Timezone); err != nil {
		logger.Fatalf("invalid SERVER_TIMEZONE %q: %v", cfg.Server.Timezone, err)
	}

	// Initialize metrics tracking
	metrics.Init()
	logger.Infof("Metrics tracking initialized")

	masterDSN, replicaDSN := config.DbConfiguration()

	if err := database.DbConnection(masterDSN, replicaDSN); err != nil {
		logger.Fatalf("database DbConnection error: %s", err)
	}
	if err := migrations.Migrate(); err != nil {
		logger.Fatalf("database migration failed: %v", err)
	}

	if config.IsDevelopment() && config.GetBool("SEED_DEMO_DATA") {
		if err := seeders.Run(); err != nil {
			logger.Fatalf("database seeding failed: %v", err)
		}
	}

	userRepo := repositories.NewUserRepository()
	refreshTokenRepo := repositories.NewRefreshTokenRepository()
	services.StartTokenCleanup(userRepo, refreshTokenRepo)

	router := routers.SetupRoute()

	addr := config.ServerConfig()
	requestTimeout := time.Duration(cfg.Server.RequestTimeoutSeconds) * time.Second
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  requestTimeout,
		WriteTimeout: requestTimeout,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("server listen error: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	logger.Infof("shutdown signal received, shutting down gracefully")

	shutdownTimeout := time.Duration(cfg.Server.ShutdownTimeoutSeconds) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Errorf("server shutdown error: %v", err)
	}

	if sqlDB, err := database.GetDB().DB(); err == nil && sqlDB != nil {
		if err := sqlDB.Close(); err != nil {
			logger.Errorf("database close error: %v", err)
		}
	}

	logger.Infof("shutdown complete")
}
