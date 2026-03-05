package database

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(ctx context.Context) (*pgxpool.Pool, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://prism:prism@localhost:5432/prism?sslmode=disable"
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("connecting to database: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	return pool, nil
}

func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	schema := `
	CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

	CREATE TABLE IF NOT EXISTS analyses (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		name TEXT NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		owner TEXT NOT NULL,
		share_token TEXT NOT NULL DEFAULT '',
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS cards (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		analysis_id UUID NOT NULL REFERENCES analyses(id) ON DELETE CASCADE,
		card_type TEXT NOT NULL,
		label TEXT NOT NULL DEFAULT '',
		config JSONB NOT NULL DEFAULT '{}',
		position_x DOUBLE PRECISION NOT NULL DEFAULT 0,
		position_y DOUBLE PRECISION NOT NULL DEFAULT 0,
		input_card_ids UUID[] NOT NULL DEFAULT '{}',
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS dashboards (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		analysis_id UUID NOT NULL REFERENCES analyses(id) ON DELETE CASCADE,
		name TEXT NOT NULL,
		published BOOLEAN NOT NULL DEFAULT FALSE,
		layout JSONB NOT NULL DEFAULT '{}',
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS datasets (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		analysis_id UUID NOT NULL REFERENCES analyses(id) ON DELETE CASCADE,
		card_id UUID NOT NULL,
		name TEXT NOT NULL,
		data JSONB NOT NULL DEFAULT '{}',
		row_count INTEGER NOT NULL DEFAULT 0,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_cards_analysis_id ON cards(analysis_id);
	CREATE INDEX IF NOT EXISTS idx_dashboards_analysis_id ON dashboards(analysis_id);
	CREATE INDEX IF NOT EXISTS idx_datasets_analysis_id ON datasets(analysis_id);
	`

	_, err := pool.Exec(ctx, schema)
	if err != nil {
		return err
	}

	// Add share_token column if missing (migration for existing DBs)
	_, _ = pool.Exec(ctx, "ALTER TABLE analyses ADD COLUMN IF NOT EXISTS share_token TEXT NOT NULL DEFAULT ''")

	return nil
}
