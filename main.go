package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	_ "time/tzdata"

	"github.com/bonarizki/gin-boilerplate/internal/adapters/database"
	"github.com/bonarizki/gin-boilerplate/internal/adapters/database/migrations"
	"github.com/bonarizki/gin-boilerplate/internal/adapters/database/seeders"
	"github.com/bonarizki/gin-boilerplate/internal/app/routers"
	"github.com/bonarizki/gin-boilerplate/internal/app/services"
	"github.com/bonarizki/gin-boilerplate/internal/app/services/auth"
	"github.com/bonarizki/gin-boilerplate/internal/domain/repositories"
	"github.com/bonarizki/gin-boilerplate/pkg/config"
	"github.com/bonarizki/gin-boilerplate/pkg/logger"
	"github.com/bonarizki/gin-boilerplate/pkg/metrics"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx); err != nil {
		logger.Errorf("%v", err)
		os.Exit(1)
	}
}

// run starts the server and blocks until ctx is cancelled or the server
// fails, then shuts down gracefully.
func run(ctx context.Context) error {
	if err := config.SetupConfig(); err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	cfg := config.Get()
	if _, err := time.LoadLocation(cfg.Server.Timezone); err != nil {
		return fmt.Errorf("invalid SERVER_TIMEZONE %q: %w", cfg.Server.Timezone, err)
	}

	metrics.Init()

	masterDSN, replicaDSN := config.DbConfiguration()
	db, err := database.Connect(masterDSN, replicaDSN, cfg.Database, !cfg.Server.Debug)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get sql.DB: %w", err)
	}
	defer func() {
		if err := sqlDB.Close(); err != nil {
			logger.Errorf("database close error: %v", err)
		}
	}()

	if err := migrations.Migrate(db); err != nil {
		return fmt.Errorf("database migration failed: %w", err)
	}
	if config.IsDevelopment() && config.GetBool("SEED_DEMO_DATA") {
		if err := seeders.Run(ctx, db); err != nil {
			return fmt.Errorf("database seeding failed: %w", err)
		}
	}

	userRepo := repositories.NewUserRepository(db)
	refreshTokenRepo := repositories.NewRefreshTokenRepository(db)

	mailer, err := auth.ResolveMailer()
	if err != nil {
		return fmt.Errorf("mailer configuration: %w", err)
	}
	authService := auth.NewAuthService(userRepo, refreshTokenRepo, mailer, auth.Config{
		JWTSecret:  cfg.Server.JWTSecret,
		Issuer:     cfg.Server.JWTIssuer,
		Audience:   cfg.Server.JWTAudience,
		AccessTTL:  time.Duration(cfg.Server.AccessTokenTTLMinutes) * time.Minute,
		RefreshTTL: time.Duration(cfg.Server.RefreshTokenTTLDays) * 24 * time.Hour,
	})

	healthService := services.NewHealthService()
	healthService.AddChecker("database", services.DatabaseChecker{DB: sqlDB})

	router := routers.SetupRoute(routers.Deps{
		Auth:    authService,
		Example: services.NewExampleService(repositories.NewExampleRepository(db)),
		Health:  healthService,
	})

	bgCtx, stopBackground := context.WithCancel(ctx)
	var background sync.WaitGroup
	defer func() {
		stopBackground()
		background.Wait()
	}()
	retentionDays := config.GetInt("TOKEN_RETENTION_DAYS")
	if retentionDays <= 0 {
		retentionDays = 30
	}
	cleanup := services.TokenCleanup{
		Users:         userRepo,
		RefreshTokens: refreshTokenRepo,
		Interval:      time.Duration(config.GetInt("TOKEN_CLEANUP_INTERVAL_MINUTES")) * time.Minute,
		Retention:     time.Duration(retentionDays) * 24 * time.Hour,
	}
	background.Add(1)
	go func() {
		defer background.Done()
		cleanup.Run(bgCtx)
	}()

	requestTimeout := time.Duration(cfg.Server.RequestTimeoutSeconds) * time.Second
	server := &http.Server{
		Addr:              config.ServerConfig(),
		Handler:           router,
		ReadHeaderTimeout: requestTimeout,
		ReadTimeout:       requestTimeout,
		WriteTimeout:      requestTimeout,
	}

	serverErr := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
		close(serverErr)
	}()

	select {
	case err := <-serverErr:
		return fmt.Errorf("server listen: %w", err)
	case <-ctx.Done():
	}

	logger.Infof("shutdown signal received, shutting down gracefully")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Server.ShutdownTimeoutSeconds)*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}
	logger.Infof("shutdown complete")
	return nil
}
