package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/david-t-01/go-catalog-product-project/app/catalog"
	"github.com/david-t-01/go-catalog-product-project/app/database"
	"github.com/david-t-01/go-catalog-product-project/models"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}

	// signal handling for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Initialize database connection
	db, closeDB := database.New(
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
		os.Getenv("POSTGRES_PORT"),
	)
	defer func() { _ = closeDB() }()

	// Initialize handlers
	prodRepo := models.NewProductsRepository(db)
	cat := catalog.NewHandler(prodRepo)

	// Set up routing
	mux := http.NewServeMux()
	mux.HandleFunc("GET /catalog", cat.HandleGet)

	// Set up the HTTP server
	srv := &http.Server{
		Addr:              fmt.Sprintf("localhost:%s", os.Getenv("HTTP_PORT")),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Start the server
	go func() {
		log.Printf("Starting server on http://%s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %s", err)
		}

		log.Println("Server stopped gracefully")
	}()

	<-ctx.Done()
	log.Println("Shutting down server...")
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("shutdown: %v", err)
	}
	stop()
}
