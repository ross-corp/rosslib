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

CREATE TABLE IF NOT EXISTS follows (
	follower_id UUID        NOT NULL REFERENCES users(id),
	followee_id UUID        NOT NULL REFERENCES users(id),
	status      VARCHAR(20) NOT NULL DEFAULT 'active',
	created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	PRIMARY KEY (follower_id, followee_id)
);

CREATE TABLE IF NOT EXISTS books (
	id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
	open_library_id VARCHAR(50)  UNIQUE NOT NULL,
	title           VARCHAR(500) NOT NULL,
	cover_url       TEXT,
	created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS collections (
	id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
	user_id         UUID         NOT NULL REFERENCES users(id),
	name            VARCHAR(255) NOT NULL,
	slug            VARCHAR(255) NOT NULL,
	is_exclusive    BOOLEAN      NOT NULL DEFAULT false,
	exclusive_group VARCHAR(100),
	is_public       BOOLEAN      NOT NULL DEFAULT true,
	created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
	UNIQUE(user_id, slug)
);

CREATE TABLE IF NOT EXISTS collection_items (
	id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
	collection_id UUID        NOT NULL REFERENCES collections(id),
	book_id       UUID        NOT NULL REFERENCES books(id),
	added_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	UNIQUE(collection_id, book_id)
);
`

func Migrate(pool *pgxpool.Pool) error {
	_, err := pool.Exec(context.Background(), schema)
	return err
}
