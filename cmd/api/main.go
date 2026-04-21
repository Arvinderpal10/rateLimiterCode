package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/arvinderpal10/ratelimiter/internal/config"
	"github.com/arvinderpal10/ratelimiter/internal/handlers"
	"github.com/arvinderpal10/ratelimiter/internal/limiter"
)

func main() {
	cfg := config.Load()
	store := limiter.NewStore(cfg)

	requestHandler := handlers.NewRequestHandler(store)
	statsHandler := handlers.NewStatsHandler(store)

	mux := http.NewServeMux()
	mux.HandleFunc("/request", limiter.WithRequestID(limiter.LoggingMiddleware(requestHandler.ServeHTTP)))
	mux.HandleFunc("/stats", limiter.WithRequestID(limiter.LoggingMiddleware(statsHandler.ServeHTTP)))

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("server listening on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("forced shutdown: %v", err)
	}
	store.Shutdown()
	log.Println("server stopped")
}
