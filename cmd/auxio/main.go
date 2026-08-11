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
	"path/filepath"
	"syscall"
	"time"

	"github.com/protob/auxio/internal/config"
	"github.com/protob/auxio/internal/dashboard"
	"github.com/protob/auxio/internal/imaging"
	"github.com/protob/auxio/internal/s3"
	"github.com/protob/auxio/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var (
	version = "dev"
	commit  = "unknown"
)

var (
	showVersion  = flag.Bool("version", false, "Print version information and exit")
	rebuildIndex = flag.Bool("rebuild-index", false, "Rebuild SQLite index from sidecar files and exit")
)

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Printf("Auxio %s (commit: %s)\n", version, commit)
		os.Exit(0)
	}

	if err := run(); err != nil {
		slog.Error("Auxio exited with error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// A deployed binary built without -tags release serves a notice page at
	// /dashboard, so say so at startup rather than waiting for a browser.
	if !dashboard.Embedded {
		slog.Warn("Dashboard is not embedded — this binary was built without -tags release")
	}

	if cfg.UsingDefaultKeys {
		slog.Warn("Using default access/secret keys — set AUXIO_ACCESS_KEY and AUXIO_SECRET_KEY for production")
	}
	if cfg.UsingDefaultDashboardCreds() {
		slog.Warn("Using default dashboard credentials — set AUXIO_USERNAME and AUXIO_PASSWORD for production")
	}

	// A non-loopback bind is reachable off-machine, which is what makes default
	// credentials dangerous rather than merely noisy (the warnings above always fire).
	if !cfg.IsLoopbackBind() {
		if cfg.UsingDefaultKeys {
			slog.Error("Refusing insecure config: non-loopback bind with default S3 keys. Set AUXIO_ACCESS_KEY/AUXIO_SECRET_KEY.", "bind", cfg.Bind)
			return errors.New("default credentials on a public bind")
		}
		if cfg.UsingDefaultDashboardCreds() {
			slog.Warn("Non-loopback bind with default dashboard credentials — set AUXIO_USERNAME/AUXIO_PASSWORD", "bind", cfg.Bind)
		}
	}

	if *rebuildIndex {
		return doRebuild(cfg)
	}

	slog.Info("Starting Auxio",
		"version", version,
		"commit", commit,
		"data_dir", cfg.DataDir,
		"bind", cfg.Bind,
		"port", cfg.HttpPort,
		"region", cfg.Region,
	)

	store, err := storage.NewFilesystemStore(storage.FilesystemOptions{
		DataDir: cfg.DataDir,
		Region:  cfg.Region,
	})
	if err != nil {
		return fmt.Errorf("failed to create storage: %w", err)
	}
	defer store.Close()

	go func() {
		ticker := time.NewTicker(time.Duration(cfg.UploadCleanupHours) * time.Hour)
		defer ticker.Stop()
		if err := store.CleanupAbandonedUploads(cfg.UploadCleanupHours); err != nil {
			slog.Warn("abandoned upload cleanup", "err", err)
		}
		for range ticker.C {
			if err := store.CleanupAbandonedUploads(cfg.UploadCleanupHours); err != nil {
				slog.Warn("abandoned upload cleanup", "err", err)
			}
		}
	}()

	auth := s3.NewAuthEngine(cfg.AccessKey, cfg.SecretKey, cfg.Region)

	var imagingMiddleware func(http.Handler) http.Handler
	if cfg.ImagingEnabled {
		slog.Info("Image processing enabled")

		imaging.Startup()
		defer imaging.Shutdown()

		processor := imaging.NewProcessor(cfg.ImagingMaxWidth, cfg.ImagingMaxHeight)
		cache := imaging.NewCache(filepath.Join(cfg.DataDir, ".imgcache"))

		middleware := imaging.NewMiddleware(store, processor, cache)
		imagingMiddleware = middleware.Handler

		store.SetMutationHook(func(bucket, key string) {
			if err := cache.InvalidateForObject(bucket, key); err != nil {
				slog.Warn("Failed to invalidate image cache", "bucket", bucket, "key", key, "err", err)
			}
		})

		// The mutation hook keeps the cache correct; this sweep bounds its
		// size - without it every ?w=…&fmt=… variant accumulates forever.
		go func() {
			const cacheMaxAge = 30 * 24 * time.Hour
			ticker := time.NewTicker(24 * time.Hour)
			defer ticker.Stop()
			if err := cache.Cleanup(cacheMaxAge); err != nil {
				slog.Warn("image cache cleanup", "err", err)
			}
			for range ticker.C {
				if err := cache.Cleanup(cacheMaxAge); err != nil {
					slog.Warn("image cache cleanup", "err", err)
				}
			}
		}()
	}

	router := chi.NewRouter()

	router.Use(middleware.Recoverer)
	router.Use(s3.LoggingMiddleware)
	router.Use(middleware.StripSlashes)

	// Same shape as s3.NewRouter's Group (the tests' wiring) - keep in sync.
	router.Group(func(r chi.Router) {
		r.Use(s3.AuthMiddleware(auth, store))
		if imagingMiddleware != nil {
			r.Use(imagingMiddleware)
		}
		s3.RegisterRoutes(r, store, cfg.Region)
	})

	router.Group(func(r chi.Router) {
		r.Use(dashboard.DashboardAuthMiddleware(cfg))
		dashboard.RegisterRoutes(r, store, cfg, cfg.DataDir, version, commit)
		r.Post("/api/export/{bucket}", dashboard.ExportBucketHandler(cfg.DataDir))
		r.Post("/api/buckets/{bucket}/upload", dashboard.UploadHandlerChi(store))
	})

	dashboardFS, err := dashboard.GetStaticFS()
	if err != nil {
		return fmt.Errorf("dashboard assets: %w", err)
	}
	fileServer := http.FileServer(http.FS(dashboardFS))
	spaHandler := http.StripPrefix("/dashboard", dashboard.SPAHandler{FS: dashboardFS, Handler: fileServer})
	// StripSlashes redirects /dashboard/ → /dashboard, so handle both
	router.Handle("/dashboard", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	}))
	router.Handle("/dashboard/*", spaHandler)

	server := &http.Server{
		Addr:         cfg.Bind + ":" + cfg.HttpPort,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("HTTP server listening", "bind", cfg.Bind, "port", cfg.HttpPort)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	case sig := <-stopCh:
		slog.Info("Shutting down", "signal", sig)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown error: %w", err)
	}

	slog.Info("Server stopped")
	return nil
}

func doRebuild(cfg *config.Config) error {
	slog.Info("Rebuilding SQLite index from sidecar files", "data_dir", cfg.DataDir)

	dbPath := filepath.Join(cfg.DataDir, "index.db")
	db, err := storage.OpenDB(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	stats, err := storage.RebuildIndex(db, cfg.DataDir)
	if err != nil {
		return fmt.Errorf("failed to rebuild index: %w", err)
	}

	slog.Info("Index rebuild complete",
		"buckets", stats.Buckets,
		"objects", stats.Objects,
		"sidecars_read", stats.SidecarReads,
		"sidecars_created", stats.SidecarCreate,
		"errors", stats.Errors,
	)

	return nil
}
