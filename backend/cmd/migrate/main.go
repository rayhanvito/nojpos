// Command migrate menjalankan migrasi SQL berversi (goose) untuk NOJPOS.
//
// Penggunaan:
//
//	migrate up      — terapkan semua migrasi tertunda
//	migrate status  — tampilkan status tiap migrasi
//	migrate down    — rollback satu migrasi terakhir
//	migrate version — tampilkan versi skema saat ini
//
// DATABASE_URL dibaca dari environment (kontrak sama seperti aplikasi).
package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/rayhanvito/nojpos/backend/migrations"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "migrate:", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("penggunaan: %s <up|status|down|version>", filepath.Base(os.Args[0]))
	}
	cmd := os.Args[1]
	switch cmd {
	case "up", "status", "down", "version":
	default:
		return fmt.Errorf("perintah tidak dikenal %q (pakai up|status|down|version)", cmd)
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return fmt.Errorf("DATABASE_URL kosong: set DATABASE_URL ke DSN Postgres terlebih dahulu")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("membuka database: %w", err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}

	provider, err := goose.NewProvider(goose.DialectPostgres, db, migrations.FS)
	if err != nil {
		return fmt.Errorf("membuat provider goose: %w", err)
	}

	switch cmd {
	case "up":
		results, err := provider.Up(ctx)
		if err != nil {
			return fmt.Errorf("up: %w", err)
		}
		if len(results) == 0 {
			fmt.Println("up: tidak ada migrasi tertunda (no pending migrations)")
			return nil
		}
		for _, r := range results {
			fmt.Println(r.String())
		}
	case "status":
		statuses, err := provider.Status(ctx)
		if err != nil {
			return fmt.Errorf("status: %w", err)
		}
		for _, s := range statuses {
			state := string(s.State)
			applied := ""
			if s.State == goose.StateApplied {
				applied = " " + s.AppliedAt.Format("2006-01-02 15:04:05")
			}
			fmt.Printf("%-8s %s%s\n", state, filepath.Base(s.Source.Path), applied)
		}
	case "down":
		result, err := provider.Down(ctx)
		if err != nil {
			return fmt.Errorf("down: %w", err)
		}
		if result == nil {
			fmt.Println("down: tidak ada migrasi untuk di-rollback")
			return nil
		}
		fmt.Println(result.String())
	case "version":
		current, target, err := provider.GetVersions(ctx)
		if err != nil {
			return fmt.Errorf("version: %w", err)
		}
		fmt.Printf("version: current=%d target=%d\n", current, target)
	}
	return nil
}
