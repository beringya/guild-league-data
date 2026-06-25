package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"nsh-guild-analytics/backend/internal/config"
	"nsh-guild-analytics/backend/internal/database"
	apphttp "nsh-guild-analytics/backend/internal/http"
	"nsh-guild-analytics/backend/internal/services"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()

	pool, err := database.OpenPostgres(ctx, cfg)
	if err != nil {
		log.Fatalf("open postgres: %v", err)
	}
	defer pool.Close()

	if err := database.Migrate(ctx, pool); err != nil {
		log.Fatalf("migrate postgres: %v", err)
	}

	redisClient := database.OpenRedis(cfg)
	defer redisClient.Close()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("connect redis: %v", err)
	}

	authService := services.NewAuthService(cfg, pool, redisClient)
	if err := authService.BootstrapAdmin(ctx); err != nil {
		log.Fatalf("bootstrap admin: %v", err)
	}

	server := apphttp.NewServer(cfg, pool, redisClient)
	httpServer := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           server.Router(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("nsh-guild-analytics listening on %s", cfg.ListenAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}
