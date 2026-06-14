// Command gateway is the entrypoint for the LLM gateway HTTP service.
//
// M1 scope: a minimal HTTP server with a /healthz endpoint and graceful
// shutdown. Later milestones add the chat endpoint, providers, routing, etc.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Structured logger writing text to stderr. We'll thread this through the
	// app in later milestones; for now it's the process-level logger.
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	if err := run(logger); err != nil {
		logger.Error("server exited with error", "err", err)
		os.Exit(1)
	}
}

// run holds the real program logic. Keeping it separate from main lets us
// return an error (main can't) and makes the startup path testable later.
func run(logger *slog.Logger) error {
	// signal.NotifyContext gives us a context that is cancelled when the
	// process receives SIGINT (Ctrl-C) or SIGTERM (e.g. `kill`, container stop).
	// We use that cancellation as the trigger for graceful shutdown.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// ServeMux is Go's router. Since Go 1.22 the pattern can include the HTTP
	// method and path, e.g. "GET /healthz" — no third-party router needed.
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Start the server in its own goroutine so main can wait on the shutdown
	// signal. ListenAndServe blocks until the server stops; we forward any
	// non-graceful error onto a channel.
	serverErr := make(chan error, 1)
	go func() {
		logger.Info("server listening", "addr", srv.Addr)
		// After Shutdown, ListenAndServe returns http.ErrServerClosed — that's
		// the normal, expected case, so we don't treat it as an error.
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
		close(serverErr)
	}()

	// Block until either the server crashes or we get a shutdown signal.
	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
		logger.Info("shutdown signal received, draining connections")
	}

	// Give in-flight requests up to 10s to finish before forcing the close.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}
	logger.Info("server stopped cleanly")
	return nil
}
