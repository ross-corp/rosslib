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

Creates a user and sets the `token` cookie. Also creates the three default labels (Want to Read, Currently Reading, Read).

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
3. **New user** — creates account with auto-derived username (from email prefix, with numeric suffix if taken), sets `display_name` from Google profile `name`, sets a random password (user authenticates via Google only), creates default labels and status tags. Returns `200`.

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

### `PUT /me/email`  *(auth required)*

Change the user's email address. Requires the current password for verification. Sets `email_verified` to `false` on the account after the change.

```json
{ "new_email": "newemail@example.com", "current_password": "current-pass" }
```

```
200 { "message": "Email updated" }
400 { "error": "New email and current password are required" }
400 { "error": "Invalid email address" }
400 { "error": "Current password is incorrect" }
400 { "error": "New email is the same as the current email" }
409 { "error": "Email is already in use" }
```

### `DELETE /me/account/data`  *(auth required)*

Permanently deletes all data owned by the authenticated user: user_books, collection_items, collections, tag_keys, tag_values, book_tag_values, genre_ratings, threads, thread_comments, follows, author_follows, book_follows, notifications, activities, book_links, book_link_votes, and book_link_edits. The user account itself is **not** deleted.

**Request body:** none

```
200 { "message": "All data deleted" }
401 { "error": "Authentication required" }
```

### `DELETE /me/account`  *(auth required)*

Permanently deletes all data owned by the authenticated user **and** the user account itself. The webapp proxy clears the auth cookie on success.

**Request body:** none

```
200 { "message": "Account deleted" }
401 { "error": "Authentication required" }
500 { "error": "Failed to delete account" }
```

---

## Books

### `GET /books/search?q=<title>[&page=1][&sort=reads|rating][&year_min=N][&year_max=N]`

Searches both local catalog and Open Library concurrently. Local matches appear first, followed by external results deduplicated by work ID. Returns up to 20 results per page.

**Query parameters:**
- `q` *(required)* — search query
- `page` *(optional, default 1)* — page number for pagination. Each page returns up to 20 results.
- `sort` *(optional)* — sort order: `reads` (most read) or `rating` (highest rated)
- `year_min` / `year_max` *(optional)* — filter by publication year range

```json
{
  "total": 1234,
  "page": 1,
  "results": [
    {
      "key": "/works/OL82592W",
      "title": "The Great Gatsby",
      "authors": ["F. Scott Fitzgerald"],
      "publish_year": 1925,
      "isbn": ["9780743273565"],
      "cover_url": "https://covers.openlibrary.org/b/id/8410459-M.jpg",
      "edition_count": 120,
      "subjects": ["Fiction", "Classic Literature", "American Literature"],
      "link_count": 3
    }
  ]
}
```

`authors`, `isbn`, `cover_url`, and `subjects` may be null. `link_count` is the number of community links (related books) for local books; 0 for Open Library-only results. For local books, `subjects` is derived from the book's comma-separated subjects column (first 3). For Open Library results, `subjects` comes from the OL search response (first 3).

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

### `GET /books/trending?period=week&limit=10`

Returns books with the most new `user_books` activity in a recent time window. Queries `user_books` rows created in the last 7 days (default) or 30 days, grouped by book, ordered by activity count descending. Used on the search landing page as a "Trending This Week" section.

**Query parameters:**
- `period` — `week` (default, 7 days) or `month` (30 days)
- `limit` — max books to return (default 10, max 50)

```json
[
  {
    "key": "OL82592W",
    "title": "The Great Gatsby",
    "authors": ["F. Scott Fitzgerald"],
    "cover_url": "https://covers.openlibrary.org/b/id/8410459-M.jpg",
    "publish_year": 1925,
    "activity_count": 8
  }
]
```

Returns an empty array if no recent activity.

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

The `authors` field returns an array of objects with `name` (string) and `key` (string or null). The `key` is the bare OL author ID (e.g. `OL23919A`) when available from Open Library, or `null` for local-only records. Includes a `subjects` array (up to 10 strings) sourced from the Open Library work's `subjects` field, with a fallback to the local book record's comma-separated `subjects` column. Response also includes a `series` array (or null) with the book's series memberships, each containing `series_id`, `name`, and `position`.

**Auto-populated series data:** On first view of a book that has no series links, the endpoint automatically checks the Open Library editions response for `series` fields and the work's subjects for series-like patterns (e.g. containing "trilogy", "saga", etc.). When found, series and book_series records are created automatically. This is best-effort — not all OL works have series data. Series population is logged for visibility into coverage.

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

### `GET /books/:workId/readers`  *(optional auth)*

Returns up to 5 users the viewer follows who have this book in their library. Returns an empty array for unauthenticated users or if no followed users have the book.

```json
[
  {
    "user_id": "abc123",
    "username": "alice",
    "display_name": "Alice",
    "avatar_url": "/api/files/users/abc123/avatar.jpg",
    "status_name": "Currently Reading"
  }
]
```

`display_name`, `avatar_url`, and `status_name` may be null.

### `GET /books/:workId/reviews`  *(optional auth)*

Returns all community reviews for a book. Each user appears at most once (most recent review).

**Query params:**
- `sort` — `newest` (default), `oldest`, `highest` (rating DESC), `lowest` (rating ASC), `most_liked` (like count DESC)

The viewer's own review always appears first regardless of sort order. When authenticated, reviews from blocked/blocking users are excluded.

```json
[
  {
    "user_id": "abc123",
    "username": "alice",
    "display_name": "Alice",
    "avatar_url": "https://...",
    "rating": 4,
    "review_text": "Loved it.",
    "spoiler": false,
    "date_read": "2025-06-15T00:00:00Z",
    "date_dnf": null,
    "date_added": "2025-06-20T14:32:10Z",
    "is_followed": true,
    "like_count": 3,
    "liked_by_me": false
  }
]
```

### `POST /books/:workId/reviews/:userId/like`  *(auth required)*

Toggle like on a review. If not liked, creates a like; if already liked, removes it. Cannot like your own review. Records a `liked_review` activity and sends a `review_liked` notification to the review author.

Returns `{ "liked": true }` or `{ "liked": false }`.

### `GET /books/:workId/reviews/:userId/like`  *(auth required)*

Check if the current user has liked a specific review.

Returns `{ "liked": true }` or `{ "liked": false }`.

### `GET /books/:workId/reviews/:userId/comments`

List comments on a review, ordered chronologically. Returns an array of comment objects with `id`, `user_id`, `username`, `display_name`, `avatar_url`, `body`, `created_at`.

### `POST /books/:workId/reviews/:userId/comments`  *(auth required)*

Add a comment to a review. The review must exist (non-empty `review_text` on the user's `user_books` record).

```json
{ "body": "Great review, I totally agree!" }
```

`body` is required, max 2000 characters. Generates a `review_comment` notification for the review author (unless commenting on own review).

### `DELETE /review-comments/:commentId`  *(auth required)*

Soft-delete a review comment. Only the comment author or a moderator can delete.

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
  "device_total_pages": 320,
  "status_slug": "currently-reading",
  "selected_edition_key": "OL123M",
  "selected_edition_cover_url": "https://covers.openlibrary.org/b/id/12345-M.jpg"
}
```

`device_total_pages` overrides the catalog `page_count` for progress percentage calculations. When set, page-based progress uses `device_total_pages` as the denominator instead of `books.page_count`. Send `0` or `null` to clear.

`selected_edition_key` and `selected_edition_cover_url` allow the user to select a specific edition of a book. When set, the edition's cover is displayed instead of the default work cover on profile pages, label views, and the book detail page.

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
  "date_started": "2026-02-01T00:00:00Z",
  "date_added": "2026-01-15T12:00:00Z",
  "progress_pages": 150,
  "progress_percent": 45,
  "device_total_pages": 320,
  "selected_edition_key": "OL123M",
  "selected_edition_cover_url": "https://covers.openlibrary.org/b/id/12345-M.jpg"
}
```

### `GET /me/books/:olId/editions`  *(auth required)*

Returns the user's currently selected edition alongside the full editions list from Open Library. Combines the `selected_edition_key` and `selected_edition_cover_url` from the user's `user_books` record with the OL editions response.

**Query parameters:**
- `limit` *(optional, default 20, max 100)* — number of editions to return
- `offset` *(optional, default 0)* — offset into the editions list

```json
{
  "selected_edition_key": "OL123M",
  "selected_edition_cover_url": "https://covers.openlibrary.org/b/id/12345-M.jpg",
  "editions": {
    "entries": [...],
    "size": 251
  }
}
```

`selected_edition_key` and `selected_edition_cover_url` are null when no edition is selected. The `editions` object is the raw Open Library editions response.

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

### `GET /users/:username/books/search?q=<query>`  *(optional auth)*

Search within a user's library by book title or author. Respects profile privacy (returns 403 for private profiles the viewer doesn't follow). Returns up to 50 results ordered by date added descending.

**Query parameters:**
- `q` *(required)* — search query, matched via case-insensitive LIKE against title and authors
- `limit` *(optional, default 50, max 100)* — maximum results

```json
{
  "books": [
    {
      "book_id": "...",
      "open_library_id": "OL82592W",
      "title": "The Great Gatsby",
      "cover_url": "https://...",
      "authors": "F. Scott Fitzgerald",
      "rating": 4,
      "added_at": "2024-06-01T00:00:00Z"
    }
  ]
}
```

Returns an empty `books` array if `q` is empty or no matches are found.

---

## Genres

### `GET /genres`

Returns genres derived from the `subjects` field on books in the local catalog, with book counts. Sorted by book count descending.

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

### `GET /genres/:slug/books?page=1&limit=20&sort=title|rating|year`

Returns books matching a genre from the local catalog, filtered by the `subjects` field on books. Paginated.

**Query parameters:**
- `page` *(optional, default 1)* — page number
- `limit` *(optional, default 20, max 100)* — results per page
- `sort` *(optional)* — sort order: `title` (A-Z), `rating` (highest first), `year` (newest first). Default: no explicit sort.

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

### `GET /books/:workId/followers/count`

Returns the number of users following a book. Public endpoint — no authentication required.

```json
{ "follower_count": 12 }
```

Returns `{ "follower_count": 0 }` if the book is not in the local catalog or has no followers.

### `POST /books/:workId/follow`  *(auth required)*

Follow a book. You'll be notified when new threads are created on it.

```
200 { "message": "Following book" }
200 { "message": "Already following" }
404 { "error": "Book not found" }
```

### `DELETE /books/:workId/follow`  *(auth required)*

Unfollow a book.

```
200 { "message": "Unfollowed book" }
200 { "message": "Not following" }
```

### `GET /books/:workId/follow`  *(auth required)*

Check if you follow a book. Returns `{ "following": true/false }`. Returns `{ "following": false }` if the book is not in the local catalog.

### `GET /me/followed-books`  *(auth required)*

List books you follow, newest first. Supports pagination via `limit` (default 50, max 50) and `offset` (default 0) query params.

```json
{
  "books": [
    {
      "open_library_id": "OL82592W",
      "title": "The Name of the Wind",
      "authors": ["Patrick Rothfuss"],
      "cover_url": "https://covers.openlibrary.org/b/id/1234567-M.jpg"
    }
  ],
  "total": 73
}
```

`authors` and `cover_url` may be null.

---

## Series

### `GET /series/search?q=<name>`

Search series by name. Returns up to 20 results sorted by book count descending.

**Query parameters:**
- `q` *(required)* — search query, matched via case-insensitive LIKE against series name

```json
{
  "results": [
    {
      "id": "abc123",
      "name": "The Lord of the Rings",
      "description": "A fantasy trilogy by J.R.R. Tolkien.",
      "book_count": 3
    }
  ]
}
```

Returns `{ "results": [] }` if `q` is empty or no matches are found.

### `GET /books/:workId/series`

Returns series memberships for a book.

```json
[
  { "series_id": "abc123", "name": "The Lord of the Rings", "description": null, "position": 3 }
]
```

### `GET /series/:seriesId`  *(optional auth)*

Returns series details with an ordered list of books. If authenticated, each book includes `viewer_status` (the logged-in user's reading status slug, e.g. `"finished"`).

```json
{
  "id": "abc123",
  "name": "The Lord of the Rings",
  "description": "...",
  "books": [
    {
      "book_id": "...",
      "open_library_id": "OL27448W",
      "title": "The Fellowship of the Ring",
      "cover_url": "...",
      "authors": "J.R.R. Tolkien",
      "position": 1,
      "viewer_status": "finished"
    }
  ]
}
```

### `PATCH /series/:seriesId`  *(auth required)*

Update a series name and/or description. Only provided fields are updated.

```json
{ "name": "The Lord of the Rings", "description": "A fantasy trilogy by J.R.R. Tolkien." }
```

Validation: `name` cannot be empty if provided. Returns `200` with the updated series `{ id, name, description }`.

### `DELETE /series/:seriesId`  *(auth required)*

Delete an empty series. The series must have zero `book_series` links (no books). Returns 400 if the series still has books.

```
200 { "ok": true }
400 { "error": "Cannot delete series that still has books" }
404 { "error": "Series not found" }
```

### `POST /books/:workId/series`  *(auth required)*

Add a book to a series. Finds or creates the series by name.

```json
{ "series_name": "The Lord of the Rings", "position": 3 }
```

Returns `201` on new link, `200` if the link already exists (position is updated if provided).

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

### `GET /authors/:authorKey/series?name=<authorName>`

Returns series containing books whose `authors` field matches the given name. The `name` query param is required.

```json
[
  {
    "id": "abc123",
    "name": "The Lord of the Rings",
    "description": "Epic fantasy trilogy",
    "book_count": 3
  }
]
```

### `POST /authors/:authorKey/follow`  *(auth required)*

Follow an author. Accepts optional `{ "author_name": "..." }` to cache the display name.

### `DELETE /authors/:authorKey/follow`  *(auth required)*

Unfollow an author.

### `GET /authors/:authorKey/follow`  *(auth required)*

Check if you follow an author. Returns `{ "following": true/false }`.

### `GET /me/followed-authors`  *(auth required)*

List authors you follow. Supports pagination via `limit` (default 50, max 50) and `offset` (default 0) query params.

```json
{
  "authors": [
    {
      "author_key": "OL26320A",
      "author_name": "J.R.R. Tolkien"
    }
  ],
  "total": 1
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
  "banner_url": null,
  "is_private": false,
  "member_since": "2024-01-01T00:00:00Z",
  "is_following": false,
  "followers_count": 10,
  "following_count": 5,
  "friends_count": 3,
  "books_read": 42,
  "currently_reading_count": 2,
  "total_books": 67,
  "total_pages_read": 14320,
  "author_key": null,
  "top_genres": [
    { "name": "fiction", "count": 15 },
    { "name": "mystery", "count": 8 }
  ]
}
```

### `PATCH /users/me`  *(auth required)*

Update own display name and byline. Accepts any subset of `{ display_name, bio, is_private }`.

Validation: `display_name` max 100 characters, `bio` max 2000 characters. Returns 400 if exceeded.

### `POST /me/avatar`  *(auth required)*

Upload or replace the authenticated user's profile picture. Accepts a `multipart/form-data` body with an `avatar` field containing the image file (JPEG, PNG, GIF, or WebP; max 5 MB).

Content-type is detected from the file's magic bytes — the `Content-Type` header on the file part is not trusted.

```
200 { "avatar_url": "http://localhost:9000/rosslib/avatars/<userId>.jpg" }
400 { "error": "unsupported image type: ..." }
400 { "error": "file too large (max 5 MB)" }
```

The returned URL is stored in `users.avatar_url` and returned on subsequent `GET /users/:username` calls. In production, point `MINIO_PUBLIC_URL` to the S3 bucket or CDN origin — the URL format is `{MINIO_PUBLIC_URL}/{MINIO_BUCKET}/avatars/{userId}.{ext}`.

### `POST /me/banner`  *(auth required)*

Upload or replace the authenticated user's profile banner image. Accepts a `multipart/form-data` body with a `banner` field containing the image file (JPEG, PNG, GIF, or WebP; max 10 MB). Recommended dimensions: 1200x300.

```
200 { "banner_url": "/api/files/<collectionId>/<userId>/<filename>" }
400 { "error": "No banner file provided" }
400 { "error": "Failed to process uploaded file" }
```

The returned URL is included in subsequent `GET /users/:username` responses as `banner_url`.

### `GET /users/:username/followers`  *(optional auth)*

Returns a paginated list of users who follow this user. Respects privacy — returns 403 for private profiles the viewer doesn't follow.

**Query parameters:**
- `page` *(optional, default 1)* — page number
- `limit` *(optional, default 50, max 50)* — results per page

```json
[
  {
    "user_id": "...",
    "username": "bob",
    "display_name": "Bob",
    "avatar_url": "/api/files/..."
  }
]
```

### `GET /users/:username/following`  *(optional auth)*

Returns a paginated list of users this user follows. Respects privacy — returns 403 for private profiles the viewer doesn't follow.

**Query parameters:**
- `page` *(optional, default 1)* — page number
- `limit` *(optional, default 50, max 50)* — results per page

```json
[
  {
    "user_id": "...",
    "username": "alice",
    "display_name": "Alice",
    "avatar_url": "/api/files/..."
  }
]
```

### `GET /users/:username/reviews?page=1&limit=20&sort=newest`

Returns paginated reviews (user_books with non-empty `review_text`) for a user.

**Query parameters:**
- `page` *(optional, default 1)* — page number
- `limit` *(optional, default 20, max 100)* — results per page
- `sort` *(optional, default `newest`)* — sort order: `newest` (date added descending), `oldest` (date added ascending), `highest_rating` (highest rated first), `lowest_rating` (lowest rated first)

```json
{
  "reviews": [
    {
      "open_library_id": "OL82592W",
      "title": "The Great Gatsby",
      "cover_url": "https://...",
      "rating": 4,
      "review_text": "A timeless classic.",
      "spoiler": false,
      "date_read": "2024-06-01T00:00:00Z",
      "date_added": "2024-06-02T00:00:00Z",
      "like_count": 3
    }
  ],
  "total": 42,
  "page": 1
}
```

`cover_url`, `rating`, and `date_read` may be null.

### `GET /users/:username/timeline?year=<YYYY>`  *(optional auth)*

Returns finished books for a user grouped by month for a given year. Defaults to the current year. Respects profile privacy settings.

```json
{
  "year": 2026,
  "months": [
    {
      "month": 1,
      "books": [
        {
          "book_id": "...",
          "open_library_id": "OL82592W",
          "title": "The Great Gatsby",
          "cover_url": "https://...",
          "rating": 4,
          "date_read": "2026-01-15T00:00:00Z"
        }
      ]
    }
  ]
}
```

Empty months are omitted. Books within each month are sorted by `date_read` ascending.

### `POST /users/:username/follow`  *(auth required)*

Follow a user. Status is `active` immediately (private account approval not yet enforced).

### `DELETE /users/:username/follow`  *(auth required)*

Unfollow a user.

### `GET /me/blocks`  *(auth required)*

List all users the current user has blocked, ordered by most recently blocked first.

```json
[
  {
    "id": "user_id",
    "username": "blockeduser",
    "display_name": "Blocked User",
    "avatar_url": "/api/files/...",
    "blocked_at": "2026-01-15T14:00:00.000Z"
  }
]
```

### `POST /users/:username/block`  *(auth required)*

Block a user. Also removes any existing follow relationship in both directions. Returns `{ "blocked": true }`.

Blocking effects:
- Blocked user's reviews are hidden from book pages for the blocker
- Blocked user is hidden from search results
- Blocked user's activity is excluded from the feed
- Neither user can follow the other
- Visiting the blocked user's profile shows restricted view

### `DELETE /users/:username/block`  *(auth required)*

Unblock a user. Returns `{ "blocked": false }`.

### `GET /users/:username/block`  *(auth required)*

Check if you have blocked a user. Returns `{ "blocked": true/false }`.

---

## Reading Goals

### `GET /me/goals`  *(auth required)*

List all reading goals for the current user.

```json
[
  { "id": "...", "year": 2026, "target": 25 }
]
```

### `PUT /me/goals/:year`  *(auth required)*

Create or update a reading goal for a year. Records a `goal_set` activity.

```json
{ "target": 25 }
```

Response:

```json
{ "id": "...", "year": 2026, "target": 25 }
```

### `GET /me/goals/:year`  *(auth required)*

Returns the goal and progress (count of finished books with `date_read` in that year) for the current user.

```json
{ "id": "...", "year": 2026, "target": 25, "progress": 12 }
```

### `GET /users/:username/goals/:year`  *(optional auth)*

Public endpoint — returns goal + progress for a user. Respects privacy settings.

```json
{ "id": "...", "year": 2026, "target": 25, "progress": 12 }
```

---

## Labels (collections)

A "label" is a `collection` row. Default labels (`read_status` exclusive group) enforce mutual exclusivity — adding a book to one removes it from the others in the group. (Note: API endpoints still use `/shelves/` in their paths for backwards compatibility.)

### `GET /users/:username/shelves`

Returns all labels for a user (default + custom + tag collections), plus an aggregate `total_books` count (distinct books across all of the user's `user_books`).

```json
{
  "total_books": 137,
  "shelves": [
    {
      "id": "...",
      "name": "Read",
      "slug": "read",
      "exclusive_group": "read_status",
      "collection_type": "shelf",
      "item_count": 42,
      "description": "My finished books"
    }
  ]
}
```

`description` is only present when non-empty (max 1000 characters). `collection_type` is one of `"shelf"`, `"tag"`, or `"computed"`. See `docs/organization.md` for the distinction.

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

Returns a label with its full book list. Computed lists also include the `computed` object. `description` is only present when non-empty.

**Query params:** `sort` — one of `date_added` (default), `title`, `author`, `rating`.

```json
{
  "id": "...",
  "name": "Read",
  "slug": "read",
  "exclusive_group": "read_status",
  "description": "My finished books",
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

Same as `GET /users/:username/shelves` but for the authenticated user. Used by the label picker on book pages.

### `POST /me/shelves`  *(auth required)*

Create a custom label or tag collection.

```json
{
  "name": "Favorites",
  "is_exclusive": false,
  "exclusive_group": null,
  "is_public": true,
  "collection_type": "shelf",
  "description": "My all-time favorite books"
}
```

`description` is optional (max 1000 characters). Slug is auto-derived from `name`. Returns 409 on slug conflict.

### `PATCH /me/shelves/:id`  *(auth required)*

Rename, toggle visibility, or update description. Accepts `{ name?, is_public?, description? }`. `description` max 1000 characters; send an empty string to clear.

### `DELETE /me/shelves/:id`  *(auth required)*

Delete a custom label. Returns 403 if `exclusive_group = 'read_status'` (default labels cannot be deleted).

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

Same as above, but also saves the result as a new label. Accepts an additional `name` field. Returns `{ id, name, slug, book_count, is_continuous }`. If `is_continuous` is true, the list auto-refreshes when viewed — the operation is re-executed against the current source collections each time `GET /users/:username/shelves/:slug` is called.

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

Same as above, but also saves the result as a new label. Accepts an additional `name` field. Returns `{ id, name, slug, book_count, is_continuous }`. If `is_continuous` is true, the list auto-refreshes when viewed, resolving the other user's collection by stored username+slug on each view (respecting privacy).

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

Add a book to a label. Upserts the book into the global `books` catalog. For exclusive labels, removes the book from all other labels in the same `exclusive_group` for this user.

```json
{
  "open_library_id": "OL82592W",
  "title": "The Great Gatsby",
  "cover_url": "https://..."
}
```

### `PATCH /shelves/:shelfId/books/:olId`  *(auth required)*

Update review metadata on a book in a label. Only provided fields are updated — absent fields are not set to null.

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

Remove a book from a label.

---

## Tags (path-based)

Tags are `collection` rows with `collection_type = 'tag'`. Slugs can contain `/` to form a hierarchy. See `docs/organization.md` for full semantics.

### `GET /users/:username/tags/*path`

Returns books tagged with the given path or any sub-path.

**Query params:** `sort` — one of `date_added` (default), `title`, `author`, `rating`.

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

**Query params:** `sort` — one of `date_added` (default), `title`, `author`, `rating`.

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
- `shelf` *(optional)* — collection ID to export a single label. Omit to export all labels.

**CSV columns:** Title, Author, ISBN13, Status, Rating, Review, Date Added, Date Read, Date DNF.

---

## Saved Searches

### `GET /me/saved-searches`  *(auth required)*

Returns the user's saved searches, newest first. Max 20 per user.

```json
[
  {
    "id": "...",
    "name": "Sci-fi favorites",
    "query": "science fiction",
    "filters": {
      "sort": "rating",
      "subject": "science fiction",
      "language": "eng"
    },
    "created_at": "2026-02-28T14:00:00Z"
  }
]
```

`filters` may be null if no filters were active when the search was saved. Possible filter keys: `sort`, `year_min`, `year_max`, `subject`, `language`, `tab`.

### `POST /me/saved-searches`  *(auth required)*

Save a search query with optional filters. Max 20 per user.

```json
{
  "name": "Sci-fi favorites",
  "query": "science fiction",
  "filters": { "sort": "rating", "subject": "science fiction" }
}
```

```
201 { "id": "...", "name": "...", "query": "...", "filters": {...}, "created_at": "..." }
400 { "error": "name and query are required" }
400 { "error": "name must be 100 characters or fewer" }
400 { "error": "maximum of 20 saved searches reached" }
```

### `DELETE /me/saved-searches/:id`  *(auth required)*

Delete a saved search. Only the owner can delete.

```
200 { "ok": true }
403 { "error": "Not your saved search" }
404 { "error": "Saved search not found" }
```

---

## Import

### `POST /me/import/goodreads/preview`  *(auth required)*

Accepts a multipart form with a `file` field containing a Goodreads CSV export. Returns a preview without writing to the database.

Response groups rows into `matched`, `ambiguous`, and `unmatched`. The lookup chain tries: local DB by ISBN, OL direct ISBN endpoint, OL search by ISBN, OL search by cleaned title+author, OL search by title only, OL comma-subtitle retry, Google Books as a fallback, and finally LLM-powered fuzzy matching. The Google Books fallback searches by ISBN then by title+author, and maps results back to Open Library by re-searching OL with the title/author from Google. Set the optional `GOOGLE_BOOKS_API_KEY` env var for higher rate limits (free tier: 1,000 req/day); the fallback works without a key.

**LLM fuzzy matching:** When all standard lookups fail, the API calls the Anthropic API (Claude Haiku) to generate alternate title/author search permutations (correcting misspellings, removing series info, trying alternate titles, reversing author names, etc.) and retries Open Library searches with each permutation. If candidates are found, the row is marked `ambiguous` with up to 5 candidates for the user to choose from. Set the optional `ANTHROPIC_API_KEY` env var to enable this feature; without it, unmatched rows go directly to the `unmatched` state.

### `POST /me/import/goodreads/commit`  *(auth required)*

Accepts the confirmed preview payload and writes to the database. Returns `{ imported, failed, errors, pending_saved }`.

The request body includes an optional `unmatched_rows` array alongside `rows` and `shelf_mappings`. Unmatched rows are persisted to the `pending_imports` collection for later manual resolution.

**`shelf_mappings` actions:**
- `"tag"` — creates a standalone tag key per Goodreads shelf (default)
- `"create_label"` — groups the Goodreads shelf as a value under a new label key specified by `label_name`
- `"existing_label"` — adds the Goodreads shelf as a value under an existing label key specified by `label_key_id`
- `"skip"` — ignores the Goodreads shelf

### `POST /me/import/storygraph/preview`  *(auth required)*

Accepts a multipart form with a `file` field containing a StoryGraph CSV export. Returns a preview without writing to the database.

StoryGraph CSV columns: `Title`, `Authors`, `ISBN/UID`, `Format`, `Read Status`, `Star Rating`, `Review`, `Tags`, `Read Dates`. The lookup chain is the same as Goodreads import (including LLM fuzzy matching as a final fallback). Status mapping: `to-read` → `want-to-read`, `currently-reading` → `currently-reading`, `read` → `finished`, `did-not-finish` → `dnf`. Tags are imported as custom labels. Read Dates may be a range (`2024/01/15-2024/02/20`); the end date is used as `date_read`.

### `POST /me/import/storygraph/commit`  *(auth required)*

Accepts the confirmed preview payload and writes to the database. Returns `{ imported, failed, errors, pending_saved }`. Same request body format and `shelf_mappings` actions as the Goodreads commit endpoint. Unmatched rows are saved with `source: "storygraph"`.

### `POST /me/import/librarything/preview`  *(auth required)*

Accepts a multipart form with a `file` field containing a LibraryThing TSV export. Returns a preview without writing to the database.

LibraryThing TSV columns: `Title`, `Author (First, Last)`, `ISBN`, `ISBNs`, `Rating`, `Review`, `Date Read`, `Entry Date`, `Collections`, `Tags`. The export is tab-separated. Author names in "Last, First" format are reversed to "First Last". Collections and Tags are both imported as custom labels. Status mapping: "Currently Reading" → `currently-reading`, "To Read"/"Wishlist" → `to-read` (want-to-read), "Read but unowned" or books with a Date Read → `read` (finished). Ratings > 5 are normalized from a 10-point to a 5-point scale.

### `POST /me/import/librarything/commit`  *(auth required)*

Accepts the confirmed preview payload and writes to the database. Returns `{ imported, failed, errors, pending_saved }`. Same request body format and `shelf_mappings` actions as the Goodreads commit endpoint. Unmatched rows are saved with `source: "librarything"`.

---

## Pending Imports

### `GET /me/imports/pending`  *(auth required)*

Returns unmatched import rows saved from previous imports (Goodreads or StoryGraph). Only returns rows with status `unmatched`.

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

### `POST /me/imports/pending/:id/retry`  *(auth required)*

Re-runs the full lookup chain (local DB, Open Library ISBN, OL search, Google Books, LLM fuzzy) for a single pending import. If a match is found, auto-resolves (creates user_book, maps status tag) and removes the pending import.

```json
// Match found — auto-resolved
{ "status": "matched", "book_id": "...", "match": { "ol_id": "...", "title": "...", "authors": [...], "cover_url": "..." } }

// Ambiguous — candidates returned for user selection
{ "status": "ambiguous", "candidates": [{ "ol_id": "...", "title": "...", "authors": [...], "cover_url": "..." }] }

// No match
{ "status": "unmatched" }
```

---

## Activity Feed

### `GET /me/feed`  *(auth required)*

Returns a chronological feed of activities from users the authenticated user follows. Cursor-based pagination.

**Query parameters:**
- `cursor` *(optional)* — RFC3339Nano timestamp from `next_cursor` to fetch the next page.
- `type` *(optional)* — comma-separated list of activity types to filter by (e.g. `?type=reviewed,rated`). Valid types: `shelved`, `started_book`, `finished_book`, `rated`, `reviewed`, `created_thread`, `followed_user`, `followed_author`, `created_link`. Default (omitted) returns all types.

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

### `GET /users/:username/stats`  *(optional auth)*

Returns detailed reading statistics for a user. Respects privacy settings — returns 403 for private profiles if the viewer is not an approved follower.

```json
{
  "books_by_year": [
    { "year": 2026, "count": 12 },
    { "year": 2025, "count": 34 }
  ],
  "books_by_month": [
    { "year": 2026, "month": 1, "count": 5 },
    { "year": 2026, "month": 2, "count": 7 }
  ],
  "average_rating": 3.8,
  "rating_distribution": [
    { "rating": 1, "count": 2 },
    { "rating": 2, "count": 5 },
    { "rating": 3, "count": 12 },
    { "rating": 4, "count": 20 },
    { "rating": 5, "count": 8 }
  ],
  "total_books": 47,
  "total_reviews": 15,
  "total_pages_read": 14320
}
```

- `books_by_year` — finished books grouped by year (from `date_read`), descending
- `books_by_month` — finished books in the current year grouped by month
- `rating_distribution` — count of books per star rating (1-5)
- `total_pages_read` — sum of `page_count` across all finished books (only books with known page counts)

### `GET /users/:username/year-in-review?year=<YYYY>`  *(optional auth)*

Returns a year-in-review summary for a user. Defaults to the current year. Respects profile privacy settings.

```json
{
  "year": 2025,
  "total_books": 42,
  "total_pages": 12500,
  "average_rating": 3.8,
  "highest_rated": {
    "open_library_id": "OL82592W",
    "title": "The Great Gatsby",
    "cover_url": "https://...",
    "rating": 5
  },
  "longest_book": {
    "open_library_id": "OL27448W",
    "title": "The Lord of the Rings",
    "cover_url": "https://...",
    "page_count": 1200
  },
  "shortest_book": {
    "open_library_id": "OL12345W",
    "title": "Animal Farm",
    "cover_url": "https://...",
    "page_count": 112
  },
  "top_genres": [
    { "name": "Fiction", "count": 20 },
    { "name": "Fantasy", "count": 8 }
  ],
  "books_by_month": [
    {
      "month": 1,
      "count": 5,
      "books": [
        {
          "open_library_id": "OL82592W",
          "title": "The Great Gatsby",
          "cover_url": "https://...",
          "rating": 4
        }
      ]
    }
  ],
  "available_years": [2025, 2024, 2023]
}
```

- `highest_rated`, `longest_book`, `shortest_book` are null when no qualifying books exist
- `average_rating` is null when no books are rated
- `top_genres` derived from books' `subjects` field; top 5 by count
- `books_by_month` only includes months with books; each month includes book covers
- `available_years` lists all years the user has finished books (for year selector)

### `GET /users/:username/activity`

Returns recent activity for a specific user. Same response format as `/me/feed`.

**Query parameters:**
- `cursor` *(optional)* — RFC3339Nano timestamp from a previous `next_cursor` value. Only returns activities created before this timestamp.

Returns `{ "activities": [...], "next_cursor": "..." }`. `next_cursor` is `null` when there are no more results. Each page returns up to 30 items.

---

## Rate Limiting (Open Library)

All outbound requests to Open Library are routed through a shared rate-limited HTTP client (`api/internal/olhttp`). This uses a token-bucket algorithm (5 requests/second steady-state, burst of 15) to prevent the API from being banned by OL for excessive traffic.

Affected routes: `GET /books/search`, `GET /books/lookup`, `GET /books/:workId`, `GET /books/:workId/editions`, `GET /authors/search`, `GET /authors/:authorKey`, and `POST /me/import/goodreads/preview`.

When the rate limit is saturated, requests wait (up to the 15s client timeout) rather than failing immediately.

---

## Discussion Threads

### `GET /books/:workId/threads?page=1&limit=20`

Returns discussion threads for a book, ordered by most recent first. Paginated.

**Query parameters:**
- `page` *(optional, default 1)* — page number
- `limit` *(optional, default 20, max 100)* — threads per page

```json
{
  "threads": [
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
  ],
  "total": 42
}
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

Find existing threads on a book whose titles are similar to the given title. Uses trigram similarity (computed in Go) with a threshold of 0.3. Returns up to 5 results sorted by similarity score. Used by the thread creation form to suggest existing discussions before posting a duplicate.

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

## Book Quotes

Users can save quotes/highlights from books. Quotes can be public (visible to everyone on the book page) or private (visible only to the quote author).

### `GET /books/:workId/quotes?page=1`

Returns public quotes for a book, paginated (20 per page), ordered by newest first.

```json
[
  {
    "id": "...",
    "user_id": "...",
    "username": "alice",
    "display_name": "Alice",
    "avatar_url": "/api/files/users/.../avatar.jpg",
    "text": "So we beat on, boats against the current...",
    "page_number": 180,
    "note": "The famous closing line",
    "created_at": "2026-02-28T14:00:00Z"
  }
]
```

`page_number`, `note`, `display_name`, and `avatar_url` may be null.

### `GET /me/books/:olId/quotes`  *(auth required)*

Returns the authenticated user's quotes for a book (both public and private), ordered by newest first.

```json
[
  {
    "id": "...",
    "text": "So we beat on...",
    "page_number": 180,
    "note": "The famous closing line",
    "is_public": true,
    "created_at": "2026-02-28T14:00:00Z"
  }
]
```

### `POST /me/books/:olId/quotes`  *(auth required)*

Create a new quote for a book.

```json
{
  "text": "So we beat on, boats against the current...",
  "page_number": 180,
  "note": "The famous closing line",
  "is_public": true
}
```

`text` is required (max 2000 chars). `page_number`, `note` (max 500 chars), and `is_public` (default true) are optional.

```
200 { "id": "...", "text": "...", "created_at": "..." }
400 { "error": "text is required" }
400 { "error": "text must be 2000 characters or fewer" }
404 { "error": "Book not found" }
```

### `DELETE /me/quotes/:quoteId`  *(auth required)*

Delete a quote owned by the authenticated user. Returns 204.

```
204 (no content)
403 { "error": "Not your quote" }
404 { "error": "Quote not found" }
```

---

## Community Links

User-submitted book-to-book connections (sequel, prequel, similar, etc.). Links are upvotable — sorted by vote count on book pages. Both books must exist in the local catalog.

### `GET /books/:workId/links?limit=50&offset=0`

Returns community links for a book, sorted by creation date descending. If authenticated, includes whether the caller has voted on each link. Paginated.

**Query parameters:**
- `limit` *(optional, default 50, max 100)* — links per page
- `offset` *(optional, default 0)* — number of links to skip

```json
{
  "links": [
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
  ],
  "total": 12
}
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

Valid `link_type` values: `sequel`, `prequel`, `companion`, `mentioned_in`, `similar`, `adaptation`. Returns 400 with `"invalid link_type"` for any other value.

Returns 201 with `{ id, created_at }`. Auto-upvotes by the creator. Returns 409 if the user already submitted this exact link.

### `DELETE /links/:linkId`  *(auth required)*

Soft-delete a link (author only). Returns 204.

### `POST /links/:linkId/vote`  *(auth required)*

Upvote a link. Idempotent. Returns 204.

### `DELETE /links/:linkId/vote`  *(auth required)*

Remove upvote. Returns 204.

### `POST /links/:linkId/edits`  *(auth required)*

Propose an edit to a community link. At least one of `proposed_type` or `proposed_note` must be provided. Only one pending edit per user per link. If `proposed_type` is provided, it must be one of: `sequel`, `prequel`, `companion`, `mentioned_in`, `similar`, `adaptation`.

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
  "status": "approved",
  "reviewer_comment": "Looks good"
}
```

`status` must be `"approved"` or `"rejected"`. `reviewer_comment` is optional.

```
200 { "message": "Edit approved" }
400 { "error": "status must be approved or rejected" }
404 { "error": "Edit not found" }
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

## Reports

### `POST /reports`  *(auth required)*

Report a piece of content (review, thread, comment, or link). Prevents duplicate reports from the same user on the same content.

```json
{
  "content_type": "review",
  "content_id": "abc123",
  "reason": "spam",
  "details": "This review is advertising a product"
}
```

`content_type` is `"review"`, `"thread"`, `"comment"`, or `"link"`. `reason` is `"spam"`, `"harassment"`, `"inappropriate"`, or `"other"`. `details` is optional.

```
201 { "id": "...", "created_at": "..." }
400 { "error": "content_type must be review, thread, comment, or link" }
400 { "error": "reason must be spam, harassment, inappropriate, or other" }
409 { "error": "You have already reported this content" }
```

### `GET /admin/reports?status=pending`  *(moderator required)*

List content reports. Filterable by `status` (`pending`, `reviewed`, `dismissed`). Returns reports with reporter info and a content preview, sorted by newest first.

```json
[
  {
    "id": "...",
    "reporter_id": "...",
    "reporter_username": "alice",
    "reporter_display_name": "Alice",
    "content_type": "review",
    "content_id": "abc123",
    "reason": "spam",
    "details": "This review is advertising a product",
    "status": "pending",
    "reviewer_id": null,
    "reviewer_username": null,
    "created_at": "2026-02-26T14:00:00Z",
    "content_preview": "Buy my product at example.com..."
  }
]
```

### `PATCH /admin/reports/:reportId`  *(moderator required)*

Update a report's status to reviewed or dismissed. Sets the reviewer to the current moderator.

```json
{ "status": "reviewed" }
```

```
200 { "ok": true, "status": "reviewed" }
400 { "error": "status must be reviewed or dismissed" }
404 { "error": "Report not found" }
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

### `DELETE /me/notifications/:notifId`  *(auth required)*

Delete a single notification. Returns `{ "ok": true }`. Returns 404 if the notification doesn't exist, 403 if it doesn't belong to the current user.

### `POST /me/notifications/read-all`  *(auth required)*

Mark all unread notifications as read. Returns `{ "ok": true }`.

---

## Notification Preferences

### `GET /me/notification-preferences`  *(auth required)*

Returns the current user's notification preferences. If no preferences row exists, all types default to `true` (enabled).

```json
{
  "new_publication": true,
  "book_new_thread": true,
  "book_new_link": true,
  "book_new_review": true,
  "review_liked": true,
  "thread_mention": true,
  "book_recommendation": true,
  "review_comment": true
}
```

### `PUT /me/notification-preferences`  *(auth required)*

Upsert notification preferences. Only the fields provided in the body are updated; omitted fields keep their current value (or default to `true` if this is the first save).

```json
{ "review_liked": false, "thread_mention": false }
```

Returns the full updated preferences object (same shape as GET).

---

## Theme

### `GET /me/theme`  *(auth required)*

Returns the user's theme preference. Defaults to `"system"` if not set.

```json
{ "theme": "system" }
```

### `PUT /me/theme`  *(auth required)*

Set the user's theme preference. Valid values: `light`, `dark`, `system`.

```json
{ "theme": "dark" }
```

Returns the saved theme preference.

---

## Recommendations

### `POST /me/recommendations`  *(auth required)*

Send a book recommendation to another user.

```json
{ "username": "bob", "book_ol_id": "OL82592W", "note": "You'll love this!" }
```

Creates a recommendation record and sends a `book_recommendation` notification to the recipient. Also records a `sent_recommendation` activity. Returns `201` with the recommendation ID. Returns `409` if the same sender/recipient/book triple already exists. Rate-limited to 10 recommendations per 24-hour window — returns `429` with `"too many recommendations, try again later"` if exceeded.

### `GET /me/recommendations`  *(auth required)*

List received recommendations for the current user.

**Query parameters:**
- `status` *(optional, default: `pending`)* — filter by `pending`, `seen`, `dismissed`, or `all`.

Returns an array of recommendation objects, each including sender info (username, display_name, avatar_url) and book info (open_library_id, title, cover_url, authors).

### `GET /me/recommendations/sent`  *(auth required)*

List recommendations the current user has sent. Returns up to 50 results ordered by newest first.

```json
[
  {
    "id": "...",
    "note": "You'll love this!",
    "status": "pending",
    "created_at": "2026-02-25T14:00:00Z",
    "recipient": {
      "user_id": "...",
      "username": "bob",
      "display_name": "Bob",
      "avatar_url": "/api/files/users/..."
    },
    "book": {
      "open_library_id": "OL82592W",
      "title": "The Great Gatsby",
      "cover_url": "https://...",
      "authors": "F. Scott Fitzgerald"
    }
  }
]
```

### `PATCH /me/recommendations/:recId`  *(auth required)*

Update a recommendation's status. Only the recipient can update.

```json
{ "status": "seen" }
```

Valid statuses: `seen`, `dismissed`.

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
