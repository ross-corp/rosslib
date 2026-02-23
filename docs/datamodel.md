# Data Model

## Entities

### User

| Column | Type | Notes |
|---|---|---|
| id | uuid PK | |
| username | varchar(40) unique | URL-safe, lowercase |
| email | varchar(255) unique | |
| password_hash | text | bcrypt |
| display_name | varchar(100) | |
| bio | text | |
| avatar_url | text | S3 key |
| is_private | boolean | default false; private profiles require follow approval |
| created_at | timestamptz | |
| deleted_at | timestamptz | soft delete |

---

### Author

| Column | Type | Notes |
|---|---|---|
| id | uuid PK | |
| name | varchar(255) | |
| bio | text | |
| born_year | integer | |
| died_year | integer | nullable |
| open_library_id | varchar(50) | |

**book_authors** (join table)

| Column | Type |
|---|---|
| book_id | uuid FK → Book |
| author_id | uuid FK → Author |
| role | varchar(50) | "author", "editor", "translator", etc. |

---

### Genre / Tag

| Column | Type | Notes |
|---|---|---|
| id | uuid PK | |
| name | varchar(100) unique | |
| slug | varchar(100) unique | |

**book_genres** (join table): `book_id`, `genre_id`

---

### Book

Global catalog. Not per-user. Records are upserted by `open_library_id` when a user first adds a book to any shelf.

| Column | Type | Notes |
|---|---|---|
| id | uuid PK | |
| open_library_id | varchar(50) unique | bare OL work ID e.g. `OL82592W` (no `/works/` prefix) |
| title | varchar(500) | |
| cover_url | text | nullable; Open Library cover URL |
| isbn13 | varchar(13) | nullable; populated from OL lookup or import |
| authors | text | nullable; comma-separated author names |
| publication_year | integer | nullable; first publish year from OL |
| created_at | timestamptz | |

> Subtitle, publisher, page_count, and per-edition ISBNs are planned but not yet stored.

---

### Collection

A named list owned by a user. Can be exclusive (like "Read" / "Want to Read") or non-exclusive (like "Favorites", "Space Books").

| Column | Type | Notes |
|---|---|---|
| id | uuid PK | |
| user_id | uuid FK → User | |
| name | varchar(255) | |
| slug | varchar(255) | unique per user |
| is_exclusive | boolean | default false |
| exclusive_group | varchar(100) | nullable; collections in the same group enforce mutual exclusivity |
| is_public | boolean | default true |
| created_at | timestamptz | |

Default collections created on user registration (or lazily on first `/me/shelves` call):

- "Want to Read" (slug: `want-to-read`, exclusive_group: `read_status`)
- "Currently Reading" (slug: `currently-reading`, exclusive_group: `read_status`)
- "Read" (slug: `read`, exclusive_group: `read_status`)

---

### CollectionItem

| Column | Type | Notes |
|---|---|---|
| id | uuid PK | |
| collection_id | uuid FK → Collection | |
| book_id | uuid FK → Book | |
| added_at | timestamptz | |
| rating | smallint | nullable; 1–5 stars |
| review_text | text | nullable |
| spoiler | boolean | default false |
| date_read | timestamptz | nullable; when the user finished the book |
| date_added | timestamptz | nullable; original Goodreads date_added (preserves shelf history) |

Unique constraint: `(collection_id, book_id)`

Application enforces mutual exclusivity within an `exclusive_group` when adding items: adding a book to any shelf in the group removes it from all other shelves in that group for that user.

Rating/review are updated via `PATCH /shelves/:shelfId/books/:olId`. Absent fields in the JSON body are ignored (not set to null); only explicitly provided fields are updated.

---

### Review

| Column | Type | Notes |
|---|---|---|
| id | uuid PK | |
| user_id | uuid FK → User | |
| book_id | uuid FK → Book | |
| rating | smallint | 1-5, nullable |
| body | text | nullable; review text |
| contains_spoilers | boolean | default false |
| created_at | timestamptz | |
| updated_at | timestamptz | |
| deleted_at | timestamptz | |

Unique constraint: `(user_id, book_id)` — one review per user per book.

---

### Follow

Asymmetric social graph: a user follows another user.

| Column | Type | Notes |
|---|---|---|
| follower_id | uuid FK → User | |
| followee_id | uuid FK → User | |
| created_at | timestamptz | |
| status | varchar(20) | "active" or "pending" (for private accounts) |

Primary key: `(follower_id, followee_id)`

---

### Thread

Discussion thread attached to a book.

| Column | Type | Notes |
|---|---|---|
| id | uuid PK | |
| book_id | uuid FK → Book | |
| user_id | uuid FK → User | author |
| title | varchar(500) | |
| body | text | |
| contains_spoilers | boolean | |
| created_at | timestamptz | |
| updated_at | timestamptz | |
| deleted_at | timestamptz | |

---

### Comment

Reply within a thread. Supports one level of nesting (reply to comment).

| Column | Type | Notes |
|---|---|---|
| id | uuid PK | |
| thread_id | uuid FK → Thread | |
| user_id | uuid FK → User | |
| parent_id | uuid FK → Comment | nullable; for replies |
| body | text | |
| created_at | timestamptz | |
| deleted_at | timestamptz | |

---

### Link

Community-submitted connections between books (e.g. "this is a sequel to", "mentioned in", "similar vibe").

| Column | Type | Notes |
|---|---|---|
| id | uuid PK | |
| from_book_id | uuid FK → Book | |
| to_book_id | uuid FK → Book | |
| link_type | varchar(50) | "sequel", "prequel", "companion", "mentioned_in", "similar", "adaptation" |
| submitted_by | uuid FK → User | |
| note | text | optional context |
| upvotes | integer | default 0 |
| created_at | timestamptz | |
| deleted_at | timestamptz | |

---

### Activity

Append-only log for building user feeds.

| Column | Type | Notes |
|---|---|---|
| id | uuid PK | |
| user_id | uuid FK → User | actor |
| activity_type | varchar(50) | see below |
| target_id | uuid | polymorphic: book_id, review_id, collection_id, etc. |
| target_type | varchar(50) | "book", "review", "collection", "thread" |
| created_at | timestamptz | |

`activity_type` values: `added_to_collection`, `wrote_review`, `started_reading`, `finished_reading`, `created_thread`, `submitted_link`, `followed_user`

---

### BookEdit

Wiki-style edit queue for user corrections to book metadata.

| Column | Type | Notes |
|---|---|---|
| id | uuid PK | |
| book_id | uuid FK → Book | |
| user_id | uuid FK → User | |
| field | varchar(100) | which field was changed |
| old_value | text | |
| new_value | text | |
| status | varchar(20) | "pending", "approved", "rejected" |
| reviewed_by | uuid FK → User | nullable; moderator |
| created_at | timestamptz | |

---

## Relationships Summary

```
User ──< Follow >── User
User ──< CollectionItem >── Collection ──< Book
User ──< Review >── Book
User ──< Thread >── Book
Thread ──< Comment
Book ──< Link >── Book
Book ──< book_authors >── Author
Book ──< book_genres >── Genre
User ──< Activity
Book ──< BookEdit ──> User
```

## Aggregate Stats (computed)

Stored as a separate `book_stats` table (updated via background job or triggers) to avoid expensive COUNT queries on hot paths:

| Column | Type |
|---|---|
| book_id | uuid PK FK → Book |
| read_count | integer |
| want_to_read_count | integer |
| rating_count | integer |
| rating_avg | numeric(3,2) |
| review_count | integer |
| updated_at | timestamptz |
