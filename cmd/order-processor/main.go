package main

import (
	"context"
	"database/sql"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dmehra2102/order-management-platform/internal/config"
	"github.com/dmehra2102/order-management-platform/internal/kafka"
	"github.com/dmehra2102/order-management-platform/internal/logger"
	"github.com/dmehra2102/order-management-platform/internal/metrics"
	"github.com/dmehra2102/order-management-platform/internal/repository"
)

func main() {
	cfg := config.Load()
	l := logger.New(cfg.LogLevel)

	l.Info("Starting Order Processor Service", map[string]any{
		"environment": cfg.Environment,
	})

	// Connect to DB
	db, err := sql.Open("postgres", cfg.DatabaseURL())
	if err != nil {
		l.Error("Failed to connect to database", map[string]any{
			"error": err,
		})
		os.Exit(1)
	}
	defer db.Close()

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		l.Error("Failed to ping database", map[string]any{
			"error": err,
		})
		os.Exit(1)
	}

	l.Info("Database connection successful", nil)

	orderRepo := repository.NewOrderRepository(db)

	m := metrics.New()
	if err := m.Register(); err != nil {
		l.Warn("Failed to register metrics", map[string]any{
			"error": err,
		})
	}

	consumer := kafka.NewConsumer(cfg.KafkaBrokers, "order-processor-group", l, orderRepo, db)
	defer consumer.Close()

	ctx, cancel = context.WithCancel(context.Background())

	// Starting kafka Consumer in goroutine
	consumerErrors := make(chan error, 1)
	go func() {
		consumerErrors <- consumer.Start(ctx)
	}()

	// Shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigChan:
		l.Info("Shutting down Order Processor Service", nil)
		cancel()
		<-consumerErrors
	case err := <-consumerErrors:
		if err != nil && err != context.Canceled {
			l.Error("Consumer error", map[string]any{
				"error": err,
			})
		}
	}

	l.Info("Order Processor Service shutdown complete", nil)
}
