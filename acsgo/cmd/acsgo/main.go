// ACSGO — High-performance distributed TR-069 ACS server in Go.
//
// Startup sequence:
//  1. Load configuration from environment
//  2. Connect to PostgreSQL, Redis, RabbitMQ
//  3. Start CWMP/TR-069 server (port 7547)
//  4. Start REST API + WebSocket server (port 7548)
//  5. Start background scheduler
//  6. Block until SIGINT/SIGTERM
package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/DWISSNET/acsgo/api/ws"
	apihttp "github.com/DWISSNET/acsgo/api/http"
	"github.com/DWISSNET/acsgo/pkg/cache"
	"github.com/DWISSNET/acsgo/pkg/config"
	"github.com/DWISSNET/acsgo/pkg/repository"
	"github.com/DWISSNET/acsgo/services/analytics"
	"github.com/DWISSNET/acsgo/services/automation"
	"github.com/DWISSNET/acsgo/services/cwmp"
	cwmpSession "github.com/DWISSNET/acsgo/services/cwmp/session"
	"github.com/DWISSNET/acsgo/services/metrics"
	"github.com/DWISSNET/acsgo/services/queue"
	"github.com/DWISSNET/acsgo/services/scheduler"
)

func main() {
	cfg := config.Load()
	logger := buildLogger(cfg.LogLevel)
	defer logger.Sync() //nolint:errcheck

	logger.Info("Starting ACSGO",
		zap.String("env", cfg.Env),
		zap.Int("cwmp_port", cfg.CWMPPort),
		zap.Int("web_port", cfg.WebPort),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ── PostgreSQL ─────────────────────────────────────────────────────────
	db, err := connectPostgres(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("PostgreSQL connection failed", zap.Error(err))
	}
	defer db.Close()
	logger.Info("PostgreSQL connected")

	// ── Redis ──────────────────────────────────────────────────────────────
	cacheClient, err := cache.NewClient(cfg.RedisURL, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		logger.Warn("Redis connection failed — running without cache", zap.Error(err))
		// Non-fatal; degrade gracefully
	} else {
		defer cacheClient.Close()
		logger.Info("Redis connected")
	}

	// ── Repositories ───────────────────────────────────────────────────────
	deviceRepo := repository.NewDeviceRepository(db)
	commandRepo := repository.NewCommandRepository(db)

	// ── Metrics ────────────────────────────────────────────────────────────
	metricsCollector := metrics.NewCollector()

	// ── Session Manager ────────────────────────────────────────────────────
	sessManager := cwmpSession.NewManager(cacheClient)

	// ── RabbitMQ queue (optional) ──────────────────────────────────────────
	var queuePub queue.CommandPublisher = &queue.NoopPublisher{}
	if cfg.RabbitMQURL != "" {
		mq, err := queue.NewClient(cfg.RabbitMQURL, commandRepo, logger)
		if err != nil {
			logger.Warn("RabbitMQ connection failed — using noop publisher", zap.Error(err))
		} else {
			defer mq.Close()
			queuePub = mq
			logger.Info("RabbitMQ connected")
		}
	}

	// ── Analytics ──────────────────────────────────────────────────────────
	analyticsEngine := analytics.NewEngine(deviceRepo, cacheClient, logger)
	_ = analyticsEngine // Used by automation

	// ── Automation ─────────────────────────────────────────────────────────
	automationEngine := automation.NewEngine(commandRepo, cacheClient, logger)
	if err := automationEngine.Subscribe(ctx); err != nil {
		logger.Warn("Automation subscribe failed", zap.Error(err))
	}

	// ── WebSocket Hub ──────────────────────────────────────────────────────
	wsHub := ws.NewHub(logger)

	// ── Background Scheduler ───────────────────────────────────────────────
	sched := scheduler.NewScheduler(deviceRepo, commandRepo, cacheClient, metricsCollector, logger)
	sched.Start()
	defer sched.Stop()

	// ── CWMP Server ────────────────────────────────────────────────────────
	cwmpServer := cwmp.NewServer(
		cfg.CWMPPort,
		deviceRepo, commandRepo,
		sessManager, cacheClient,
		queuePub, metricsCollector, logger,
		cwmp.WithBasicAuth(cfg.ACSUsername, cfg.ACSPassword),
	)

	// ── REST API + WS Server ───────────────────────────────────────────────
	apiServer := apihttp.NewServer(
		cfg.WebPort,
		deviceRepo, commandRepo,
		queuePub, wsHub,
		cfg.JWTSecret, logger,
		cfg.PrometheusEnabled,
	)

	// ── Start servers ──────────────────────────────────────────────────────
	errCh := make(chan error, 2)

	go func() {
		if err := cwmpServer.Start(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("cwmp server: %w", err)
		}
	}()

	go func() {
		if err := apiServer.Start(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("api server: %w", err)
		}
	}()

	logger.Info("ACSGO running",
		zap.String("cwmp", fmt.Sprintf("0.0.0.0:%d", cfg.CWMPPort)),
		zap.String("api", fmt.Sprintf("0.0.0.0:%d", cfg.WebPort)),
	)

	// ── Graceful shutdown ──────────────────────────────────────────────────
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		logger.Info("Shutting down", zap.String("signal", sig.String()))
	case err := <-errCh:
		logger.Error("Server error", zap.Error(err))
	}

	shutCtx, shutCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutCancel()

	if err := cwmpServer.Shutdown(shutCtx); err != nil {
		logger.Error("CWMP shutdown error", zap.Error(err))
	}
	if err := apiServer.Shutdown(shutCtx); err != nil {
		logger.Error("API shutdown error", zap.Error(err))
	}

	logger.Info("ACSGO stopped gracefully")
}

func connectPostgres(ctx context.Context, url string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("parse postgres url: %w", err)
	}
	config.MaxConns = 50
	config.MinConns = 5
	config.MaxConnLifetime = 30 * time.Minute
	config.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return pool, nil
}

func buildLogger(level string) *zap.Logger {
	lvl := zapcore.InfoLevel
	_ = lvl.UnmarshalText([]byte(level))

	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(lvl)
	cfg.EncoderConfig.TimeKey = "ts"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := cfg.Build()
	if err != nil {
		panic("build logger: " + err.Error())
	}
	return logger
}

// Keep database/sql import to avoid "imported and not used".
var _ = sql.ErrNoRows
