package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/drug-verification/server/internal/config"
)

type MigrationFile struct {
	Version int
	Name    string
	Path    string
}

func main() {
	direction := flag.String("direction", "up", "Migration direction: up, down, or status")
	migrationsDir := flag.String("path", "db/migrations", "Path to migrations directory")
	flag.Parse()

	// Also allow positional argument: go run ./cmd/migrate up|down|status
	if flag.NArg() > 0 {
		*direction = flag.Arg(0)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	conn, err := pgx.Connect(ctx, cfg.Database.URL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	log.Println("Connected to database successfully.")

	// Ensure schema_migrations table exists
	_, err = conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version BIGINT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
	`)
	if err != nil {
		log.Fatalf("Failed to create schema_migrations table: %v", err)
	}

	switch strings.ToLower(*direction) {
	case "status":
		showStatus(ctx, conn, *migrationsDir)
	case "up":
		runUp(ctx, conn, *migrationsDir)
	case "down":
		runDown(ctx, conn, *migrationsDir)
	default:
		log.Fatalf("Unknown direction %q. Supported: up, down, status", *direction)
	}
}

func getAppliedVersions(ctx context.Context, conn *pgx.Conn) (map[int]time.Time, error) {
	rows, err := conn.Query(ctx, "SELECT version, applied_at FROM schema_migrations ORDER BY version ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[int]time.Time)
	for rows.Next() {
		var v int
		var t time.Time
		if err := rows.Scan(&v, &t); err != nil {
			return nil, err
		}
		applied[v] = t
	}
	return applied, rows.Err()
}

func loadMigrationFiles(dir string, suffix string) ([]MigrationFile, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading migrations directory %q: %w", dir, err)
	}

	var files []MigrationFile
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, suffix) {
			parts := strings.SplitN(name, "_", 2)
			if len(parts) < 2 {
				continue
			}
			ver, err := strconv.Atoi(parts[0])
			if err != nil {
				continue
			}
			files = append(files, MigrationFile{
				Version: ver,
				Name:    name,
				Path:    filepath.Join(dir, name),
			})
		}
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].Version < files[j].Version
	})
	return files, nil
}

func showStatus(ctx context.Context, conn *pgx.Conn, dir string) {
	applied, err := getAppliedVersions(ctx, conn)
	if err != nil {
		log.Fatalf("Failed to get applied migrations: %v", err)
	}

	upFiles, err := loadMigrationFiles(dir, ".up.sql")
	if err != nil {
		log.Fatalf("Failed to load migration files: %v", err)
	}

	fmt.Println("--- Migration Status ---")
	for _, file := range upFiles {
		appliedAt, ok := applied[file.Version]
		if ok {
			fmt.Printf("[APPLIED]   Version %06d: %s (at %s)\n", file.Version, file.Name, appliedAt.Format(time.RFC3339))
		} else {
			fmt.Printf("[PENDING]   Version %06d: %s\n", file.Version, file.Name)
		}
	}
}

func runUp(ctx context.Context, conn *pgx.Conn, dir string) {
	applied, err := getAppliedVersions(ctx, conn)
	if err != nil {
		log.Fatalf("Failed to get applied migrations: %v", err)
	}

	upFiles, err := loadMigrationFiles(dir, ".up.sql")
	if err != nil {
		log.Fatalf("Failed to load migration files: %v", err)
	}

	appliedCount := 0
	for _, file := range upFiles {
		if _, ok := applied[file.Version]; ok {
			log.Printf("Migration %s already applied, skipping", file.Name)
			continue
		}

		log.Printf("Applying migration: %s ...", file.Name)
		content, err := os.ReadFile(file.Path)
		if err != nil {
			log.Fatalf("Failed to read migration file %s: %v", file.Path, err)
		}

		tx, err := conn.Begin(ctx)
		if err != nil {
			log.Fatalf("Failed to begin transaction for %s: %v", file.Name, err)
		}

		if _, err := tx.Exec(ctx, string(content)); err != nil {
			_ = tx.Rollback(ctx)
			log.Fatalf("Failed to execute migration %s: %v", file.Name, err)
		}

		if _, err := tx.Exec(ctx, "INSERT INTO schema_migrations (version, applied_at) VALUES ($1, now())", file.Version); err != nil {
			_ = tx.Rollback(ctx)
			log.Fatalf("Failed to record migration %s: %v", file.Name, err)
		}

		if err := tx.Commit(ctx); err != nil {
			log.Fatalf("Failed to commit transaction for %s: %v", file.Name, err)
		}

		log.Printf("Successfully applied %s", file.Name)
		appliedCount++
	}

	if appliedCount == 0 {
		log.Println("Database is already up to date. No new migrations applied.")
	} else {
		log.Printf("Migration complete. Applied %d migration(s).", appliedCount)
	}
}

func runDown(ctx context.Context, conn *pgx.Conn, dir string) {
	applied, err := getAppliedVersions(ctx, conn)
	if err != nil {
		log.Fatalf("Failed to get applied migrations: %v", err)
	}

	downFiles, err := loadMigrationFiles(dir, ".down.sql")
	if err != nil {
		log.Fatalf("Failed to load migration files: %v", err)
	}

	// Sort descending for rollback
	sort.Slice(downFiles, func(i, j int) bool {
		return downFiles[i].Version > downFiles[j].Version
	})

	rolledBackCount := 0
	for _, file := range downFiles {
		if _, ok := applied[file.Version]; !ok {
			continue
		}

		log.Printf("Rolling back migration: %s ...", file.Name)
		content, err := os.ReadFile(file.Path)
		if err != nil {
			log.Fatalf("Failed to read migration file %s: %v", file.Path, err)
		}

		tx, err := conn.Begin(ctx)
		if err != nil {
			log.Fatalf("Failed to begin transaction for %s: %v", file.Name, err)
		}

		if _, err := tx.Exec(ctx, string(content)); err != nil {
			_ = tx.Rollback(ctx)
			log.Fatalf("Failed to execute rollback %s: %v", file.Name, err)
		}

		if _, err := tx.Exec(ctx, "DELETE FROM schema_migrations WHERE version = $1", file.Version); err != nil {
			_ = tx.Rollback(ctx)
			log.Fatalf("Failed to remove migration record %s: %v", file.Name, err)
		}

		if err := tx.Commit(ctx); err != nil {
			log.Fatalf("Failed to commit transaction for %s: %v", file.Name, err)
		}

		log.Printf("Successfully rolled back %s", file.Name)
		rolledBackCount++
		break // Only rollback the latest migration by default
	}

	if rolledBackCount == 0 {
		log.Println("No applied migrations to rollback.")
	} else {
		log.Println("Rollback complete.")
	}
}
