package main

import (
	"context"
	"log/slog"
	"notifier/internal/config"
	"notifier/internal/eventlistener"
	"notifier/internal/infrastructure/brokers"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config load failed", "err", err)
		os.Exit(1)
	}

	consumer, err := brokers.NewDefaultKafkaConsumer(cfg.Brokers, cfg.GroupID)
	if err != nil {
		slog.Error("cannot init KafkaConsumer", "err", err)
		os.Exit(1)
	}

	handler := eventlistener.NewEventHandler()
	listener := eventlistener.NewDefaultEventListener(consumer, []string{cfg.Topic}, handler)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := listener.Listen(ctx); err != nil {
		slog.Error("error listening for events", "err", err)
		os.Exit(1)
	}
	slog.Info("EventListener started", "topic", cfg.Topic, "group", cfg.GroupID)
	<-ctx.Done()
	slog.Info("shutdown signal received, stopping listener…")
	listener.Stop()
	slog.Info("EventListener stopped")
}
