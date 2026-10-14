package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/david-t-01/go-catalog-product-project/app/api"
	"github.com/david-t-01/go-catalog-product-project/app/apidocs"
	"github.com/david-t-01/go-catalog-product-project/app/catalog"
	"github.com/david-t-01/go-catalog-product-project/app/categories"
	"github.com/david-t-01/go-catalog-product-project/app/database"
	"github.com/david-t-01/go-catalog-product-project/models"
)

const shutdownTimeout = 10 * time.Second

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("server", "err", err)
		os.Exit(1)
	}
}

// run holds every defer so main can exit with a status without skipping them.
func run(logger *slog.Logger) error {
	// A missing .env is fine (a container gets its variables from compose).
	if err := godotenv.Load(".env"); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("loading .env: %w", err)
	}

	// Database
	db, closeDB, err := database.New(
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
		os.Getenv("POSTGRES_PORT"),
	)
	if err != nil {
		return err
	}
	defer func() { _ = closeDB() }()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Repositories
	productsRepo := models.NewProductsRepository(db)
	categoriesRepo := models.NewCategoriesRepository(db)

	// Services
	catalogSvc := catalog.NewService(productsRepo)
	categoriesSvc := categories.NewService(categoriesRepo)

	// Handlers
	catalogHandler := catalog.NewHandler(catalogSvc, logger)
	categoriesHandler := categories.NewHandler(categoriesSvc, logger)

	// Routes
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		api.OKResponse(w, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /catalog", catalogHandler.HandleList)
	mux.HandleFunc("GET /catalog/{code}", catalogHandler.HandleDetails)
	mux.HandleFunc("GET /categories", categoriesHandler.HandleList)
	mux.HandleFunc("POST /categories", categoriesHandler.HandleCreate)
	apidocs.Register(mux)

	srv := &http.Server{
		Addr:              ":" + os.Getenv("HTTP_PORT"),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("starting server", "addr", srv.Addr)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("listen: %w", err)
	case <-ctx.Done():
	}
	logger.Info("shutting down server")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	return nil
}
