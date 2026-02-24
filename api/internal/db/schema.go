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
	isbn13          VARCHAR(13),
	authors         TEXT,
	publication_year INT,
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
	collection_type VARCHAR(20)  NOT NULL DEFAULT 'shelf',
	created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
	UNIQUE(user_id, slug)
);

CREATE TABLE IF NOT EXISTS collection_items (
	id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
	collection_id UUID        NOT NULL REFERENCES collections(id),
	book_id       UUID        NOT NULL REFERENCES books(id),
	added_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	rating        SMALLINT,
	review_text   TEXT,
	spoiler       BOOLEAN     NOT NULL DEFAULT false,
	date_read     TIMESTAMPTZ,
	date_added    TIMESTAMPTZ,
	UNIQUE(collection_id, book_id)
);

CREATE TABLE IF NOT EXISTS tag_keys (
	id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
	user_id    UUID         NOT NULL REFERENCES users(id),
	name       VARCHAR(100) NOT NULL,
	slug       VARCHAR(100) NOT NULL,
	mode       VARCHAR(20)  NOT NULL DEFAULT 'select_one',
	created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
	UNIQUE(user_id, slug)
);

CREATE TABLE IF NOT EXISTS tag_values (
	id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
	tag_key_id UUID         NOT NULL REFERENCES tag_keys(id) ON DELETE CASCADE,
	name       VARCHAR(100) NOT NULL,
	slug       VARCHAR(100) NOT NULL,
	created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
	UNIQUE(tag_key_id, slug)
);

-- PK is (user_id, book_id, tag_value_id) so multi-select keys can have
-- multiple rows per book. select_one enforcement is done in application code.
CREATE TABLE IF NOT EXISTS book_tag_values (
	user_id      UUID        NOT NULL REFERENCES users(id),
	book_id      UUID        NOT NULL REFERENCES books(id),
	tag_key_id   UUID        NOT NULL REFERENCES tag_keys(id) ON DELETE CASCADE,
	tag_value_id UUID        NOT NULL REFERENCES tag_values(id) ON DELETE CASCADE,
	created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	PRIMARY KEY (user_id, book_id, tag_value_id)
);

CREATE INDEX IF NOT EXISTS idx_book_tag_values_key
	ON book_tag_values (user_id, book_id, tag_key_id);

-- ── Idempotent migrations for existing deployments ────────────────────────────

ALTER TABLE books ADD COLUMN IF NOT EXISTS isbn13           VARCHAR(13);
ALTER TABLE books ADD COLUMN IF NOT EXISTS authors          TEXT;
ALTER TABLE books ADD COLUMN IF NOT EXISTS publication_year INT;

ALTER TABLE collection_items ADD COLUMN IF NOT EXISTS rating      SMALLINT;
ALTER TABLE collection_items ADD COLUMN IF NOT EXISTS review_text TEXT;
ALTER TABLE collection_items ADD COLUMN IF NOT EXISTS spoiler     BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE collection_items ADD COLUMN IF NOT EXISTS date_read   TIMESTAMPTZ;
ALTER TABLE collection_items ADD COLUMN IF NOT EXISTS date_added  TIMESTAMPTZ;

ALTER TABLE collections ADD COLUMN IF NOT EXISTS collection_type VARCHAR(20) NOT NULL DEFAULT 'shelf';

ALTER TABLE tag_keys ADD COLUMN IF NOT EXISTS mode VARCHAR(20) NOT NULL DEFAULT 'select_one';

ALTER TABLE users ADD COLUMN IF NOT EXISTS is_ghost BOOLEAN NOT NULL DEFAULT false;

-- Widen tag_values.slug to support nested paths like "history/engineering".
ALTER TABLE tag_values ALTER COLUMN slug TYPE VARCHAR(255);

CREATE TABLE IF NOT EXISTS threads (
	id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
	book_id     UUID         NOT NULL REFERENCES books(id),
	user_id     UUID         NOT NULL REFERENCES users(id),
	title       VARCHAR(500) NOT NULL,
	body        TEXT         NOT NULL,
	spoiler     BOOLEAN      NOT NULL DEFAULT false,
	created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
	deleted_at  TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_threads_book_id ON threads (book_id);

CREATE TABLE IF NOT EXISTS thread_comments (
	id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
	thread_id   UUID        NOT NULL REFERENCES threads(id),
	user_id     UUID        NOT NULL REFERENCES users(id),
	parent_id   UUID        REFERENCES thread_comments(id),
	body        TEXT        NOT NULL,
	created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	deleted_at  TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_thread_comments_thread_id ON thread_comments (thread_id);

CREATE TABLE IF NOT EXISTS activities (
	id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
	user_id         UUID         NOT NULL REFERENCES users(id),
	activity_type   VARCHAR(50)  NOT NULL,
	book_id         UUID         REFERENCES books(id),
	target_user_id  UUID         REFERENCES users(id),
	collection_id   UUID         REFERENCES collections(id),
	thread_id       UUID         REFERENCES threads(id),
	metadata        JSONB,
	created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_activities_user_id    ON activities (user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_activities_created_at ON activities (created_at DESC);

-- Migrate book_tag_values PK from (user_id, book_id, tag_key_id) to
-- (user_id, book_id, tag_value_id) on deployments that have the old schema.
-- The check prevents this from running on fresh installs or after it has
-- already been applied.
DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.key_column_usage
    WHERE table_name        = 'book_tag_values'
      AND constraint_name   = 'book_tag_values_pkey'
      AND column_name       = 'tag_key_id'
  ) THEN
    ALTER TABLE book_tag_values DROP CONSTRAINT book_tag_values_pkey;
    ALTER TABLE book_tag_values ADD PRIMARY KEY (user_id, book_id, tag_value_id);
  END IF;
END $$;
`

func Migrate(pool *pgxpool.Pool) error {
	_, err := pool.Exec(context.Background(), schema)
	return err
}
