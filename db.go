package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
)

var migrations = []string{
	`CREATE TABLE IF NOT EXISTS notification (
    notification_id INTEGER PRIMARY KEY ASC,
    title TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    sended_at DATETIME NULL DEFAULT NULL
);`,
}

func CreateSchema(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	sql := `CREATE TABLE IF NOT EXISTS migration (
    applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    query TEXT NOT NULL
);`

	_, err = tx.ExecContext(ctx, sql, nil)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	migrationRows, err := tx.QueryContext(ctx, "SELECT query FROM migration")
	if err != nil {
		return fmt.Errorf("failed to get migrations records from db: %w", err)
	}
	defer migrationRows.Close()

	var queries []string
	for migrationRows.Next() {
		var query string
		err = migrationRows.Scan(&query)
		if err != nil {
			return err
		}

		queries = append(queries, query)
	}

	for index, query := range queries {
		if index >= len(migrations) {
			errors.New("db scheme too new")
		}

		if query != migrations[index] {
			return fmt.Errorf("invalid db scheme. at index %d expected query: %s, got: %s",
				index, migrations[index], query)
		}
	}

	for i := len(queries); i < len(migrations); i++ {
		slog.Info("applying migration", "#", i)
		_, err = tx.ExecContext(ctx, migrations[i])
		if err != nil {
			return fmt.Errorf("failed to apply %d migration: %w", i, err)
		}

		_, err = tx.ExecContext(ctx, "INSERT INTO migration (query) VALUES (?)",
			migrations[i])
		if err != nil {
			return fmt.Errorf("failed to save executed query into migration table: %w", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func SaveUser(ctx context.Context, tx *sql.Tx, login string, hashedPassword string) error {
	return nil
}
