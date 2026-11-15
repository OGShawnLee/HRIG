package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"hrig/server/internal/config"
	"hrig/server/internal/db"
	serverhttp "hrig/server/internal/http"
	"hrig/server/internal/ingestion"
	"hrig/server/internal/query"
	"hrig/server/internal/realtime"
)

func main() {
	cfg := config.Load()
	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL (or HRIG_DATABASE_URL) is required")
	}

	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to create database pool: %v", err)
	}
	defer pool.Close()

	pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
	err = pool.Ping(pingCtx)
	pingCancel()
	if err != nil {
		pool.Close()
		log.Fatalf("failed to ping database: %v", err)
	}

	hub := realtime.NewHub(cfg.StaleAfter, cfg.MissionEpochUTC)
	defer hub.Close()

	hydrateCtx, hydrateCancel := context.WithTimeout(context.Background(), 10*time.Second)
	err = hub.HydrateFromDB(hydrateCtx, db.New(pool))
	hydrateCancel()
	if err != nil {
		pool.Close()
		log.Fatalf("failed to hydrate realtime hub: %v", err)
	}

	ingester := ingestion.NewService(pool)
	ingester.SetPublisher(hub)
	reader := query.NewService(pool, cfg.MissionEpochUTC)

	router := serverhttp.NewRouterWithDeps(cfg, serverhttp.RouterDeps{
		Ingester: ingester,
		Realtime: hub,
		DBPinger: pool,
		Reader:   reader,
	})

	srv := &http.Server{
		Addr:              cfg.APIAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go hub.RunStaleMonitor(ctx, time.Second)

	errCh := make(chan error, 1)
	go func() {
		log.Printf("hrig server listening on %s", cfg.APIAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		if err != nil {
			hub.Close()
			pool.Close()
			log.Fatalf("server error: %v", err)
		}
	case <-ctx.Done():
		log.Println("shutting down server...")
		hub.Close()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			pool.Close()
			log.Fatalf("graceful shutdown failed: %v", err)
		}
		log.Println("server stopped")
	}
}
