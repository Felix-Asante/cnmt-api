package migrate

import (
	"context"
	"database/sql"
	"fmt"

	"cnmt/db/migrations"

	"github.com/pressly/goose/v3"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Up applies all pending migrations (goose + river).
func Up(databaseURL string) error {
	db, err := open(databaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := goose.Up(db, "."); err != nil {
		return fmt.Errorf("migrate up: %w", err)
	}

	if err := riverUp(databaseURL); err != nil {
		return fmt.Errorf("river migrate up: %w", err)
	}

	return nil
}

// Status prints migration status.
func Status(databaseURL string) error {
	db, err := open(databaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := goose.Status(db, "."); err != nil {
		return fmt.Errorf("migrate status: %w", err)
	}
	return nil
}

func open(databaseURL string) (*sql.DB, error) {
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	goose.SetBaseFS(migrations.FS)
	if err := goose.SetDialect("postgres"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("set dialect: %w", err)
	}
	return db, nil
}

func riverUp(databaseURL string) error {
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return fmt.Errorf("open pgxpool for river: %w", err)
	}
	defer pool.Close()

	migrator, err := rivermigrate.New(riverpgxv5.New(pool), nil)
	if err != nil {
		return fmt.Errorf("create river migrator: %w", err)
	}

	if _, err := migrator.Migrate(ctx, rivermigrate.DirectionUp, nil); err != nil {
		return fmt.Errorf("apply river migrations: %w", err)
	}

	return nil
}
