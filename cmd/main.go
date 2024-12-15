package main

import (
	"context"
	"errors"
	"github.com/labstack/echo/v4"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"peer-drop/internal/server"
	"peer-drop/pkg/utils"
	"syscall"
	"time"
)

func main() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Initialize Echo
	e := server.InitServer()
	go startServer(e)
	go utils.RemoveInactivePeers()

	<-stop
	slog.Info("Received shutdown signal, initiating graceful shutdown...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		slog.Error("Error shutting down the server", "error", err)
	} else {
		slog.Info("Server shutdown complete")
	}
}

func startServer(e *echo.Echo) {
	// Start server
	if err := e.Start(":8080"); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("failed to start server", "error", err)
	}
}
