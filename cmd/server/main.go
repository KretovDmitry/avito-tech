package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/KretovDmitry/avito-tech/internal/auth"
	"github.com/KretovDmitry/avito-tech/internal/banner"
	"github.com/KretovDmitry/avito-tech/internal/config"
	"github.com/KretovDmitry/avito-tech/pkg/accesslog"
	"github.com/KretovDmitry/avito-tech/pkg/log"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"
	sqldblogger "github.com/simukti/sqldb-logger"
)

// Version indicates the current version of the application.
var Version = "1.0.0"

var flagConfig = flag.String("config", "./config/local.yml", "path to the config file")

func main() {
	flag.Parse()

	// Server run context
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	// Create root logger tagged with server version
	logger := log.New().With(serverCtx, "version", Version)

	// Load application configurations
	cfg, err := config.Load(*flagConfig, logger)
	if err != nil {
		logger.Errorf("failed to load application configuration: %s", err)
		os.Exit(-1)
	}

	db, err := sql.Open("pgx", cfg.DSN)
	if err != nil {
		logger.Error(err)
		os.Exit(-1)
	}

	// Log every query to the database
	db = sqldblogger.OpenDriver(cfg.DSN, db.Driver(), logger)

	// to check connectivity and DSN correctness
	err = db.Ping()
	if err != nil {
		logger.Errorf("failed to connect to the database: %s", err)
		os.Exit(-1)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// to check connectivity and DSN correctness
	if res := rdb.Ping(serverCtx); res.Err() != nil {
		logger.Errorf("failed to connect to the redis: %s", res.Err())
		os.Exit(-1)
	}

	// close connections
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error(err)
		}
		if err := rdb.Close(); err != nil {
			logger.Error(err)
		}
	}()

	// Init repository for banner service
	repo, err := banner.NewRepository(db, rdb, logger, cfg)
	if err != nil {
		logger.Errorf("failed to create banner repository: %s", err)
		os.Exit(-1)
	}

	// Init service
	bannerService, err := banner.NewService(repo, logger, cfg)
	if err != nil {
		logger.Error("failed to init banner service")
		os.Exit(-1)
	}
	// Do not loose banners being asynchronously deleted
	defer bannerService.Stop()

	// Init repository for auth service
	authRepo, err := auth.NewRepository(db, logger)
	if err != nil {
		logger.Errorf("failed to create auth repository: %s", err)
		os.Exit(-1)
	}

	authService, err := auth.NewService(authRepo, logger, cfg)
	if err != nil {
		logger.Error("failed to init auth service")
		os.Exit(-1)
	}

	router := chi.NewRouter()
	router.Use(accesslog.Handler(logger))
	router.Use(middleware.Recoverer)
	router.Use(authService.Middleware)

	handler := banner.HandlerWithOptions(bannerService, banner.ChiServerOptions{
		BaseRouter:       router,
		ErrorHandlerFunc: banner.ErrorHandlerFunc,
	})

	// Build HTTP server
	address := fmt.Sprintf(":%v", cfg.ServerPort)
	hs := &http.Server{
		Addr:    address,
		Handler: handler,
	}

	// Graceful shutdown if not live reload dev mode is on
	if !cfg.LiveMode {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT,
			syscall.SIGTERM, syscall.SIGQUIT, os.Interrupt)
		go func() {
			<-sig

			logger.Infof("shutting down server with %s timeout", cfg.ShutdownTimeout)

			if err := hs.Shutdown(serverCtx); err != nil {
				logger.Errorf("graceful shutdown failed: %v", err)
			}
			serverStopCtx()
		}()
	}

	// Start the HTTP server with graceful shutdown
	logger.Infof("server %v is running at %v", Version, address)
	if err := hs.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error(err)
		os.Exit(-1)
	}

	// Wait for server context to be stopped
	if !cfg.LiveMode {
		select {
		case <-serverCtx.Done():
		case <-time.After(cfg.ShutdownTimeout):
			logger.Error("graceful shutdown timed out.. forcing exit")
			os.Exit(-1)
		}
	}
}
