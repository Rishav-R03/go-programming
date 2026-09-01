package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"ratelimiter/internal/analytics"
	"ratelimiter/internal/config"
	"ratelimiter/internal/db"
	"ratelimiter/internal/middleware"
	"ratelimiter/internal/ratelimiter"
	"syscall"
	"time"
)

func main() {
	cfg := config.Load()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pgPool, err := db.InitPGXPool(ctx, cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("postgres init failed: %v", err)
	}
	defer pgPool.Close()
	log.Println("connected to postgres")

	rdb, err := db.InitRedis(ctx, cfg.RedisAddr)
	if err != nil {
		log.Fatalf("redis init failed: %v", err)
	}
	defer rdb.Close()
	log.Println("connected to redis")

	script, err := db.LoadLuaScript("scripts/sliding_window.lua")
	if err != nil {
		log.Fatalf("loading lua script failed: %v", err)
	}
	limiter := ratelimiter.NewLimiter(rdb, script)
	policies := ratelimiter.NewPolicyStore()
	log.Println("rate limiter ready")

	collector := analytics.NewCollector(10_000, 4, pgPool)
	collector.Start(ctx)
	log.Println("metrics collector started (4 workers)")

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	target := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]bool{"allowed": true})
	})
	rateLimited := middleware.RateLimit(limiter, policies, collector)(target)
	mux.Handle("/check", rateLimited)

	server := &http.Server{
		Addr:    ":" + cfg.HTTPport,
		Handler: mux,
	}

	go func() {
		log.Println("listening on: %s", cfg.HTTPport)
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("http server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}

	// Stop accepting new metrics and drain whatever's still buffered
	// before we close the DB pool, so no in-flight metrics are lost.
	collector.Stop()
	log.Println("metrics collector drained")

	log.Println("shutdown complete")
}
