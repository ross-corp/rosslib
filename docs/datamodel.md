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

### Book

Global catalog. Not per-user.

| Column | Type | Notes |
|---|---|---|
| id | uuid PK | |
| title | varchar(500) | |
| subtitle | varchar(500) | |
| isbn | varchar(13) | nullable |
| isbn13 | varchar(13) | nullable |
| published_year | integer | |
| publisher | varchar(255) | |
| page_count | integer | |
| language | varchar(10) | ISO 639-1 |
| description | text | |
| cover_url | text | S3 key |
| open_library_id | varchar(50) | for sync |
| created_at | timestamptz | |
| updated_at | timestamptz | |

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

### Collection

A named list owned by a user. Can be exclusive (like "Read" / "Want to Read") or non-exclusive (like "Favorites", "Space Books").

| Column | Type | Notes |
|---|---|---|
| id | uuid PK | |
| user_id | uuid FK → User | |
| name | varchar(255) | |
| slug | varchar(255) | unique per user |
| description | text | |
| is_exclusive | boolean | if true, a book can only appear in one collection of this type per user |
| exclusive_group | varchar(100) | nullable; collections in the same group enforce mutual exclusivity |
| is_public | boolean | default true |
| created_at | timestamptz | |
| updated_at | timestamptz | |

Default collections created on user registration:

- "Read" (exclusive_group: "read_status")
- "Want to Read" (exclusive_group: "read_status")
- "Currently Reading" (exclusive_group: "read_status")

---

### CollectionItem

| Column | Type | Notes |
|---|---|---|
| id | uuid PK | |
| collection_id | uuid FK → Collection | |
| book_id | uuid FK → Book | |
| added_at | timestamptz | |
| notes | text | user's private note on this item |
| sort_order | integer | manual ordering |

Unique constraint: `(collection_id, book_id)`

Application enforces mutual exclusivity within an `exclusive_group` when adding items.

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
