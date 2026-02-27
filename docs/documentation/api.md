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

Sets the `token` cookie on success. Returns `401` with `"this account uses Google sign-in"` if the account has no password (Google OAuth-only).

### `POST /auth/google`

```json
{ "google_id": "123456789", "email": "alice@gmail.com", "name": "Alice" }
```

Sign in or register via Google OAuth. Called by the webapp after exchanging a Google authorization code for tokens and fetching user info. Three flows:

1. **Existing Google user** — finds user by `google_id`, issues JWT. Returns `200`.
2. **Existing email user** — finds user by email, links `google_id` to the account, issues JWT. Returns `200`.
3. **New user** — creates account with auto-derived username (from email prefix, with numeric suffix if taken), sets `display_name` from Google profile `name`, sets a random password (user authenticates via Google only), creates default shelves and status tags. Returns `200`.

The webapp handles the full OAuth flow:
- `GET /api/auth/google` — redirects to Google consent screen
- `GET /api/auth/google/callback` — exchanges code for tokens, calls `POST /auth/google`, sets cookie

**Environment variables required (webapp):** `NEXT_PUBLIC_GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `NEXT_PUBLIC_URL`. Set `GOOGLE_CLIENT_ID` and `GOOGLE_CLIENT_SECRET` in the repo root `.env` — `docker-compose.yml` maps `GOOGLE_CLIENT_ID` to `NEXT_PUBLIC_GOOGLE_CLIENT_ID`.

### `POST /auth/forgot-password`

Request a password reset email. Always returns 200 regardless of whether the email exists (to avoid leaking account info).

```json
{ "email": "alice@example.com" }
```

```
200 { "ok": true, "message": "If an account with that email exists, a password reset link has been sent." }
```

Requires SMTP to be configured (`SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASSWORD`, `SMTP_FROM` env vars). When SMTP is not configured, the request is logged but no email is sent. The reset link expires after 1 hour. Previous unused tokens are invalidated when a new one is requested.

### `POST /auth/reset-password`

Reset password using the token from the email link. The token is single-use and expires after 1 hour.

```json
{ "token": "<64-char hex token>", "new_password": "new-pass-8chars" }
```

```
200 { "ok": true }
400 { "error": "invalid or expired reset token" }
400 { "error": "token and new password (min 8 chars) are required" }
```

### `GET /me/account`  *(auth required)*

Returns whether the current user has a password set and whether a Google account is linked. Used by the settings UI to determine which password form to show.

```json
{ "has_password": false, "has_google": true }
```

### `PUT /me/password`  *(auth required)*

Set or change the user's password. If the user already has a password, `current_password` is required. Google OAuth-only users can call this with just `new_password` to enable email+password sign-in.

```json
{ "current_password": "old-pass", "new_password": "new-pass-8chars" }
```

```
200 { "ok": true }
400 { "error": "new password must be at least 8 characters" }
400 { "error": "current password is required" }
401 { "error": "current password is incorrect" }
```

### `DELETE /me/account/data`  *(auth required)*

Permanently deletes all data owned by the authenticated user: user_books, collection_items, collections, tag_keys, tag_values, book_tag_values, genre_ratings, threads, thread_comments, follows, author_follows, book_follows, notifications, activities, book_links, book_link_votes, and book_link_edits. The user account itself is **not** deleted.

**Request body:** none

```
200 { "message": "All data deleted" }
401 { "error": "Authentication required" }
```

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

### `GET /books/popular`

Returns up to 12 popular books from the local catalog, ordered by reads count. Draws from the `book_stats` table. Used as the landing state on the search page when no query is entered.

```json
[
  {
    "key": "OL82592W",
    "title": "The Great Gatsby",
    "authors": ["F. Scott Fitzgerald"],
    "cover_url": "https://covers.openlibrary.org/b/id/8410459-M.jpg",
    "publish_year": 1925,
    "average_rating": 4.2,
    "rating_count": 15,
    "already_read_count": 42
  }
]
```

Returns an empty array if no books have stats yet.

### `GET /books/lookup?isbn=<isbn>`

Looks up a single book by ISBN via Open Library. Upserts the result into the local `books` table and returns it. Returns 404 if not found.

### `POST /books/scan`

Accepts a `multipart/form-data` image upload (field name `image`). Decodes the image, detects an EAN-13 barcode (ISBN), looks up the book via Open Library, upserts it locally, and returns the result. Supported formats: JPEG, PNG, GIF.

Returns `422` if no barcode is detected (with a `hint` field), `404` if the ISBN is detected but no book matches, and `200` on success:

```json
{
  "isbn": "9780261103573",
  "book": {
    "key": "/works/OL27448W",
    "title": "The Lord of the Rings",
    "authors": ["J.R.R. Tolkien"],
    "publish_year": 1954,
    "cover_url": "https://covers.openlibrary.org/b/id/...-M.jpg",
    ...
  }
}
```

### `GET /books/:workId`

Returns book details by its bare OL work ID (e.g. `OL82592W`). Fetches enriched data from Open Library with local fallback. 404 if not found.

The `authors` field returns an array of objects with `name` (string) and `key` (string or null). The `key` is the bare OL author ID (e.g. `OL23919A`) when available from Open Library, or `null` for local-only records. Includes a `subjects` array (up to 10 strings) sourced from the Open Library work's `subjects` field, with a fallback to the local book record's comma-separated `subjects` column.

```json
{ "authors": [{ "name": "Author Name", "key": "OL23919A" }] }
```

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

### `GET /books/:workId/stats`

Returns precomputed aggregate stats for a book from the `book_stats` cache table. Stats are refreshed asynchronously when users change statuses, ratings, or reviews.

```json
{
  "reads_count": 42,
  "want_to_read_count": 15,
  "average_rating": 4.2,
  "rating_count": 30,
  "review_count": 8
}
```

`average_rating` is null when no users have rated the book. Returns 404 if the book is not in the local catalog.

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

## User Books

### `POST /me/books`  *(auth required)*

Add a book to the user's library. Upserts the book into the global catalog.

```json
{
  "open_library_id": "OL82592W",
  "title": "The Great Gatsby",
  "cover_url": "https://...",
  "isbn13": "9780743273565",
  "authors": ["F. Scott Fitzgerald"],
  "publication_year": 1925,
  "status_slug": "want-to-read"
}
```

### `PATCH /me/books/:olId`  *(auth required)*

Update metadata on a book in the user's library. Only provided fields are updated.

```json
{
  "rating": 4,
  "review_text": "Great book.",
  "spoiler": false,
  "date_read": "2024-06-01T00:00:00Z",
  "date_dnf": null,
  "progress_pages": 150,
  "progress_percent": 45,
  "status_slug": "currently-reading",
  "selected_edition_key": "OL123M",
  "selected_edition_cover_url": "https://covers.openlibrary.org/b/id/12345-M.jpg"
}
```

`selected_edition_key` and `selected_edition_cover_url` allow the user to select a specific edition of a book. When set, the edition's cover is displayed instead of the default work cover on profile pages, shelf views, and the book detail page.

### `DELETE /me/books/:olId`  *(auth required)*

Remove a book from the user's library. Also cleans up tag assignments and collection items.

### `GET /me/books/:olId/status`  *(auth required)*

Returns the user's status, rating, review, progress, and edition selection for a book.

```json
{
  "status_value_id": "...",
  "status_name": "Currently Reading",
  "status_slug": "currently-reading",
  "rating": 4,
  "review_text": "Great so far.",
  "spoiler": false,
  "date_read": null,
  "date_dnf": null,
  "progress_pages": 150,
  "progress_percent": 45,
  "selected_edition_key": "OL123M",
  "selected_edition_cover_url": "https://covers.openlibrary.org/b/id/12345-M.jpg"
}
```

### `PUT /me/books/:olId/status`  *(auth required)*

Set the reading status for a book.

```json
{ "slug": "currently-reading" }
```

### `GET /me/books/status-map`  *(auth required)*

Returns a map of all the user's books to their status slugs.

```json
{ "OL82592W": "read", "OL27448W": "currently-reading" }
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

## Genre Ratings

Users can rate how strongly a book fits each genre on a 0–10 scale. Aggregate averages are shown publicly on book detail pages; individual ratings are visible to the authenticated user.

### `GET /books/:workId/genre-ratings`

Returns aggregate genre ratings for a book, sorted by rater count descending.

```json
[
  {
    "genre": "Science fiction",
    "average": 7.3,
    "rater_count": 12
  },
  {
    "genre": "Fiction",
    "average": 9.1,
    "rater_count": 8
  }
]
```

Returns an empty array if no ratings exist.

### `GET /me/books/:olId/genre-ratings`  *(auth required)*

Returns the current user's genre ratings for a book.

```json
[
  {
    "genre": "Science fiction",
    "rating": 8,
    "updated_at": "2026-02-25T14:00:00Z"
  }
]
```

### `PUT /me/books/:olId/genre-ratings`  *(auth required)*

Set or update genre ratings for a book. Accepts an array of `{genre, rating}` objects. Ratings of 0 or null remove the genre rating. Genres must be from the predefined list.

```json
[
  { "genre": "Science fiction", "rating": 8 },
  { "genre": "Fiction", "rating": 6 },
  { "genre": "Horror", "rating": 0 }
]
```

```
200 { "ok": true }
400 { "error": "invalid genre: ..." }
400 { "error": "rating must be 0-10" }
404 { "error": "book not found" }
```

**Valid genres:** Fiction, Non-fiction, Fantasy, Science fiction, Mystery, Romance, Horror, Thriller, Biography, History, Poetry, Children.

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

### `GET /authors/:authorKey?limit=24&offset=0`

Fetches author detail from Open Library including a paginated slice of their works.

**Query parameters:**
- `limit` *(optional, default 24, max 100)* — number of works to return.
- `offset` *(optional, default 0)* — offset into the works list.

```json
{
  "key": "OL26320A",
  "name": "J.R.R. Tolkien",
  "bio": "John Ronald Reuel Tolkien was an English writer...",
  "birth_date": "3 January 1892",
  "death_date": "2 September 1973",
  "photo_url": "https://covers.openlibrary.org/a/id/6257741-L.jpg",
  "links": [{ "title": "Wikipedia", "url": "https://en.wikipedia.org/wiki/J._R._R._Tolkien" }],
  "work_count": 392,
  "works": [
    {
      "key": "OL27448W",
      "title": "The Lord of the Rings",
      "cover_url": "https://covers.openlibrary.org/b/id/8406786-M.jpg"
    }
  ]
}
```

`bio`, `birth_date`, `death_date`, `photo_url`, `links`, and `cover_url` on works may be null.

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

### `GET /users?q=<query>&page=<n>&sort=<newest|books|followers>`

Search/browse users by username or display name. 20 per page.

**Query parameters:**
- `q` *(optional)* — search by username or display name
- `page` *(optional, default 1)* — pagination
- `sort` *(optional, default `newest`)* — sort order: `newest` (registration date), `books` (most books in library), `followers` (most followers)

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
  "books_read": 42,
  "author_key": null
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

`collection_type` is one of `"shelf"`, `"tag"`, or `"computed"`. See `docs/organization.md` for the distinction.

Computed lists include an additional `computed` object with metadata about the set operation:

```json
{
  "id": "...",
  "name": "Books in both lists",
  "slug": "books-in-both-lists",
  "exclusive_group": "",
  "collection_type": "computed",
  "item_count": 12,
  "computed": {
    "operation": "intersection",
    "is_continuous": true,
    "last_computed_at": "2026-02-25T14:00:00Z",
    "source_a_name": "Want to Read",
    "source_b_name": "Favorites"
  }
}
```

### `GET /users/:username/shelves/:slug`

Returns a shelf with its full book list. Computed lists also include the `computed` object.

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

### `POST /me/shelves/set-operation`  *(auth required)*

Compute a set operation (union, intersection, or difference) between two collections. Both must belong to the authenticated user.

```json
{
  "collection_a": "<uuid>",
  "collection_b": "<uuid>",
  "operation": "intersection"
}
```

Returns `{ operation, collection_a, collection_b, result_count, books[] }`.

### `POST /me/shelves/set-operation/save`  *(auth required)*

Same as above, but also saves the result as a new shelf. Accepts an additional `name` field. Returns `{ id, name, slug, book_count, is_continuous }`. If `is_continuous` is true, the list auto-refreshes when viewed — the operation is re-executed against the current source collections each time `GET /users/:username/shelves/:slug` is called.

```json
{
  "collection_a": "<uuid>",
  "collection_b": "<uuid>",
  "operation": "union",
  "name": "Combined Reading",
  "is_continuous": true
}
```

### `POST /me/shelves/cross-user-compare`  *(auth required)*

Compare one of your collections with another user's public collection. Respects privacy — returns 403 for private profiles you don't follow.

```json
{
  "my_collection": "<uuid>",
  "their_username": "alice",
  "their_slug": "want-to-read",
  "operation": "intersection"
}
```

Returns `{ operation, my_collection, their_username, their_slug, result_count, books[] }`.

### `POST /me/shelves/cross-user-compare/save`  *(auth required)*

Same as above, but also saves the result as a new shelf. Accepts an additional `name` field. Returns `{ id, name, slug, book_count, is_continuous }`. If `is_continuous` is true, the list auto-refreshes when viewed, resolving the other user's collection by stored username+slug on each view (respecting privacy).

```json
{
  "my_collection": "<uuid>",
  "their_username": "alice",
  "their_slug": "want-to-read",
  "operation": "intersection",
  "name": "Books Alice wants that I've read",
  "is_continuous": true
}
```

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

Response groups rows into `matched`, `ambiguous`, and `unmatched`. The lookup chain tries: local DB by ISBN, OL direct ISBN endpoint, OL search by ISBN, OL search by cleaned title+author, OL search by title only, OL comma-subtitle retry, and finally Google Books as a fallback. The Google Books fallback searches by ISBN then by title+author, and maps results back to Open Library by re-searching OL with the title/author from Google. Set the optional `GOOGLE_BOOKS_API_KEY` env var for higher rate limits (free tier: 1,000 req/day); the fallback works without a key.

### `POST /me/import/goodreads/commit`  *(auth required)*

Accepts the confirmed preview payload and writes to the database. Returns `{ imported, failed, errors, pending_saved }`.

The request body includes an optional `unmatched_rows` array alongside `rows` and `shelf_mappings`. Unmatched rows are persisted to the `pending_imports` collection for later manual resolution.

**`shelf_mappings` actions:**
- `"tag"` — creates a standalone tag key per shelf (default)
- `"create_label"` — groups the shelf as a value under a new label key specified by `label_name`
- `"existing_label"` — adds the shelf as a value under an existing label key specified by `label_key_id`
- `"skip"` — ignores the shelf

---

## Pending Imports

### `GET /me/imports/pending`  *(auth required)*

Returns unmatched import rows saved from previous Goodreads imports. Only returns rows with status `unmatched`.

```json
[
  {
    "id": "...",
    "title": "Destinies, Feb 1980",
    "author": "James Baen",
    "isbn13": "",
    "exclusive_shelf": "read",
    "custom_shelves": [],
    "rating": null,
    "review_text": "",
    "date_read": "",
    "date_added": "2024/01/15",
    "created": "2026-02-26T12:00:00Z"
  }
]
```

### `PATCH /me/imports/pending/:id`  *(auth required)*

Resolve or dismiss a pending import.

**Dismiss** — marks as resolved without importing:
```json
{ "action": "dismiss" }
```

**Resolve** — matches to an Open Library work ID, creates the book and user_book:
```json
{ "action": "resolve", "ol_id": "OL82592W" }
```

```
200 { "ok": true }
200 { "ok": true, "book_id": "..." }  (resolve)
400 { "error": "action must be 'resolve' or 'dismiss'" }
404 { "error": "Pending import not found" }
```

### `DELETE /me/imports/pending/:id`  *(auth required)*

Permanently deletes a pending import row.

```
200 { "ok": true }
404 { "error": "Pending import not found" }
```

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

## Discussion Threads

### `GET /books/:workId/threads`

Returns all discussion threads for a book, ordered by most recent first.

```json
[
  {
    "id": "...",
    "book_id": "...",
    "user_id": "...",
    "username": "alice",
    "display_name": "Alice",
    "avatar_url": "https://...",
    "title": "What did the ending mean?",
    "body": "I just finished and...",
    "spoiler": true,
    "created_at": "2026-02-25T14:00:00Z",
    "comment_count": 3
  }
]
```

### `GET /threads/:threadId`

Returns a single thread with all its comments.

```json
{
  "thread": { ... },
  "comments": [
    {
      "id": "...",
      "thread_id": "...",
      "user_id": "...",
      "username": "bob",
      "display_name": "Bob",
      "avatar_url": null,
      "parent_id": null,
      "body": "I think it meant...",
      "created_at": "2026-02-25T15:00:00Z"
    }
  ]
}
```

### `POST /books/:workId/threads`  *(auth required)*

Create a new discussion thread on a book. Records a `created_thread` activity and notifies book followers.

```json
{ "title": "What did the ending mean?", "body": "I just finished and...", "spoiler": true }
```

`title` max 500 characters, `body` max 10,000 characters.

```
201 { "id": "...", "created_at": "..." }
400 { "error": "title and body are required" }
400 { "error": "title must be 500 characters or fewer" }
400 { "error": "body must be 10,000 characters or fewer" }
404 { "error": "book not found" }
```

### `DELETE /threads/:threadId`  *(auth required)*

Soft-delete a thread (author only). Returns 204.

### `POST /threads/:threadId/comments`  *(auth required)*

Add a comment to a thread. Set `parent_id` to reply to a top-level comment (one level of nesting only).

```json
{ "body": "I think it meant...", "parent_id": null }
```

`body` max 5,000 characters.

```
201 { "id": "...", "created_at": "..." }
400 { "error": "comment must be 5,000 characters or fewer" }
400 { "error": "replies can only be one level deep" }
```

### `DELETE /threads/:threadId/comments/:commentId`  *(auth required)*

Soft-delete a comment (author only). Returns 204.

### `GET /books/:workId/similar-threads?title=<title>`

Find existing threads on a book whose titles are similar to the given title. Uses PostgreSQL `pg_trgm` trigram similarity with a threshold of 0.3. Returns up to 5 results sorted by similarity score. Used by the thread creation form to suggest existing discussions before posting a duplicate.

```json
[
  {
    "id": "...",
    "title": "Thoughts on the ending?",
    "username": "alice",
    "comment_count": 5,
    "similarity": 0.65,
    ...
  }
]
```

### `GET /threads/:threadId/similar`

Returns threads on the same book whose titles are similar to the given thread. Same similarity mechanism and response format as the title-based search. Shown on thread detail pages under "Similar Discussions".

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
      "is_moderator": false,
      "author_key": null
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

### `PUT /admin/users/:userId/author`

Set or clear the Open Library author key for a user. When set, the user's profile displays an "Author" badge linking to their author page.

```json
{ "author_key": "OL23919A" }
```

```
200 { "ok": true, "author_key": "OL23919A" }
404 { "error": "user not found" }
```

Send `{ "author_key": "" }` or `{ "author_key": null }` to clear.

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

## Feedback

### `POST /feedback`  *(auth required)*

Submit a bug report or feature request.

```json
{
  "type": "bug",
  "title": "Search results don't load",
  "description": "When I search for a book, the page stays blank.",
  "steps_to_reproduce": "1. Go to search\n2. Type a query\n3. See blank page",
  "severity": "high"
}
```

`type` is `"bug"` or `"feature"`. For bug reports, `steps_to_reproduce` (optional) and `severity` (`"low"`, `"medium"`, `"high"`, optional) are accepted. Feature requests only need `title` and `description`.

```
201 { "id": "...", "created_at": "..." }
400 { "error": "title and description are required" }
400 { "error": "type must be bug or feature" }
```

### `GET /admin/feedback?status=open`  *(moderator required)*

List feedback submissions. Filterable by `status` (`open`, `closed`). Returns all submissions sorted by newest first.

```json
[
  {
    "id": "...",
    "user_id": "...",
    "username": "alice",
    "display_name": "Alice",
    "type": "bug",
    "title": "Search results don't load",
    "description": "When I search...",
    "steps_to_reproduce": "1. Go to search...",
    "severity": "high",
    "status": "open",
    "created_at": "2026-02-25T14:00:00Z"
  }
]
```

### `PATCH /admin/feedback/:feedbackId`  *(moderator required)*

Toggle feedback status between open and closed.

```json
{ "status": "closed" }
```

```
200 { "ok": true, "status": "closed" }
400 { "error": "status must be open or closed" }
404 { "error": "Feedback not found" }
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
