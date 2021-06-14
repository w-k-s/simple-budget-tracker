package main

import (
	"os"

	"github.com/w-k-s/simple-budget-tracker/migrations"
)

func main() {
	migrations.MustRunMigrations("postgres", "postgres://localhost:5432/simple_budget_tracker?sslmode=disable",os.Getenv("MIGRATIONS_DIRECTORY"))
}
