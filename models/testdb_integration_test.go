//go:build integration

package models

import (
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/david-t-01/go-catalog-product-project/app/database"
)

func testDB(t *testing.T) *gorm.DB {
	t.Helper()
	_ = godotenv.Load("../.env")
	db, closeDB, err := database.New(
		os.Getenv("POSTGRES_HOST"), os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"), os.Getenv("POSTGRES_PORT"),
	)
	if err != nil {
		t.Skipf("database unavailable: %v", err)
	}
	t.Cleanup(func() { _ = closeDB() })
	return db
}

func dec(s string) *decimal.Decimal {
	d := decimal.RequireFromString(s)
	return &d
}
