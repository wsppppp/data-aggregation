package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wsppppp/data-aggregation/internal/config"
	"github.com/wsppppp/data-aggregation/internal/repository/postgres"
	"github.com/wsppppp/data-aggregation/internal/service"
	"github.com/wsppppp/data-aggregation/internal/transport/rest"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// 1. БД
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	dbPool, err := postgres.NewClient(ctx, cfg.DB)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	// 2. Инициализация слоев
	repo := postgres.NewSubscriptionRepository(dbPool)
	svc := service.NewSubscriptionService(repo)
	handler := rest.NewHandler(svc)

	// 3. Запуск HTTP сервера
	srv := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: handler.InitRoutes(),
	}

	// Запускаем сервер в горутине, чтобы не блокировать main
	go func() {
		logger.Info("Starting HTTP server", "port", cfg.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("failed to start server", "error", err)
		}
	}()

	// Graceful Shutdown
	<-ctx.Done()
	logger.Info("Shutting down server...")

	// Даем 5 секунд на завершение текущих запросов
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	logger.Info("Server exited properly")
}
