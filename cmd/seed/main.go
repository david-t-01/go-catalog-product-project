package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
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
	// Load environment variables from .env file
	if err := godotenv.Load(".env"); err != nil {
		return fmt.Errorf("loading .env file: %w", err)
	}

	// Initialize database connection
	db, closeDB, err := database.New(
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
		os.Getenv("POSTGRES_PORT"),
	)
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	defer func() { _ = closeDB() }()

	dir := os.Getenv("POSTGRES_SQL_DIR")
	files, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("reading directory: %w", err)
	}

	var sqlFiles []os.DirEntry
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			sqlFiles = append(sqlFiles, file)
		}
	}
	sort.Slice(sqlFiles, func(i, j int) bool {
		return sqlFiles[i].Name() < sqlFiles[j].Name()
	})

	for _, file := range sqlFiles {
		path := filepath.Join(dir, file.Name())

		content, err := os.ReadFile(path)
		if err != nil {
			log.Printf("reading file %s failed: %v", file.Name(), err)
		}

		sql := string(content)
		if err := db.Exec(sql).Error; err != nil {
			return fmt.Errorf("executing %s: %w", file.Name(), err)
		}

		log.Printf("Executed %s successfully\n", file.Name())
	}
	return nil
}
