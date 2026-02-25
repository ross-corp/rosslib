# API Reference

All routes are served by the Go API on `:8080`. The webapp proxies them through Next.js route handlers in `webapp/src/app/api/` so the browser never calls `:8080` directly.

**Interactive docs:** Visit [`/docs`](http://localhost:8080/docs) for a Swagger UI with the full OpenAPI 3.0 spec. The raw spec is at [`/docs/openapi.yaml`](http://localhost:8080/docs/openapi.yaml).

Auth is via a 30-day JWT in an `httpOnly` cookie named `token`. The Go API reads it via `Authorization: Bearer <token>` header — the Next.js proxy extracts the cookie and forwards it as a header.

---

## Auth

### `POST /auth/register`

```json
{ "username": "alice", "email": "alice@example.com", "password": "..." }
```

Creates a user and sets the `token` cookie. Also creates the three default shelves (Want to Read, Currently Reading, Read).

### `POST /auth/login`

```json
{ "email": "alice@example.com", "password": "..." }
```

Sets the `token` cookie on success.

---

## Books

### `GET /books/search?q=<title>[&sort=reads|rating][&year_min=N][&year_max=N]`

Searches both Meilisearch (local catalog) and Open Library concurrently. Local matches appear first, followed by external results deduplicated by work ID. Returns up to 20 results.

Optional `year_min` and `year_max` filter results by publication year. Meilisearch uses its `publication_year` filterable attribute; Open Library uses the `first_publish_year` range parameter. Books without a year are excluded when a year filter is active.

```json
{
  "total": 1234,
  "results": [
    {
      "key": "/works/OL82592W",
      "title": "The Great Gatsby",
      "authors": ["F. Scott Fitzgerald"],
      "publish_year": 1925,
      "isbn": ["9780743273565"],
      "cover_url": "https://covers.openlibrary.org/b/id/8410459-M.jpg",
      "edition_count": 120
    }
  ]
}
```

`authors`, `isbn`, and `cover_url` may be null.

### `GET /books/lookup?isbn=<isbn>`

Looks up a single book by ISBN via Open Library. Upserts the result into the local `books` table and returns it. Returns 404 if not found.

### `GET /books/:workId`

Returns a book from the local `books` table by its bare OL work ID (e.g. `OL82592W`). 404 if not in the local catalog yet.

### `GET /books/:workId/editions?limit=50&offset=0`

Returns editions for a work from Open Library. Each edition includes cover, format, publisher, page count, ISBN, and language. Paginated via `limit` and `offset`.

```json
{
  "total": 251,
  "editions": [
    {
      "key": "OL58959679M",
      "title": "The Lord of the Rings",
      "publisher": "HarperCollins",
      "publish_date": "20 December 2021",
      "page_count": 1376,
      "isbn": "9786555112511",
      "cover_url": "https://covers.openlibrary.org/b/id/15024346-M.jpg",
      "format": "Hardcover",
      "language": "por"
    }
  ]
}
```

`publisher`, `page_count`, `isbn`, and `cover_url` may be null. `format` and `language` may be empty strings when the data is unavailable. Editions are also included inline in the `GET /books/:workId` response (up to 50, with `edition_count` for the total).

### `GET /books/:workId/reviews`  *(optional auth)*

Returns all community reviews for a book. Each user appears at most once (most recent review).

When a valid token is provided, reviews from users the caller follows are sorted first (`is_followed: true`), then all reviews are ordered by `date_added` descending.

```json
[
  {
    "username": "alice",
    "display_name": "Alice",
    "avatar_url": "https://...",
    "rating": 4,
    "review_text": "Loved it.",
    "spoiler": false,
    "date_read": "2025-06-15T00:00:00Z",
    "date_dnf": null,
    "date_added": "2025-06-20T14:32:10Z",
    "is_followed": true
  }
]
```

---

## Genres

### `GET /genres`

Returns the list of predefined genres with book counts from the local catalog.

```json
[
  {
    "slug": "fiction",
    "name": "Fiction",
    "book_count": 142
  },
  {
    "slug": "science-fiction",
    "name": "Science fiction",
    "book_count": 37
  }
]
```

### `GET /genres/:slug/books?page=1&limit=20`

Returns books matching a genre, browsing the local Meilisearch index (or DB fallback) without requiring a search query. Paginated.

```json
{
  "genre": "Fiction",
  "total": 142,
  "page": 1,
  "results": [
    {
      "key": "/works/OL82592W",
      "title": "The Great Gatsby",
      "authors": ["F. Scott Fitzgerald"],
      "publish_year": 1925,
      "isbn": ["9780743273565"],
      "cover_url": "https://covers.openlibrary.org/b/id/8410459-M.jpg",
      "subjects": ["Fiction", "Classic"]
    }
  ]
}
```

Returns 404 for unknown genre slugs.

---

## Authors

### `GET /authors/search?q=<name>`

Proxies to Open Library's author search API. Returns up to 20 results.

```json
{
  "total": 41,
  "results": [
    {
      "key": "OL26320A",
      "name": "J.R.R. Tolkien",
      "birth_date": "3 January 1892",
      "death_date": "2 September 1973",
      "top_work": "The Hobbit",
      "work_count": 392,
      "top_subjects": ["Fiction", "Fantasy", "Juvenile fiction"],
      "photo_url": "https://covers.openlibrary.org/a/olid/OL26320A-M.jpg"
    }
  ]
}
```

`birth_date`, `death_date`, `top_work`, `top_subjects`, and `photo_url` may be null.

### `POST /authors/:authorKey/follow`  *(auth required)*

Follow an author. Accepts optional `{ "author_name": "..." }` to cache the display name.

### `DELETE /authors/:authorKey/follow`  *(auth required)*

Unfollow an author.

### `GET /authors/:authorKey/follow`  *(auth required)*

Check if you follow an author. Returns `{ "following": true/false }`.

### `GET /me/followed-authors`  *(auth required)*

List authors you follow.

```json
{
  "authors": [
    {
      "author_key": "OL26320A",
      "author_name": "J.R.R. Tolkien",
      "created_at": "2026-02-25T14:00:00Z"
    }
  ]
}
```

---

## Users

### `GET /users?q=<query>&page=<n>`

Search/browse users by username or display name. Alphabetical, 20 per page.

### `GET /users/:username`  *(optional auth)*

Returns a user profile. With a valid token, also returns `is_following` for the requesting user.

```json
{
  "user_id": "...",
  "username": "alice",
  "display_name": "Alice",
  "bio": "...",
  "avatar_url": null,
  "is_private": false,
  "member_since": "2024-01-01T00:00:00Z",
  "is_following": false,
  "followers_count": 10,
  "following_count": 5,
  "friends_count": 3,
  "books_read": 42
}
```

### `PATCH /users/me`  *(auth required)*

Update own display name and byline. Accepts any subset of `{ display_name, bio }`.

### `POST /me/avatar`  *(auth required)*

Upload or replace the authenticated user's profile picture. Accepts a `multipart/form-data` body with an `avatar` field containing the image file (JPEG, PNG, GIF, or WebP; max 5 MB).

Content-type is detected from the file's magic bytes — the `Content-Type` header on the file part is not trusted.

```
200 { "avatar_url": "http://localhost:9000/rosslib/avatars/<userId>.jpg" }
400 { "error": "unsupported image type: ..." }
400 { "error": "file too large (max 5 MB)" }
```

The returned URL is stored in `users.avatar_url` and returned on subsequent `GET /users/:username` calls. In production, point `MINIO_PUBLIC_URL` to the S3 bucket or CDN origin — the URL format is `{MINIO_PUBLIC_URL}/{MINIO_BUCKET}/avatars/{userId}.{ext}`.

### `GET /users/:username/reviews`

Returns all reviews (collection items with non-empty `review_text`) for a user, ordered by date added descending.

```json
[
  {
    "book_id": "...",
    "open_library_id": "OL82592W",
    "title": "The Great Gatsby",
    "cover_url": "https://...",
    "authors": "F. Scott Fitzgerald",
    "rating": 4,
    "review_text": "A timeless classic.",
    "spoiler": false,
    "date_read": "2024-06-01T00:00:00Z",
    "date_dnf": null,
    "date_added": "2024-06-02T00:00:00Z"
  }
]
```

`authors`, `cover_url`, `rating`, `date_read`, and `date_dnf` may be null.

### `POST /users/:username/follow`  *(auth required)*

Follow a user. Status is `active` immediately (private account approval not yet enforced).

### `DELETE /users/:username/follow`  *(auth required)*

Unfollow a user.

---

## Shelves (collections)

A "shelf" is a `collection` row. Default shelves (`read_status` exclusive group) enforce mutual exclusivity — adding a book to one removes it from the others in the group.

### `GET /users/:username/shelves`

Returns all shelves for a user (default + custom + tag collections).

```json
[
  {
    "id": "...",
    "name": "Read",
    "slug": "read",
    "exclusive_group": "read_status",
    "collection_type": "shelf",
    "item_count": 42
  }
]
```

`collection_type` is one of `"shelf"` or `"tag"`. See `docs/organization.md` for the distinction.

### `GET /users/:username/shelves/:slug`

Returns a shelf with its full book list.

```json
{
  "id": "...",
  "name": "Read",
  "slug": "read",
  "exclusive_group": "read_status",
  "books": [
    {
      "book_id": "...",
      "open_library_id": "OL82592W",
      "title": "The Great Gatsby",
      "cover_url": "https://...",
      "added_at": "2024-01-01T00:00:00Z",
      "rating": 4
    }
  ]
}
```

### `GET /me/shelves`  *(auth required)*

Same as `GET /users/:username/shelves` but for the authenticated user. Used by the shelf picker on book pages.

### `POST /me/shelves`  *(auth required)*

Create a custom shelf or tag collection.

```json
{
  "name": "Favorites",
  "is_exclusive": false,
  "exclusive_group": null,
  "is_public": true,
  "collection_type": "shelf"
}
```

Slug is auto-derived from `name`. Returns 409 on slug conflict.

### `PATCH /me/shelves/:id`  *(auth required)*

Rename or toggle visibility. Accepts `{ name?, is_public? }`.

### `DELETE /me/shelves/:id`  *(auth required)*

Delete a custom shelf. Returns 403 if `exclusive_group = 'read_status'` (default shelves cannot be deleted).

### `POST /shelves/:shelfId/books`  *(auth required)*

Add a book to a shelf. Upserts the book into the global `books` catalog. For exclusive shelves, removes the book from all other shelves in the same `exclusive_group` for this user.

```json
{
  "open_library_id": "OL82592W",
  "title": "The Great Gatsby",
  "cover_url": "https://..."
}
```

### `PATCH /shelves/:shelfId/books/:olId`  *(auth required)*

Update review metadata on a book in a shelf. Only provided fields are updated — absent fields are not set to null.

```json
{
  "rating": 4,
  "review_text": "Great book.",
  "spoiler": false,
  "date_read": "2024-06-01T00:00:00Z"
}
```

`rating` is 1–5 or null to clear.

### `DELETE /shelves/:shelfId/books/:olId`  *(auth required)*

Remove a book from a shelf.

---

## Tags (path-based)

Tags are `collection` rows with `collection_type = 'tag'`. Slugs can contain `/` to form a hierarchy. See `docs/organization.md` for full semantics.

### `GET /users/:username/tags/*path`

Returns books tagged with the given path or any sub-path.

```
GET /users/alice/tags/scifi           → books tagged "scifi" or "scifi/*"
GET /users/alice/tags/scifi/dystopian → books tagged "scifi/dystopian" or "scifi/dystopian/*"
```

```json
{
  "path": "scifi",
  "sub_tags": ["scifi/dystopian", "scifi/space-opera"],
  "books": [ { "book_id": "...", ... } ]
}
```

Tags are created via `POST /me/shelves` with `collection_type: "tag"`.

---

## Labels (key/value)

Labels are structured key/value annotations on books. Values can be nested using `/` (e.g. `history/engineering`). See `docs/organization.md` for full semantics.

### `GET /me/tag-keys`  *(auth required)*

Returns all label categories with their values.

```json
[
  {
    "id": "...",
    "name": "Gifted from",
    "slug": "gifted-from",
    "mode": "select_one",
    "values": [
      { "id": "...", "name": "mom", "slug": "mom" }
    ]
  }
]
```

### `POST /me/tag-keys`  *(auth required)*

Create a label category.

```json
{ "name": "Gifted from", "mode": "select_one" }
```

`mode` is `"select_one"` or `"select_multiple"`.

### `DELETE /me/tag-keys/:keyId`  *(auth required)*

Delete a label category and all its values and assignments (cascade).

### `POST /me/tag-keys/:keyId/values`  *(auth required)*

Add a predefined value to a label category.

```json
{ "name": "mom" }
```

`name` may contain `/` to create a nested value (e.g. `"History/Engineering"` → slug `history/engineering`). Each segment is slugified individually.

### `DELETE /me/tag-keys/:keyId/values/:valueId`  *(auth required)*

Remove a predefined value (and all book assignments of that value).

### `GET /me/books/:olId/tags`  *(auth required)*

Get all label assignments for a book.

```json
[
  {
    "key_id": "...",
    "key_name": "Gifted from",
    "key_slug": "gifted-from",
    "value_id": "...",
    "value_name": "mom",
    "value_slug": "mom"
  }
]
```

### `PUT /me/books/:olId/tags/:keyId`  *(auth required)*

Assign a label value to a book. Supply either an existing value ID or a free-form name (which find-or-creates the value in `tag_values`).

```json
{ "value_id": "..." }
// or
{ "value_name": "grandma" }
```

For `select_one` keys, replaces any existing value. For `select_multiple`, adds the value.

### `DELETE /me/books/:olId/tags/:keyId`  *(auth required)*

Remove all label assignments for a given key from a book.

### `DELETE /me/books/:olId/tags/:keyId/values/:valueId`  *(auth required)*

Remove a single value assignment (for `select_multiple` keys).

### `GET /users/:username/labels/:keySlug/*valuePath`  *(public)*

Returns all books for a user that have the given key+value label. Nested: querying `history` also returns books tagged `history/engineering`, `history/science/ancient`, etc.

```
GET /users/alice/labels/genre/history             → books tagged genre:history or genre:history/*
GET /users/alice/labels/genre/history/engineering → books tagged genre:history/engineering or deeper
```

```json
{
  "key_slug": "genre",
  "key_name": "Genre",
  "value_slug": "history",
  "value_name": "History",
  "sub_labels": ["history/engineering", "history/science"],
  "books": [
    {
      "book_id": "...",
      "open_library_id": "OL82592W",
      "title": "The Great Gatsby",
      "cover_url": "https://...",
      "added_at": "2024-01-01T00:00:00Z",
      "rating": 4
    }
  ]
}
```

Returns 404 if the key doesn't exist for this user, or if no values (exact or nested) match the path.

---

## Export

### `GET /me/export/csv`  *(auth required)*

Exports the authenticated user's library as a CSV download. Returns `Content-Type: text/csv` with a `Content-Disposition: attachment` header.

**Query parameters:**
- `shelf` *(optional)* — collection ID to export a single shelf. Omit to export all shelves.

**CSV columns:** Title, Author, ISBN13, Status, Rating, Review, Date Added, Date Read, Date DNF.

---

## Import

### `POST /me/import/goodreads/preview`  *(auth required)*

Accepts a multipart form with a `file` field containing a Goodreads CSV export. Returns a preview without writing to the database.

Response groups rows into `matched`, `ambiguous`, and `unmatched`. See `docs/TODO.md` for full details on the import pipeline.

### `POST /me/import/goodreads/commit`  *(auth required)*

Accepts the confirmed preview payload and writes to the database. Returns `{ imported, failed, errors }`.

---

## Activity Feed

### `GET /me/feed`  *(auth required)*

Returns a chronological feed of activities from users the authenticated user follows. Cursor-based pagination.

**Query parameters:**
- `cursor` *(optional)* — RFC3339Nano timestamp from `next_cursor` to fetch the next page.

```json
{
  "activities": [
    {
      "id": "...",
      "type": "shelved",
      "created_at": "2026-02-24T14:00:00.000000Z",
      "user": {
        "user_id": "...",
        "username": "alice",
        "display_name": "Alice",
        "avatar_url": "https://..."
      },
      "book": {
        "open_library_id": "OL82592W",
        "title": "The Great Gatsby",
        "cover_url": "https://..."
      },
      "shelf_name": "Read",
      "rating": 4,
      "review_snippet": "A timeless classic...",
      "thread_title": null,
      "target_user": null
    }
  ],
  "next_cursor": "2026-02-24T13:00:00.000000Z"
}
```

**Activity types:** `shelved`, `started_book`, `finished_book`, `rated`, `reviewed`, `created_thread`, `followed_user`, `followed_author`, `created_link`.

Fields are conditional on type — `book` is null for `followed_user`, `target_user` is null for book-related activities, etc. `created_link` includes `link_type`, `to_book_ol_id`, and `to_book_title` for the target book. `followed_author` includes `author_key` and `author_name` in the response.

### `GET /users/:username/activity`

Returns recent activity for a specific user. Same response format as `/me/feed`. Cursor-based pagination.

---

## Rate Limiting (Open Library)

All outbound requests to Open Library are routed through a shared rate-limited HTTP client (`api/internal/olhttp`). This uses a token-bucket algorithm (5 requests/second steady-state, burst of 15) to prevent the API from being banned by OL for excessive traffic.

Affected routes: `GET /books/search`, `GET /books/lookup`, `GET /books/:workId`, `GET /books/:workId/editions`, `GET /authors/search`, `GET /authors/:authorKey`, and `POST /me/import/goodreads/preview`.

When the rate limit is saturated, requests wait (up to the 15s client timeout) rather than failing immediately.

---

## Community Links

User-submitted book-to-book connections (sequel, prequel, similar, etc.). Links are upvotable — sorted by vote count on book pages. Both books must exist in the local catalog.

### `GET /books/:workId/links`

Returns all community links for a book, sorted by votes descending. If authenticated, includes whether the caller has voted on each link.

```json
[
  {
    "id": "...",
    "from_book_ol_id": "OL82592W",
    "to_book_ol_id": "OL27448W",
    "to_book_title": "Tender Is the Night",
    "to_book_authors": "F. Scott Fitzgerald",
    "to_book_cover_url": "https://...",
    "link_type": "companion",
    "note": "Same author, similar themes",
    "username": "alice",
    "display_name": "Alice",
    "votes": 3,
    "user_voted": true,
    "created_at": "2026-02-24T14:00:00Z"
  }
]
```

**Link types:** `sequel`, `prequel`, `companion`, `mentioned_in`, `similar`, `adaptation`.

### `POST /books/:workId/links`  *(auth required)*

Submit a community link from this book to another.

```json
{
  "to_work_id": "OL27448W",
  "link_type": "companion",
  "note": "Same author, similar themes"
}
```

Returns 201 with `{ id, created_at }`. Auto-upvotes by the creator. Returns 409 if the user already submitted this exact link.

### `DELETE /links/:linkId`  *(auth required)*

Soft-delete a link (author only). Returns 204.

### `POST /links/:linkId/vote`  *(auth required)*

Upvote a link. Idempotent. Returns 204.

### `DELETE /links/:linkId/vote`  *(auth required)*

Remove upvote. Returns 204.

### `POST /links/:linkId/edits`  *(auth required)*

Propose an edit to a community link. At least one of `proposed_type` or `proposed_note` must be provided. Only one pending edit per user per link.

```json
{
  "proposed_type": "sequel",
  "proposed_note": "Updated explanation"
}
```

```
201 { "id": "...", "created_at": "..." }
400 { "error": "at least one of proposed_type or proposed_note is required" }
409 { "error": "you already have a pending edit for this link" }
```

---

## Admin  *(moderator required)*

All `/admin/*` routes require authentication **and** `is_moderator = true` on the JWT. Non-moderators receive `403 Forbidden`.

### `GET /admin/users?q=<query>&page=<n>`

List all users with moderator status. Supports search by username, display name, or email. Paginated (20 per page).

```json
{
  "users": [
    {
      "user_id": "...",
      "username": "alice",
      "display_name": "Alice",
      "email": "alice@example.com",
      "is_moderator": false
    }
  ],
  "page": 1,
  "has_next": true
}
```

### `PUT /admin/users/:userId/moderator`

Grant or revoke moderator status for a user. The change takes effect on the user's next login (JWT must be re-issued).

```json
{ "is_moderator": true }
```

```
200 { "ok": true, "is_moderator": true }
404 { "error": "user not found" }
```

### `GET /admin/link-edits?status=pending`

List proposed community link edits. Filterable by status (`pending`, `approved`, `rejected`). Returns edits with current and proposed values, book titles, and reviewer info.

```json
[
  {
    "id": "...",
    "book_link_id": "...",
    "username": "alice",
    "display_name": "Alice",
    "proposed_type": "sequel",
    "proposed_note": null,
    "current_type": "similar",
    "current_note": "Same author",
    "from_book_ol_id": "OL82592W",
    "from_book_title": "The Great Gatsby",
    "to_book_ol_id": "OL27448W",
    "to_book_title": "Tender Is the Night",
    "status": "pending",
    "reviewer_name": null,
    "reviewer_comment": null,
    "created_at": "2026-02-25T14:00:00Z",
    "reviewed_at": null
  }
]
```

### `PUT /admin/link-edits/:editId`

Approve or reject a pending community link edit. Approved edits are applied to the link immediately within a transaction.

```json
{
  "action": "approve",
  "comment": "Looks good"
}
```

```
200 { "ok": true, "status": "approved" }
400 { "error": "action must be approve or reject" }
404 { "error": "pending edit not found" }
```

---

## Notifications

### `GET /me/notifications`  *(auth required)*

Returns paginated notifications for the current user, newest first. Cursor-based pagination.

**Query parameters:**
- `cursor` *(optional)* — RFC3339Nano timestamp from `next_cursor` to fetch the next page.

```json
{
  "notifications": [
    {
      "id": "...",
      "notif_type": "new_publication",
      "title": "New work by Brandon Sanderson",
      "body": "Brandon Sanderson published a new work: Wind and Truth",
      "metadata": {
        "author_key": "OL1394865A",
        "author_name": "Brandon Sanderson",
        "new_count": "1",
        "new_titles": "Wind and Truth"
      },
      "read": false,
      "created_at": "2026-02-25T14:00:00.000000Z"
    }
  ],
  "next_cursor": "2026-02-24T13:00:00.000000Z"
}
```

### `GET /me/notifications/unread-count`  *(auth required)*

Returns the count of unread notifications.

```json
{ "count": 3 }
```

### `POST /me/notifications/:notifId/read`  *(auth required)*

Mark a single notification as read. Returns `{ "ok": true }`. Returns 404 if the notification doesn't exist or doesn't belong to the user.

### `POST /me/notifications/read-all`  *(auth required)*

Mark all unread notifications as read. Returns `{ "ok": true }`.

---

## Background: Author Publication Poller

A background goroutine (`notifications.StartPoller`) runs every 6 hours. It:

1. Queries `author_follows` for all distinct followed author keys.
2. For each author, fetches the current work count from `https://openlibrary.org/authors/{key}/works.json`.
3. Compares against the stored snapshot in `author_works_snapshot`.
4. If the count has increased, creates a `new_publication` notification for each follower of that author.
5. On first poll for an author, the snapshot is seeded without generating notifications (to avoid flooding users when they first follow an author).

The poller uses the shared rate-limited OL HTTP client and respects the server's context for graceful shutdown.

---

## Health

### `GET /health`

Returns `200` with `{ "status": "ok", "db": "ok" }` when healthy, `503` if the database is unreachable.
