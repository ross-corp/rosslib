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
ALTER TABLE books ADD COLUMN IF NOT EXISTS publisher        TEXT;
ALTER TABLE books ADD COLUMN IF NOT EXISTS page_count       INT;
ALTER TABLE books ADD COLUMN IF NOT EXISTS subjects         TEXT;

ALTER TABLE collection_items ADD COLUMN IF NOT EXISTS rating      SMALLINT;
ALTER TABLE collection_items ADD COLUMN IF NOT EXISTS review_text TEXT;
ALTER TABLE collection_items ADD COLUMN IF NOT EXISTS spoiler     BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE collection_items ADD COLUMN IF NOT EXISTS date_read   TIMESTAMPTZ;
ALTER TABLE collection_items ADD COLUMN IF NOT EXISTS date_added  TIMESTAMPTZ;

ALTER TABLE collections ADD COLUMN IF NOT EXISTS collection_type VARCHAR(20) NOT NULL DEFAULT 'shelf';

ALTER TABLE tag_keys ADD COLUMN IF NOT EXISTS mode VARCHAR(20) NOT NULL DEFAULT 'select_one';

ALTER TABLE users ADD COLUMN IF NOT EXISTS is_ghost BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_moderator BOOLEAN NOT NULL DEFAULT false;

-- ── Google OAuth support ─────────────────────────────────────────────────────
ALTER TABLE users ADD COLUMN IF NOT EXISTS google_id VARCHAR(255);
-- Make password_hash nullable for OAuth-only users.
ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL;
-- Unique index on google_id (partial, non-null only).
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_google_id ON users (google_id) WHERE google_id IS NOT NULL;

ALTER TABLE user_books ADD COLUMN IF NOT EXISTS progress_pages    INT;
ALTER TABLE user_books ADD COLUMN IF NOT EXISTS progress_percent  SMALLINT;
ALTER TABLE user_books ADD COLUMN IF NOT EXISTS device_total_pages INT;
ALTER TABLE user_books ADD COLUMN IF NOT EXISTS date_dnf TIMESTAMPTZ;

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

-- ── user_books: canonical user-book relationship ─────────────────────────────

CREATE TABLE IF NOT EXISTS user_books (
	id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
	user_id    UUID        NOT NULL REFERENCES users(id),
	book_id    UUID        NOT NULL REFERENCES books(id),
	rating     SMALLINT,
	review_text TEXT,
	spoiler    BOOLEAN     NOT NULL DEFAULT false,
	date_read  TIMESTAMPTZ,
	date_added TIMESTAMPTZ DEFAULT NOW(),
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	UNIQUE(user_id, book_id)
);

CREATE INDEX IF NOT EXISTS idx_user_books_user_id ON user_books (user_id, date_added DESC);

-- ── Data migration: copy read_status shelf books → user_books + status labels ─

DO $$
DECLARE
  v_user_id   UUID;
  v_key_id    UUID;
  v_value_id  UUID;
  v_book_id   UUID;
  v_slug      TEXT;
  v_status_slug TEXT;
  rec         RECORD;
BEGIN
  -- For each user that has books in read_status shelves but no user_books rows
  FOR v_user_id IN
    SELECT DISTINCT c.user_id
    FROM collection_items ci
    JOIN collections c ON c.id = ci.collection_id
    WHERE c.exclusive_group = 'read_status'
      AND NOT EXISTS (SELECT 1 FROM user_books ub WHERE ub.user_id = c.user_id)
  LOOP
    -- Ensure Status tag key exists for this user
    SELECT tk.id INTO v_key_id
    FROM tag_keys tk
    WHERE tk.user_id = v_user_id AND tk.slug = 'status';

    IF v_key_id IS NULL THEN
      INSERT INTO tag_keys (user_id, name, slug, mode)
      VALUES (v_user_id, 'Status', 'status', 'select_one')
      ON CONFLICT (user_id, slug) DO UPDATE SET name = tag_keys.name
      RETURNING id INTO v_key_id;

      INSERT INTO tag_values (tag_key_id, name, slug) VALUES (v_key_id, 'Want to Read', 'want-to-read') ON CONFLICT DO NOTHING;
      INSERT INTO tag_values (tag_key_id, name, slug) VALUES (v_key_id, 'Owned to Read', 'owned-to-read') ON CONFLICT DO NOTHING;
      INSERT INTO tag_values (tag_key_id, name, slug) VALUES (v_key_id, 'Currently Reading', 'currently-reading') ON CONFLICT DO NOTHING;
      INSERT INTO tag_values (tag_key_id, name, slug) VALUES (v_key_id, 'Finished', 'finished') ON CONFLICT DO NOTHING;
      INSERT INTO tag_values (tag_key_id, name, slug) VALUES (v_key_id, 'DNF', 'dnf') ON CONFLICT DO NOTHING;
    END IF;

    -- Copy each book from read_status shelves into user_books + set status label
    FOR rec IN
      SELECT ci.book_id, ci.rating, ci.review_text, ci.spoiler, ci.date_read, ci.date_added, ci.added_at, c.slug AS shelf_slug
      FROM collection_items ci
      JOIN collections c ON c.id = ci.collection_id
      WHERE c.user_id = v_user_id AND c.exclusive_group = 'read_status'
    LOOP
      -- Insert user_books row
      INSERT INTO user_books (user_id, book_id, rating, review_text, spoiler, date_read, date_added)
      VALUES (v_user_id, rec.book_id, rec.rating, rec.review_text, rec.spoiler, rec.date_read, COALESCE(rec.date_added, rec.added_at))
      ON CONFLICT (user_id, book_id) DO NOTHING;

      -- Map shelf slug to status label slug
      v_status_slug := CASE rec.shelf_slug
        WHEN 'read' THEN 'finished'
        WHEN 'currently-reading' THEN 'currently-reading'
        WHEN 'want-to-read' THEN 'want-to-read'
        WHEN 'owned-to-read' THEN 'owned-to-read'
        WHEN 'dnf' THEN 'dnf'
        ELSE rec.shelf_slug
      END;

      -- Look up the tag value ID
      SELECT tv.id INTO v_value_id
      FROM tag_values tv
      WHERE tv.tag_key_id = v_key_id AND tv.slug = v_status_slug;

      IF v_value_id IS NOT NULL THEN
        INSERT INTO book_tag_values (user_id, book_id, tag_key_id, tag_value_id)
        VALUES (v_user_id, rec.book_id, v_key_id, v_value_id)
        ON CONFLICT DO NOTHING;
      END IF;
    END LOOP;
  END LOOP;
END $$;

-- ── Backfill: create user_books rows for books that have Status labels but
--    were never in a read_status shelf (e.g. set directly via the label UI).

INSERT INTO user_books (user_id, book_id, rating, review_text, spoiler, date_read, date_added)
SELECT DISTINCT ON (btv.user_id, btv.book_id)
       btv.user_id,
       btv.book_id,
       ci.rating,
       ci.review_text,
       COALESCE(ci.spoiler, false),
       ci.date_read,
       COALESCE(ci.date_added, ci.added_at, NOW())
FROM book_tag_values btv
JOIN tag_keys tk ON tk.id = btv.tag_key_id AND tk.slug = 'status'
LEFT JOIN collection_items ci
       ON ci.book_id = btv.book_id
      AND ci.collection_id IN (SELECT id FROM collections WHERE user_id = btv.user_id)
WHERE NOT EXISTS (
  SELECT 1 FROM user_books ub
  WHERE ub.user_id = btv.user_id AND ub.book_id = btv.book_id
)
ORDER BY btv.user_id, btv.book_id, ci.rating DESC NULLS LAST
ON CONFLICT (user_id, book_id) DO NOTHING;

-- ── Community links: book-to-book connections ────────────────────────────────

CREATE TABLE IF NOT EXISTS book_links (
	id           UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
	from_book_id UUID         NOT NULL REFERENCES books(id),
	to_book_id   UUID         NOT NULL REFERENCES books(id),
	user_id      UUID         NOT NULL REFERENCES users(id),
	link_type    VARCHAR(50)  NOT NULL,
	note         TEXT,
	created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
	deleted_at   TIMESTAMPTZ,
	UNIQUE(from_book_id, to_book_id, link_type, user_id)
);

CREATE INDEX IF NOT EXISTS idx_book_links_from ON book_links (from_book_id);
CREATE INDEX IF NOT EXISTS idx_book_links_to   ON book_links (to_book_id);

CREATE TABLE IF NOT EXISTS book_link_votes (
	user_id      UUID        NOT NULL REFERENCES users(id),
	book_link_id UUID        NOT NULL REFERENCES book_links(id) ON DELETE CASCADE,
	created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	PRIMARY KEY (user_id, book_link_id)
);

-- ── Community link edit queue: proposed edits awaiting moderator review ───────

CREATE TABLE IF NOT EXISTS book_link_edits (
	id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
	book_link_id    UUID         NOT NULL REFERENCES book_links(id) ON DELETE CASCADE,
	user_id         UUID         NOT NULL REFERENCES users(id),
	proposed_type   VARCHAR(50),
	proposed_note   TEXT,
	status          VARCHAR(20)  NOT NULL DEFAULT 'pending',
	reviewer_id     UUID         REFERENCES users(id),
	reviewer_comment TEXT,
	created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
	reviewed_at     TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_book_link_edits_status  ON book_link_edits (status);
CREATE INDEX IF NOT EXISTS idx_book_link_edits_link_id ON book_link_edits (book_link_id);

-- ── Author follows: users following OL authors ──────────────────────────────

CREATE TABLE IF NOT EXISTS author_follows (
	user_id     UUID        NOT NULL REFERENCES users(id),
	author_key  VARCHAR(50) NOT NULL,
	author_name VARCHAR(500) NOT NULL DEFAULT '',
	created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	PRIMARY KEY (user_id, author_key)
);

CREATE INDEX IF NOT EXISTS idx_author_follows_author_key ON author_follows (author_key);

-- ── Author works snapshot: tracks known work count per author for new-pub detection

CREATE TABLE IF NOT EXISTS author_works_snapshot (
	author_key  VARCHAR(50) PRIMARY KEY,
	work_count  INT         NOT NULL DEFAULT 0,
	checked_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ── Notifications: per-user notifications (e.g. new publication by followed author)

CREATE TABLE IF NOT EXISTS notifications (
	id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
	user_id     UUID         NOT NULL REFERENCES users(id),
	notif_type  VARCHAR(50)  NOT NULL,
	title       TEXT         NOT NULL,
	body        TEXT,
	metadata    JSONB,
	read        BOOLEAN      NOT NULL DEFAULT false,
	created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_notifications_user_unread ON notifications (user_id, read, created_at DESC);

-- ── Book follows: users subscribing to books for activity notifications ──────

CREATE TABLE IF NOT EXISTS book_follows (
	user_id    UUID        NOT NULL REFERENCES users(id),
	book_id    UUID        NOT NULL REFERENCES books(id),
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	PRIMARY KEY (user_id, book_id)
);

CREATE INDEX IF NOT EXISTS idx_book_follows_book_id ON book_follows (book_id);

-- ── Password reset tokens ───────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS password_reset_tokens (
	id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
	user_id     UUID         NOT NULL REFERENCES users(id),
	token_hash  TEXT         NOT NULL,
	expires_at  TIMESTAMPTZ  NOT NULL,
	used        BOOLEAN      NOT NULL DEFAULT false,
	created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_user_id ON password_reset_tokens (user_id);
`

func Migrate(pool *pgxpool.Pool) error {
	_, err := pool.Exec(context.Background(), schema)
	return err
}
