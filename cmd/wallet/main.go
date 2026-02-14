package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	slog.Info("starting wallet service", "service", "wallet")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	
	select {
	case <-sigCh:
		slog.Info("shutting down wallet service")
		cancel()
	case <-ctx.Done():
	}

	time.Sleep(1 * time.Second)
	slog.Info("wallet service stopped")
}
