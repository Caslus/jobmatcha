package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/caslus/jobmatcha/internal/api"
	"github.com/caslus/jobmatcha/internal/app"
)

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "data/app.db"
	}

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8181"
	}

	application, err := app.New(context.Background(), app.Options{
		DBPath: dbPath, StaticDir: findStaticDir(), CookieSecure: api.CookieSecureFromEnv(os.Getenv("COOKIE_SECURE")),
	})
	if err != nil {
		slog.Error("failed to initialize application", "error", err)
		os.Exit(1)
	}
	if staticDir := findStaticDir(); staticDir != "" {
		slog.Info("serving static frontend", "dir", staticDir)
	} else {
		slog.Warn("static frontend not found; serving API only", "hint", "build web or set STATIC_DIR")
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: application.Router,
	}

	// Start server in background
	go func() {
		slog.Info("listening", "port", port, "db", dbPath)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down...")

	// Stop the scheduler first (no new scans)
	application.Stop()

	// Then shut down the HTTP server with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
	}
	if err := application.Close(); err != nil {
		slog.Error("failed to close application", "error", err)
	}

	slog.Info("server stopped")
}

func findStaticDir() string {
	if configured := os.Getenv("STATIC_DIR"); configured != "" {
		return configured
	}

	for _, candidate := range []string{"web/dist/client", "../web/dist/client"} {
		if _, err := os.Stat(filepath.Join(candidate, "index.html")); err == nil {
			return candidate
		}
	}

	return ""
}
