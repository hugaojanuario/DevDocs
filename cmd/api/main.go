package main

import (
	"context"
	"github.com/hugojanuario/devdocs/internal/api"
	"github.com/hugojanuario/devdocs/internal/bookmarks"
	"github.com/hugojanuario/devdocs/internal/config"
	"github.com/hugojanuario/devdocs/internal/database"
	"github.com/hugojanuario/devdocs/internal/search"
	"github.com/hugojanuario/devdocs/internal/security"
	"github.com/hugojanuario/devdocs/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	if err := run(); err != nil {
		slog.Error("startup failed", "reason", err.Error())
		os.Exit(1)
	}
}
func run() error {
	c, err := config.Load()
	if err != nil {
		return err
	}
	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		return database.Migrate(c.DatabaseURL, c.Migrations)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	pc, err := pgxpool.ParseConfig(c.DatabaseURL)
	if err != nil {
		return generic("invalid database config")
	}
	pc.MaxConns = 10
	pc.ConnConfig.ConnectTimeout = 5 * time.Second
	pool, err := pgxpool.NewWithConfig(ctx, pc)
	if err != nil {
		return generic("database unavailable")
	}
	defer pool.Close()
	if err = pool.Ping(ctx); err != nil {
		return generic("database unavailable")
	}
	store := bookmarks.New(pool)
	if err = store.Ready(ctx); err != nil {
		return generic("run migrations first")
	}
	fetcher := security.NewFetcher()
	official := &search.Official{Fetcher: fetcher}
	var provider search.SearchProvider = official
	if c.SearchProvider == "brave" {
		provider = &search.Brave{Key: c.BraveKey, Fetcher: fetcher, Fallback: official}
	}
	engine := service.New(provider, fetcher, c.CacheTTL, c.CacheEntries)
	server := &http.Server{Addr: c.Addr, Handler: api.New(c, store, engine), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 60 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 16 << 10}
	done := make(chan error, 1)
	go func() { slog.Info("api listening", "addr", c.Addr); done <- server.ListenAndServe() }()
	select {
	case err := <-done:
		if err != http.ErrServerClosed {
			return generic("HTTP server failed")
		}
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return server.Shutdown(shutdown)
	}
	return nil
}

type generic string

func (e generic) Error() string { return string(e) }
