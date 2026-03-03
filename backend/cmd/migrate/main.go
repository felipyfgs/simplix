package main

import (
	"database/sql"
	"flag"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/joho/godotenv"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func main() {
	_ = godotenv.Load()

	command := flag.String("cmd", "up", "goose command: up, down, status, create")
	name := flag.String("name", "", "migration name (for create)")
	flag.Parse()

	if flag.NArg() > 0 {
		*command = flag.Arg(0)
	}
	if flag.NArg() > 1 {
		*name = flag.Arg(1)
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://simplix:simplix@localhost:5432/simplix"
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	_, filename, _, _ := runtime.Caller(0)
	migrationsDir := filepath.Join(filepath.Dir(filename), "..", "..", "migrations")

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("set dialect: %v", err)
	}

	switch *command {
	case "up":
		if err := goose.Up(db, migrationsDir); err != nil {
			log.Fatalf("migrate up: %v", err)
		}
	case "down":
		if err := goose.Down(db, migrationsDir); err != nil {
			log.Fatalf("migrate down: %v", err)
		}
	case "status":
		if err := goose.Status(db, migrationsDir); err != nil {
			log.Fatalf("migrate status: %v", err)
		}
	case "create":
		if *name == "" {
			log.Fatal("name required for create")
		}
		if err := goose.Create(db, migrationsDir, *name, "sql"); err != nil {
			log.Fatalf("create migration: %v", err)
		}
	default:
		log.Fatalf("unknown command: %s", *command)
	}
	log.Printf("[migrate] %s done", *command)
}
