package main

import (
	"database/sql"
	"embed"
	"fmt"
	"log"

	"github.com/dmehra2102/order-management-platform/internal/config"
)

var migrationFiles embed.FS

func main() {
	cfg := config.Load()

	db,err := sql.Open("postgres", cfg.DatabaseURL())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping: %v", err)
	}

	files,err := migrationFiles.ReadDir("../migrations")
	if err != nil {
		log.Fatalf("Unable to read migration folder: %v", err)
	}

	for _,f := range files {
		sqlBytes,err := migrationFiles.ReadFile("../migrations/" + f.Name())
		if err != nil {
			log.Fatalf("Failed reading migration file %s: %v", f.Name(), err)
		}

		fmt.Printf("-> Running migration: %s\n", f.Name())

		if _,err := db.Exec(string(sqlBytes)); err != nil {
			log.Fatalf("Migration %s failed: %v", f.Name(), err)
		}
	}

	fmt.Println("✓ All migrations executed")
}