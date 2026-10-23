package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"mercatus/platform/config"
	"mercatus/platform/database"
	"mercatus/platform/observability"
	"mercatus/schema"
	"mercatus/seed"
)

var version = "dev"

const usage = `mercatus - multi-seller software marketplace in one binary

usage:
  mercatus serve [--only=auth,catalog,...]   run the http server
  mercatus migrate                           apply database migrations
  mercatus seed                              insert idempotent demo data
  mercatus healthcheck                       probe the local server (for docker healthchecks)
  mercatus version                           print the version

configuration is read from environment variables, see README.md
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	observability.SetupLogger(cfg.LogLevel, cfg.LogFormat)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	switch os.Args[1] {
	case "serve":
		fs := flag.NewFlagSet("serve", flag.ExitOnError)
		only := fs.String("only", os.Getenv("ONLY"), "comma separated modules to mount")
		_ = fs.Parse(os.Args[2:])
		for _, m := range strings.Split(*only, ",") {
			if m = strings.TrimSpace(m); m != "" {
				cfg.Only = append(cfg.Only, m)
			}
		}
		err = serve(ctx, cfg)
	case "migrate":
		err = migrate(ctx, cfg)
	case "seed":
		err = runSeed(ctx, cfg)
	case "healthcheck":
		err = pingHTTP(ctx, "http://127.0.0.1:"+cfg.Port+"/healthz")
	case "version":
		fmt.Println(version)
	case "help", "-h", "--help":
		fmt.Print(usage)
	default:
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
	if err != nil {
		slog.Error("command failed", "command", os.Args[1], "error", err)
		os.Exit(1)
	}
}

func openDB(ctx context.Context, cfg *config.Config) (*database.Options, error) {
	return &database.Options{
		MaxOpenConns:    cfg.DBMaxOpenConns,
		MaxIdleConns:    cfg.DBMaxIdleConns,
		ConnMaxLifetime: cfg.DBConnMaxLifetime,
	}, cfg.RequireDatabase()
}

func migrate(ctx context.Context, cfg *config.Config) error {
	opts, err := openDB(ctx, cfg)
	if err != nil {
		return err
	}
	db, err := database.Open(ctx, cfg.DatabaseURL, *opts)
	if err != nil {
		return err
	}
	defer db.Close()
	migrations, err := database.LoadMigrations(schema.Files)
	if err != nil {
		return err
	}
	n, err := database.Migrate(ctx, db, migrations)
	if err != nil {
		return err
	}
	slog.Info("migrations complete", "applied", n, "total", len(migrations))
	return nil
}

func runSeed(ctx context.Context, cfg *config.Config) error {
	opts, err := openDB(ctx, cfg)
	if err != nil {
		return err
	}
	db, err := database.Open(ctx, cfg.DatabaseURL, *opts)
	if err != nil {
		return err
	}
	defer db.Close()
	migrations, err := database.LoadMigrations(schema.Files)
	if err != nil {
		return err
	}
	if pending, err := database.PendingMigrations(ctx, db, migrations); err != nil {
		return err
	} else if len(pending) > 0 {
		return fmt.Errorf("database has %d pending migrations, run `mercatus migrate` first", len(pending))
	}
	return seed.Run(ctx, db, seed.Options{Password: os.Getenv("SEED_PASSWORD")})
}

func serve(ctx context.Context, cfg *config.Config) error {
	if err := cfg.RequireServe(); err != nil {
		return err
	}
	opts, _ := openDB(ctx, cfg)
	db, err := database.Open(ctx, cfg.DatabaseURL, *opts)
	if err != nil {
		return err
	}
	defer db.Close()

	migrations, err := database.LoadMigrations(schema.Files)
	if err != nil {
		return err
	}
	pending, err := database.PendingMigrations(ctx, db, migrations)
	if err != nil {
		return err
	}
	if len(pending) > 0 {
		return fmt.Errorf("database schema is behind (pending: %s), run `mercatus migrate` first", strings.Join(pending, ", "))
	}

	app, err := NewApp(cfg, db)
	if err != nil {
		return err
	}
	handler, err := app.Router()
	if err != nil {
		return err
	}

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
		MaxHeaderBytes:    1 << 16,
	}

	var metricsSrv *http.Server
	if cfg.MetricsAddr != "" && cfg.MetricsAddr != "off" {
		mux := http.NewServeMux()
		mux.Handle("/metrics", app.metrics.Handler())
		metricsSrv = &http.Server{Addr: cfg.MetricsAddr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
		go func() {
			if err := metricsSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				slog.Error("metrics server failed", "error", err)
			}
		}()
	}

	workersCtx, stopWorkers := context.WithCancel(context.Background())
	defer stopWorkers()
	app.StartWorkers(workersCtx)

	errCh := make(chan error, 1)
	go func() {
		slog.Info("mercatus listening", "addr", srv.Addr, "version", version, "modules", cfg.Only)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()
	app.ready.Store(true)

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	slog.Info("shutdown signal received, draining", "delay", cfg.ShutdownDelay.String())
	app.ready.Store(false)
	time.Sleep(cfg.ShutdownDelay)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown incomplete", "error", err)
	}
	stopWorkers()
	if metricsSrv != nil {
		_ = metricsSrv.Shutdown(shutdownCtx)
	}
	slog.Info("shutdown complete")
	return nil
}
