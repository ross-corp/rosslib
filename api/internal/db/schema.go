package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

const schema = `
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS users (
	id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
	username      VARCHAR(40) UNIQUE NOT NULL,
	email         VARCHAR(255) UNIQUE NOT NULL,
	password_hash TEXT        NOT NULL,
	display_name  VARCHAR(100),
	bio           TEXT,
	avatar_url    TEXT,
	is_private    BOOLEAN     NOT NULL DEFAULT false,
	created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	deleted_at    TIMESTAMPTZ
);
`

func Migrate(pool *pgxpool.Pool) error {
	_, err := pool.Exec(context.Background(), schema)
	return err
}
