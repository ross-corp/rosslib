# API Reference

All routes are served by the Go API on `:8080`. The webapp proxies them through Next.js route handlers in `webapp/src/app/api/` so the browser never calls `:8080` directly.

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

### `GET /books/search?q=<title>`

Proxies to Open Library search. Returns up to 20 results.

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

Update own profile. Accepts any subset of `{ display_name, bio, avatar_url }`.

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

## Import

### `POST /me/import/goodreads/preview`  *(auth required)*

Accepts a multipart form with a `file` field containing a Goodreads CSV export. Returns a preview without writing to the database.

Response groups rows into `matched`, `ambiguous`, and `unmatched`. See `docs/TODO.md` for full details on the import pipeline.

### `POST /me/import/goodreads/commit`  *(auth required)*

Accepts the confirmed preview payload and writes to the database. Returns `{ imported, failed, errors }`.

---

## Health

### `GET /health`

Returns `200` with `{ "status": "ok", "db": "ok" }` when healthy, `503` if the database is unreachable.
