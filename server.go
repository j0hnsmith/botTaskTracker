package main

import (
	"context"
	"embed"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/j0hnsmith/botTaskTracker/handlers"
)

//go:embed static
var staticFS embed.FS

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	ctx := context.Background()

	server, err := handlers.NewServer(ctx)
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}
	defer func() {
		if cerr := server.Close(); cerr != nil {
			slog.Error("failed to close server", "error", cerr)
		}
	}()

	mux := server.Routes(staticFS)
	handler := handlers.LoggingMiddleware(mux)

	addr := ":7002"
	slog.Info("starting server", "addr", addr)
	log.Printf("🚀 botTaskTracker running at http://localhost%s", addr)

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
