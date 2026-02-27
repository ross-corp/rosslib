# Data Model

Schema is applied idempotently at API startup via `db.Migrate` in `api/internal/db/schema.go`. New columns use `ALTER TABLE ... ADD COLUMN IF NOT EXISTS`. There is no migration tool.

---

## Implemented tables

### `users`

| Column | Type | Notes |
|---|---|---|
| id | uuid PK | `gen_random_uuid()` |
| username | text | URL-safe; added via migration `1700000002_add_username.go` (not a built-in PocketBase auth field) |
| email | varchar(255) unique | |
| password_hash | text | nullable; bcrypt hash (null for Google OAuth-only accounts) |
| google_id | text | nullable; Google user ID for OAuth accounts; set when user signs in with Google |
| display_name | varchar(100) | nullable |
| bio | text | nullable |
| avatar_url | text | nullable; S3 key |
| is_private | boolean | default false |
| is_moderator | boolean | default false; grants moderation privileges (e.g. deleting community links); managed via admin UI (`/admin`) |
| author_key | varchar(50) | nullable; Open Library author ID (e.g. `OL23919A`); links user account to their author page; shows "Author" badge on profile; managed via admin UI |
| created_at | timestamptz | |
| deleted_at | timestamptz | soft delete |

### `follows`

Asymmetric social graph.

| Column | Type | Notes |
|---|---|---|
| follower_id | uuid FK → users | |
| followee_id | uuid FK → users | |
| status | varchar(20) | `'active'` or `'pending'`; private accounts create follows with `'pending'` status requiring approval |
| created_at | timestamptz | |

PK: `(follower_id, followee_id)`

### `books`

Global catalog. Not per-user. Records are upserted by `open_library_id` when a user first adds a book to any shelf.

| Column | Type | Notes |
|---|---|---|
| id | uuid PK | |
| open_library_id | varchar(50) unique | bare OL work ID e.g. `OL82592W` (no `/works/` prefix) |
| title | varchar(500) | |
| cover_url | text | nullable; Open Library cover URL |
| isbn13 | varchar(13) | nullable |
| authors | text | nullable; comma-separated author names |
| publication_year | integer | nullable; first publish year from OL |
| publisher | text | nullable; from OL editions API |
| page_count | integer | nullable; from OL editions API |
| subjects | text | nullable; comma-separated subjects from OL (up to 10) |
| created_at | timestamptz | |

### `collections`

A named list owned by a user. Covers default shelves, custom shelves, and tag collections.

| Column | Type | Notes |
|---|---|---|
| id | uuid PK | |
| user_id | uuid FK → users | |
| name | varchar(255) | |
| slug | varchar(255) | unique per user |
| is_exclusive | boolean | default false |
| exclusive_group | varchar(100) | nullable; shelves in the same group enforce mutual exclusivity |
| is_public | boolean | default true |
| collection_type | varchar(20) | `'shelf'` (default) or `'tag'` |
| created_at | timestamptz | |

Unique constraint: `(user_id, slug)`

**Default shelves** — created on registration (or lazily on first `/me/shelves` call):

| name | slug | exclusive_group |
|---|---|---|
| Want to Read | `want-to-read` | `read_status` |
| Currently Reading | `currently-reading` | `read_status` |
| Read | `read` | `read_status` |

All three have `is_exclusive = true`. Adding a book to any of them removes it from the other two for that user.

**Default tag** — created idempotently (via `ensureDefaultFavorites`) when shelves are listed:

| name | slug | collection_type | is_exclusive |
|---|---|---|---|
| Favorites | `favorites` | `tag` | false |

**collection_type values:**
- `'shelf'` — a bookshelf (default or custom). Shown in the shelf sidebar and profile shelf cards.
- `'tag'` — a path-based tag. Slug may contain `/` for hierarchy (e.g. `scifi/dystopian`). Shown as tag chips on the profile page. See `docs/organization.md`.

### `user_books`

Per-user book ownership. Replaces `collection_items` for user-book metadata (rating, review, dates). Status is tracked via `book_tag_values` (the Status label key), not via shelf placement.

| Column | Type | Notes |
|---|---|---|
| id | uuid PK | `gen_random_uuid()` |
| user_id | uuid FK → users | |
| book_id | uuid FK → books | |
| rating | smallint | nullable; 1–5 |
| review_text | text | nullable |
| spoiler | boolean | default false |
| date_read | timestamptz | nullable; when the user finished the book |
| date_dnf | timestamptz | nullable; when the user stopped reading (DNF) |
| date_added | timestamptz | default now(); original add date (preserves Goodreads history on import) |
| progress_pages | integer | nullable; current page number |
| progress_percent | smallint | nullable; 0–100 reading percentage |
| device_total_pages | integer | nullable; user's edition page count (overrides `books.page_count` for % calc) |
| selected_edition_key | text | nullable; Open Library edition key (e.g. `OL123M`); when set, the frontend displays this edition's cover |
| selected_edition_cover_url | text | nullable; cached cover URL for the selected edition; avoids extra API calls |
| created_at | timestamptz | |

Unique constraint: `(user_id, book_id)`
Index: `(user_id, date_added DESC)`

Rating and review are updated via `PATCH /me/books/:olId`. Absent fields in the PATCH body are ignored — only explicitly provided fields are updated. Edition selection is also updated via PATCH with `selected_edition_key` and `selected_edition_cover_url` — when set, SQL queries use `COALESCE(ub.selected_edition_cover_url, b.cover_url)` so the edition cover takes precedence.

### `collection_items` (legacy)

Still used for tag collections (Favorites, custom tags). Old read_status shelf data was migrated to `user_books` + status labels at startup.

| Column | Type | Notes |
|---|---|---|
| id | uuid PK | |
| collection_id | uuid FK → collections | |
| book_id | uuid FK → books | |
| added_at | timestamptz | server-set timestamp |
| rating | smallint | nullable; 1–5 |
| review_text | text | nullable |
| spoiler | boolean | default false |
| date_read | timestamptz | nullable |
| date_added | timestamptz | nullable |

Unique constraint: `(collection_id, book_id)`

### `tag_keys`

Label categories, owned per-user. E.g. "Gifted from", "Read in".

| Column | Type | Notes |
|---|---|---|
| id | uuid PK | |
| user_id | uuid FK → users | |
| name | varchar(100) | |
| slug | varchar(100) | |
| mode | varchar(20) | `'select_one'` or `'select_multiple'` |
| created_at | timestamptz | |

Unique: `(user_id, slug)`

### `tag_values`

Predefined options for a label category. E.g. "mom", "dad", "History/Engineering".

| Column | Type | Notes |
|---|---|---|
| id | uuid PK | |
| tag_key_id | uuid FK → tag_keys (cascade) | |
| name | varchar(100) | |
| slug | varchar(255) | Widened to support nested paths like `history/engineering` |
| created_at | timestamptz | |

Unique: `(tag_key_id, slug)`. New values can be created inline when assigning a label to a book — the `PUT /me/books/:olId/tags/:keyId` endpoint accepts `{ value_name }` to find-or-create.

Slugs may contain `/` to express nesting (e.g. `history/engineering`). Each segment is individually slugified; the separator is preserved. Querying a parent path (`history`) also returns books tagged with any descendant (`history/engineering`, `history/science/ancient`, etc.).

### `book_tag_values`

Which label values a user has assigned to a given book.

| Column | Type | Notes |
|---|---|---|
| user_id | uuid FK → users | |
| book_id | uuid FK → books | |
| tag_key_id | uuid FK → tag_keys (cascade) | |
| tag_value_id | uuid FK → tag_values (cascade) | |
| created_at | timestamptz | |

PK: `(user_id, book_id, tag_value_id)` — allows multiple rows per book per key for `select_multiple` keys. `select_one` enforcement (delete existing before insert) is done in application code, not a DB constraint.

Index: `(user_id, book_id, tag_key_id)` for efficient "all labels for this book" lookups.

### `threads`

Discussion threads on a book's page. Any logged-in user can create a thread.

| Column | Type | Notes |
|---|---|---|
| id | uuid PK | `gen_random_uuid()` |
| book_id | uuid FK → books | |
| user_id | uuid FK → users | thread author |
| title | varchar(500) | |
| body | text | |
| spoiler | boolean | default false |
| created_at | timestamptz | |
| deleted_at | timestamptz | soft delete |

Indexes: `book_id` for listing threads by book; GIN trigram index on `title` (`gin_trgm_ops`) for similar-thread lookups via `pg_trgm` `similarity()`.

### `thread_comments`

Comments on a thread. Supports one level of nesting (replies to top-level comments only).

| Column | Type | Notes |
|---|---|---|
| id | uuid PK | `gen_random_uuid()` |
| thread_id | uuid FK → threads | |
| user_id | uuid FK → users | comment author |
| parent_id | uuid FK → thread_comments | nullable; if set, this is a reply |
| body | text | |
| created_at | timestamptz | |
| deleted_at | timestamptz | soft delete |

Index: `thread_id` for listing comments by thread. Nesting constraint: if `parent_id` is set, the referenced comment must have `parent_id IS NULL` (enforced in application code).

### `activities`

Append-only event log for social feeds. Written fire-and-forget from handlers — failures never block the primary operation.

| Column | Type | Notes |
|---|---|---|
| id | uuid PK | `gen_random_uuid()` |
| user_id | uuid FK → users | the actor |
| activity_type | varchar(50) | `shelved`, `rated`, `reviewed`, `created_thread`, `followed_user`, `followed_author`, `followed_book`, `started_book`, `finished_book`, `created_link` |
| book_id | uuid FK → books | nullable |
| target_user_id | uuid FK → users | nullable; for `followed_user` |
| collection_id | uuid FK → collections | nullable; for `shelved` |
| thread_id | uuid FK → threads | nullable; for `created_thread` |
| metadata | jsonb | nullable; extra context (shelf_name, rating, review_snippet, thread_title) |
| created_at | timestamptz | |

Indexes: `(user_id, created_at DESC)`, `(created_at DESC)`.

---

## Relationships

```
users ──< follows >── users
users ──< user_books >── books
users ──< collections ──< collection_items >── books  (tag collections)
users ──< tag_keys ──< tag_values
users ──< book_tag_values >── tag_values >── tag_keys  (includes Status labels)
users ──< threads >── books
users ──< thread_comments >── threads
users ──< activities >── books, users, collections, threads
```

### `book_links`

Community-submitted book-to-book connections. Any authenticated user can suggest a link between two books already in the local catalog. Links can be soft-deleted by the original author or by any user with `is_moderator = true`.

| Column | Type | Notes |
|---|---|---|
| id | uuid PK | `gen_random_uuid()` |
| from_book_id | uuid FK → books | source book |
| to_book_id | uuid FK → books | target book |
| user_id | uuid FK → users | who submitted the link |
| link_type | varchar(50) | `sequel`, `prequel`, `companion`, `mentioned_in`, `similar`, `adaptation` |
| note | text | nullable; optional explanation of the connection |
| created_at | timestamptz | |
| deleted_at | timestamptz | soft delete |

Unique constraint: `(from_book_id, to_book_id, link_type, user_id)` — each user can submit a given link type between two books once.

Indexes: `from_book_id`, `to_book_id`.

### `book_link_votes`

Upvotes on community links. Sorted by vote count on book pages.

| Column | Type | Notes |
|---|---|---|
| user_id | uuid FK → users | |
| book_link_id | uuid FK → book_links (cascade) | |
| created_at | timestamptz | |

PK: `(user_id, book_link_id)` — one vote per user per link.

### `book_link_edits`

Proposed edits to community links. Any authenticated user can propose a change to a link's type or note. Moderators approve or reject edits from the admin panel. Approved edits are applied to the link immediately.

| Column | Type | Notes |
|---|---|---|
| id | uuid PK | `gen_random_uuid()` |
| book_link_id | uuid FK → book_links (cascade) | the link being edited |
| user_id | uuid FK → users | who proposed the edit |
| proposed_type | varchar(50) | nullable; new link_type if changed |
| proposed_note | text | nullable; new note if changed |
| status | varchar(20) | `'pending'`, `'approved'`, or `'rejected'` |
| reviewer_id | uuid FK → users | nullable; moderator who reviewed |
| reviewer_comment | text | nullable; optional reviewer note |
| created_at | timestamptz | |
| reviewed_at | timestamptz | nullable; when reviewed |

Indexes: `status`, `book_link_id`.

At least one of `proposed_type` or `proposed_note` must be non-null (enforced in application code). One pending edit per user per link (enforced in application code).

### `author_follows`

Users following Open Library authors. Authors are not stored locally — only the OL author key and a cached name are persisted.

| Column | Type | Notes |
|---|---|---|
| user_id | uuid FK → users | |
| author_key | varchar(50) | Open Library author ID (e.g. `OL23919A`) |
| author_name | varchar(500) | cached display name; default `''` |
| created_at | timestamptz | |

PK: `(user_id, author_key)`
Index: `author_key`

### `author_works_snapshot`

Tracks the last-known work count for each followed author. Used by the background poller to detect new publications. Seeded on first poll (no notification generated); subsequent polls compare against the snapshot.

| Column | Type | Notes |
|---|---|---|
| author_key | varchar(50) PK | Open Library author ID |
| work_count | integer | last-known number of works on OL |
| checked_at | timestamptz | when the snapshot was last refreshed |

### `book_follows`

Users following specific books. Followers receive notifications when new threads, community links, or reviews are posted on the book.

| Column | Type | Notes |
|---|---|---|
| user_id | uuid FK → users | |
| book_id | uuid FK → books | |
| created_at | timestamptz | |

PK: `(user_id, book_id)`
Index: `book_id`

### `notifications`

Per-user notification rows. Types include `new_publication` (author works poller), `book_new_thread`, `book_new_link`, and `book_new_review` (book follow notifications). The schema is generic for future notification types.

| Column | Type | Notes |
|---|---|---|
| id | uuid PK | `gen_random_uuid()` |
| user_id | uuid FK → users | recipient |
| notif_type | varchar(50) | e.g. `new_publication` |
| title | text | short title (e.g. "New work by Author") |
| body | text | nullable; longer description |
| metadata | jsonb | nullable; extra context (author_key, author_name, new_count, new_titles) |
| read | boolean | default false |
| created_at | timestamptz | |

Index: `(user_id, read, created_at DESC)` for efficient unread count and listing.

### `notification_preferences`

Per-user notification preferences. Each boolean field controls whether that notification type is sent. If no row exists for a user, all types default to enabled.

| Column | Type | Notes |
|---|---|---|
| id | uuid PK | `gen_random_uuid()` |
| user | uuid FK → users (cascade) | unique; one row per user |
| new_publication | bool | default true; author new work notifications |
| book_new_thread | bool | default true; new thread on followed book |
| book_new_link | bool | default true; new link on followed book |
| book_new_review | bool | default true; new review on followed book |
| review_liked | bool | default true; someone liked your review |
| thread_mention | bool | default true; @mentioned in a comment |
| book_recommendation | bool | default true; someone recommended a book |
| created | timestamptz | PocketBase auto-generated |

Index: unique on `user`.

---

## Relationships

```
users ──< follows >── users
users ──< user_books >── books
users ──< collections ──< collection_items >── books  (tag collections)
users ──< tag_keys ──< tag_values
users ──< book_tag_values >── tag_values >── tag_keys  (includes Status labels)
users ──< threads >── books
users ──< thread_comments >── threads
users ──< activities >── books, users, collections, threads
users ──< book_links >── books  (from/to)
users ──< book_link_votes >── book_links
users ──< book_link_edits >── book_links  (proposed edits)
users ──< author_follows     (OL author keys)
users ──< book_follows >── books  (book subscriptions)
users ──< notifications      (per-user notifications)
users ──< notification_preferences  (per-user notification settings)
users ──< password_reset_tokens  (password reset tokens)
users ──< genre_ratings >── books  (per-user genre dimension scores)
author_works_snapshot        (OL author key → work count snapshot)
collections ──< computed_collections  (operation definition for live lists)
books ──< book_stats               (precomputed aggregate stats)
users ──< pending_imports          (unmatched import rows)
users ──< reports                  (content reports, reviewer)

```

### `genre_ratings`

Per-user genre dimension scores on books. Users rate how strongly a book fits each genre on a 0–10 scale. Aggregate averages are shown on book detail pages.

| Column | Type | Notes |
|---|---|---|
| id | uuid PK | `gen_random_uuid()` |
| user_id | uuid FK → users | |
| book_id | uuid FK → books | |
| genre | varchar(100) | genre name (from predefined list) |
| rating | smallint | 0–10; CHECK constraint enforced |
| created_at | timestamptz | |
| updated_at | timestamptz | |

Unique constraint: `(user_id, book_id, genre)` — one rating per user per book per genre.
Index: `book_id` for efficient aggregate queries.

Allowed genres (same as the predefined genre list): Fiction, Non-fiction, Fantasy, Science fiction, Mystery, Romance, Horror, Thriller, Biography, History, Poetry, Children.

### `computed_collections`

Stores the operation definition for computed (live) lists. When a user saves a set operation result as a continuous list, this table records which collections and operation were used. On each view, the operation is re-executed against current data.

| Column | Type | Notes |
|---|---|---|
| id | uuid PK | `gen_random_uuid()` |
| collection_id | uuid FK → collections | unique; the result shelf |
| operation | varchar(50) | `union`, `intersection`, or `difference` |
| source_collection_a | uuid | source collection A (always the user's own) |
| source_collection_b | uuid | nullable; source collection B (for same-user operations) |
| source_username_b | varchar(40) | nullable; other user's username (for cross-user operations) |
| source_slug_b | varchar(255) | nullable; other user's collection slug (for cross-user operations) |
| is_continuous | boolean | default false; if true, list auto-refreshes on each view |
| last_computed_at | timestamptz | updated each time the list is dynamically recomputed |
| created_at | timestamptz | |

Index: `collection_id`

For same-user operations, `source_collection_a` and `source_collection_b` are both set. For cross-user operations, `source_collection_a` is the user's own collection and the other user's collection is identified by `source_username_b` + `source_slug_b` (resolved on each view, respecting privacy).

### `password_reset_tokens`

Tokens for the forgot-password flow. Tokens are stored as SHA-256 hashes (not raw values). Single-use, expire after 1 hour. Previous unused tokens for a user are invalidated when a new one is requested.

| Column | Type | Notes |
|---|---|---|
| id | uuid PK | `gen_random_uuid()` |
| user_id | uuid FK → users | |
| token_hash | text | SHA-256 hash of the raw token sent in the email |
| expires_at | timestamptz | 1 hour after creation |
| used | boolean | default false; set to true after successful reset |
| created_at | timestamptz | |

Index: `user_id`

### `book_stats`

Precomputed aggregate stats per book. Avoids expensive multi-join COUNT/AVG queries on hot paths (book detail page, etc.). Stats are refreshed asynchronously via fire-and-forget goroutines whenever a user changes a book's status, rating, or review. Backfilled for all books at API startup.

| Column | Type | Notes |
|---|---|---|
| book_id | uuid PK FK → books | one row per book |
| reads_count | int | users with "finished" status; default 0 |
| want_to_read_count | int | users with "want-to-read" status; default 0 |
| rating_sum | bigint | sum of all user ratings (1–5); default 0 |
| rating_count | int | number of users who rated; default 0 |
| review_count | int | number of users with non-empty review_text; default 0 |
| updated_at | timestamptz | last refresh time |

Average rating is computed as `rating_sum / rating_count` at query time.

API: `GET /books/:workId/stats` returns all cached stats. `GET /books/:workId` reads `reads_count` and `want_to_read_count` from this table instead of running the expensive aggregate query.

### `feedback`

User-submitted bug reports and feature requests. Moderators can toggle status via the admin panel.

| Column | Type | Notes |
|---|---|---|
| id | uuid PK | `gen_random_uuid()` |
| user | uuid FK → users (cascade) | submitter |
| type | select | `bug` or `feature` |
| title | text | required |
| description | text | required |
| steps_to_reproduce | text | nullable; bug reports only |
| severity | select | nullable; `low`, `medium`, or `high`; bug reports only |
| status | select | `open` or `closed`; default `open` |
| created | timestamptz | PocketBase auto-generated |

Indexes: `user`, `status`.

### `reports`

User-submitted content reports. Moderators can review or dismiss reports via the admin panel. A unique constraint on `(reporter, content_type, content_id)` prevents duplicate reports from the same user on the same content.

| Column | Type | Notes |
|---|---|---|
| id | uuid PK | `gen_random_uuid()` |
| reporter | uuid FK → users (cascade) | who submitted the report |
| content_type | select | `review`, `thread`, `comment`, or `link` |
| content_id | text | required; ID of the reported content (e.g. user_books ID, thread ID, etc.) |
| reason | select | `spam`, `harassment`, `inappropriate`, or `other` |
| details | text | nullable; additional context from the reporter |
| status | select | `pending`, `reviewed`, or `dismissed`; default `pending` |
| reviewer | uuid FK → users | nullable; moderator who reviewed |
| created | timestamptz | PocketBase auto-generated |

Indexes: `status`, unique on `(reporter, content_type, content_id)`.

### `pending_imports`

Unmatched rows from Goodreads CSV imports. Saved so users can retry or manually resolve them later.

| Column | Type | Notes |
|---|---|---|
| id | uuid PK | `gen_random_uuid()` |
| user | uuid FK → users (cascade) | |
| source | text | `'goodreads'` |
| title | text | required |
| author | text | nullable |
| isbn13 | text | nullable |
| exclusive_shelf | text | nullable; Goodreads shelf name |
| custom_shelves | json | array of shelf names |
| rating | number | nullable; 1–5 |
| review_text | text | nullable |
| date_read | text | nullable |
| date_added | text | nullable |
| status | select | `'unmatched'` or `'resolved'` |
| created | timestamptz | auto |

Index: `(user, status)` for efficient listing of unresolved imports.

---

## Planned tables (not yet in schema.go)

These are designed but not built. Do not reference them in code until they exist.

### `authors` / `book_authors`

Author records and the book↔author join table. Currently authors are stored as a plain `TEXT` comma-separated string on `books.authors`.

### `reviews`

Standalone review table. The current model keeps rating and review on `user_books` (one per user per book), which already enforces uniqueness. A dedicated reviews table may be useful if reviews gain features like upvotes, replies, or rich formatting.

### `book_edits`

Wiki-style metadata correction queue.
