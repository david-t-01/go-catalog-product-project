package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"

	"github.com/david-t-01/go-catalog-product-project/app/database"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	if err := godotenv.Load(".env"); err != nil {
		return fmt.Errorf("loading .env file: %w", err)
	}

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

	dir := os.Getenv("POSTGRES_SQL_DIR")
	entries, err := os.ReadDir(dir) // sorted by file name, so 000-truncate runs first
	if err != nil {
		return fmt.Errorf("reading directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", entry.Name(), err)
		}

		sql := string(content)
		if err := db.Exec(sql).Error; err != nil {
			return fmt.Errorf("executing %s: %w", entry.Name(), err)
		}

		log.Printf("Executed %s successfully\n", entry.Name())
	}
	return nil
}
