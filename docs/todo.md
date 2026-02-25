# Features

Backlog of all things we need #todo

Once we're further along we'll move to GH projects. this is fine for now

## UNSORTED

- [ ] map page lengths to my device
- [ ] bookshelf picture scan

## Webapp

- [ ] User Accounts
  - [x] Registration & Login
  - [x] Register with username + email + password (bcrypt hashed, stored in `users` table).
  - [x] Login with email + password, returns a 30-day JWT set as an httpOnly cookie.
  - [ ] Email verification required before full access.
  - [ ] Password reset via email link.
  - [ ] OAuth via Google.
- [ ] Profile
  - [x] Public profile page at `/{username}` — display name, byline, member since.
  - [x] Edit profile page at `/settings` — set display name and byline.
  - [x] Status labels (Want to Read, Owned to Read, Currently Reading, Finished, DNF) shown on profile; `user_books` table tracks user-book relationships (rating, review, dates), status tracked via label system (`book_tag_values`).
  - [x] Shelf pages at `/{username}/shelves/{slug}` — cover grid with title, owner can remove books inline. Status slugs fetch from `user_books`; tag slugs use existing shelf behavior.
  - [x] Library manager at `/{username}/shelves/{slug}` (owner only) — full-page layout with sidebar (status values, tags, labels), dense cover grid with multi-select checkboxes, and bulk action toolbar (Rate, Change status, Labels, Tags, Remove). Non-owners see the original read-only grid.
  - [x] Label value navigation in library manager sidebar — clicking a label value fetches and displays all books with that key+value assignment (including nested sub-values) via `GET /users/:username/labels/:keySlug/*valuePath`. Nested values are indented in the sidebar by depth.
  - [x] Nested labels — label values can contain `/` to form a hierarchy (e.g. `genre: History/Engineering`). Viewing a parent label (`genre: History`) includes books tagged at any depth. Public label pages at `/{username}/labels/:keySlug/*valuePath` support breadcrumb navigation and sub-label drill-down, matching the nested tags page behaviour.
  - [x] Avatar upload — `POST /me/avatar` (multipart); stored in MinIO; URL written to `users.avatar_url`. Swap `MINIO_*` env vars for S3 in production.
  - [x] Recent activity (reviews, threads, list updates) on profile.
  - [x] Stats: books read, reviews written, followers/following count shown on profile.
  - [x] Profiles can be set to private; followers must be approved.
  - [x] Two-column profile redesign — main content (2/3) + activity sidebar (1/3) on desktop; single column on mobile.
  - [x] Currently Reading spotlight with large cover row on profile.
  - [x] Default "Favorites" tag auto-created for every user (idempotent), shown as cover row on profile.
  - [x] Reading stats section on profile: books this year, average rating, total read, reviews count.
  - [x] Recent reviews section on profile with cover, rating, spoiler-gated snippet.
  - [x] Tabbed shelf browser on profile — click shelf tabs to browse books inline without navigating away.

- [ ] Search & Discovery
  - [x] Search bar in nav — submits GET to /search.
  - [x] `/search` page — Books tab searches by title via Open Library API (up to 20 results with cover, authors, year). People tab searches users by username or display name.
  - [x] `/users` page — browse all users, alphabetical, paginated (20/page).
  - [x] Tab selector on `/search` to filter between Books and People.
  - [x] StatusPicker on each book search result — logged-in users can add books to library with a status, change status, or remove inline.
  - [x] Author tab in search — searches Open Library's author search API; results show name, dates, top work, work count, and subjects. Clicking an author searches their books.
  - [x] Full-text book search via Meilisearch — local catalog is indexed into Meilisearch on startup and on book upsert. `/books/search` queries both Meilisearch (local results first) and Open Library (discovery) concurrently, deduplicating by OL work ID.

- [ ] Social
  - [x] Follow / unfollow users (asymmetric). Follow button on profile page; `is_following` returned from profile endpoint.
  - [x] `follows` table with `(follower_id, followee_id)` PK and `status` field.
  - [x] Private accounts require follow approval (status = 'pending').
  - [x] Followers / following counts on profile.
  - [x] "Friends" (mutual follows) surfaced in UI — friends_count (mutual follows) returned from profile endpoint and shown on profile page.
  - [ ] Follow authors, see new publications.
  - [ ] Follow works, see sequels / new discussions / links.
  - [ ] A way to represent users who are also authors
    - [ ] badges

## Data Model

- [ ] Collections
  - [x] `user_books` table — per-user book ownership with rating, review_text, spoiler, date_read, date_added. Replaces `collection_items` for user-book metadata.
  - [x] Status tracked via label system (`tag_keys` where slug='status', `book_tag_values`) — select_one key with values: Want to Read, Owned to Read, Currently Reading, Finished, DNF.
  - [x] `books` table — global catalog keyed by `open_library_id`; upserted when a user adds a book.
  - [x] `collections` + `collection_items` tables retained for tag collections (Favorites, custom tags). Old read_status shelves migrated to user_books + status labels.
  - [x] API: `POST /me/books`, `PATCH /me/books/:olId`, `DELETE /me/books/:olId`, `GET /me/books/:olId/status`, `GET /me/books/status-map`, `GET /users/:username/books` (with `?status=slug` filter).
  - [x] `GET /users/:username/reviews` supports `?limit=N` for preview on profile page.
  - [x] `GET /users/:username` returns `books_this_year` and `average_rating` in profile response (stats from `user_books` + `book_tag_values`).
  - [x] Expand `books` table: add `isbn13 VARCHAR(13)`, `authors TEXT`, `publication_year INT` — needed for import matching and book pages.
  - [x] custom collections
    - [x] `POST /me/shelves` — create a custom collection (name, slug auto-derived, is_exclusive, exclusive_group, is_public). Returns 409 on slug conflict.
    - [x] `PATCH /me/shelves/:id` — rename or change visibility of a custom collection.
    - [x] `DELETE /me/shelves/:id` — delete a custom collection (403 if `exclusive_group = 'read_status'`).
    - [x] Non-exclusive by default (a book can appear in multiple custom collections).
    - [x] `EnsureShelf` package-level helper: get-or-create by slug using `ON CONFLICT DO UPDATE SET name = collections.name RETURNING id` — used by the import pipeline.
    - [x] Show custom shelves on profile and on shelf pages.
  - [ ] Computed collections
    - [ ] Union: books in list A or list B.
    - [ ] Intersection: books in both list A and list B.
    - [ ] Difference: books in list A but not list B.
    - [ ] compute an operation + save as new collection
      - [ ] Example: "Books I've read that are also in my friend's Want to Read list."
    - [ ] enable continuous v. one-time computed collections

- [ ] Sublists / Hierarchical Tags
  - [ ] A collection can have sub-labels that form a hierarchy:
    - [ ] Example: a "Science Fiction" collection with sub-labels "Space Opera", "Hard SF", "Cyberpunk".
    - [ ] Sub-labels are tags on `CollectionItem`, not separate collections.
    - [ ] Display as nested groupings on the collection page.

## Connection to Book DBs

- [ ] Search
  - [x] Book title search via Open Library API (`GET /books/search?q=<title>`). Returns title, authors, cover image, first publish year, ISBNs.
  - [ ] Author search.
  - [x] ISBN-direct lookup via Open Library — `GET /books/lookup?isbn=<isbn>` endpoint that searches OL by ISBN, upserts into the local `books` table (using bare OL ID e.g. `OL82592W`), and returns the book. Used as primary lookup during Goodreads import; falls back to title+author search when no ISBN match.
  - [ ] Faceted filters: genre, language.
  - [x] Published year range filter — `year_min` / `year_max` query params on `GET /books/search`; Meilisearch + Open Library filtered concurrently; year range inputs on search page.
  - [ ] Results ranked by relevance, with popular books surfaced higher.
- [ ] Book pages
  - [x] Metadata: title, author(s), cover, description (from Open Library).
  - [x] Aggregate stats: average rating (from Open Library).
  - [x] User's own status — status label, rating, review shown on book page; StatusPicker to add/change status.
  - [x] Community reviews — `GET /books/:workId/reviews` returns all user reviews from the local DB; shown on book page with spoiler gating.
  - [x] Publisher, year, page count — fetched from Open Library editions API (`/works/{id}/editions.json`); displayed on book page below authors. Publication year also pulled from local DB when available.
  - [x] Read count / want-to-read count from local DB — shown on book page.
  - [ ] Community links to related works.
- [x] Author page — `/authors/:authorKey` fetches from Open Library (`/authors/{key}.json` + `/authors/{key}/works.json`); displays bio, dates, photo, external links, and a grid of works with covers linking to book pages. Author search results now link to author pages.
- [ ] Genre pages

- [ ] Edition handling

## Reviews & Ratings

- [x] Rating and review text live on `user_books` (one per user per book). Fields: `rating`, `review_text`, `spoiler`, `date_read`, `date_added`.
- [x] `PATCH /me/books/:olId` — partial update of rating, review_text, spoiler, date_read. Uses `map[string]json.RawMessage` to distinguish absent fields from explicit nulls; dynamically builds SET clause.
- [x] A user can rate a book 1–5 stars (integers; half stars are a future enhancement).
- [x] A rating alone (no review text) is valid.
- [x] Review text is optional; can include a spoiler flag.
- [x] One review per user per book; can be edited or deleted.
- [x] `GET /users/:username/reviews` — list all reviews by a user. Page at `/{username}/reviews` with cover, rating, spoiler-gated text, and date read.
- [x] Reviews are shown on book pages sorted by recency and follower relationships (reviews from people you follow shown first).
- [x] Display star rating on shelf item cards (profile shelf pages).

## Discussion Threads

- [x] Any user can open a thread on a book's page.
- [x] Thread has a title, body, and optional spoiler flag.
- [ ] Threads get reccommended for union if they're similar enough
  - [ ] link to similar comments if similarity score > some percentage
- [x] Threaded comments support one level of nesting (reply to a comment, not reply to a reply).
- [x] No upvotes at MVP; chronological sort only.
- [x] Author can delete their own thread or comments; soft delete.

## Community Links (Wiki)

- [ ] Link types: `sequel`, `prequel`, `companion`, `mentioned_in`, `similar`, `adaptation`.
- [ ] Optional note explaining the connection.
- [ ] Links are upvotable; sorted by upvotes on book pages.
- [ ] Soft-deleted by moderators if spam or incorrect.
- [ ] Future: edit queue similar to book metadata edits.

## Import / Export

- [x] Goodreads Import
  - [x] `POST /me/import/goodreads/preview` — accepts multipart CSV upload; returns JSON preview (no DB writes). 5-worker goroutine pool for concurrent OL lookups (up to ~30s for large imports).
    - [x] Parse Goodreads CSV format: strip `=""...""` Excel formula wrapper from ISBN fields; handle quoted multi-line review text.
    - [x] For each row: try ISBN13 lookup first (`LookupBookByISBN` with nil pool = no DB write), then fall back to title + author OL search.
    - [x] Categorise results as `matched` (single OL match), `ambiguous` (multiple candidates), `unmatched` (no match).
    - [x] Map `Exclusive Shelf` column: `read` → Finished, `currently-reading` → Currently Reading, `to-read` → Want to Read, `dnf` → DNF, `owned-to-read` → Owned to Read.
    - [x] Map `Bookshelves` column: each tag → proposed as a non-exclusive custom shelf, reusing existing ones by slug via `EnsureShelf`.
    - [x] Per-row payload includes: title, author, isbn13, matched OL ID, target shelf slugs, rating, review_text, date_read, date_added.
  - [x] `POST /me/import/goodreads/commit` — accepts the confirmed preview payload; writes to DB sequentially.
    - [x] Upserts books into `books` table (isbn13, authors, publication_year populated from OL data).
    - [x] Writes `user_books` rows + sets Status labels via `book_tag_values`. Ensures Status tag key exists via `tags.EnsureStatusLabel`.
    - [x] Skips rows the user excluded in the preview.
    - [x] Returns a summary: imported count, failed count, error list (null-safe in UI).
  - [x] Import UI at `/settings/import` (`app/settings/import/page.tsx`):
    - [x] File picker for `.csv` upload.
    - [x] Calls preview endpoint; shows three groups: Matched (collapsed), Ambiguous (expanded, edition dropdown), Unmatched (info-only, no checkboxes).
    - [x] User can choose an edition for ambiguous rows or uncheck matched rows to skip them.
    - [x] "Import N books" confirm button calls commit endpoint.
    - [x] Shows import summary (imported/failed counts, expandable error list) on completion.
    - [x] Unmatched books saved to `localStorage` (`rosslib:import:unmatched`) after commit; shown as a persistent "Not found" panel on the idle screen with per-book search links and dismiss controls.
- [x]  CSV Export
  - [x] Export any collection (or all collections) to CSV.
  - [x] Columns: title, author, ISBN13, date added, date read, rating, review, collection name.
  - [x] `GET /me/export/csv` — streams CSV with optional `?shelf=<id>` filter. Served as direct download (Content-Disposition attachment).
  - [x] Export UI at `/settings/export` — shelf selector dropdown and download button.
- [ ] kindle integration

## Feed

- [x] Chronological feed of activity from users you follow.
- [x] Activity types surfaced: added to collection, wrote a review, rated a book, created a thread, followed a new user.
- [x] No algorithmic ranking at MVP; pure chronological.
- [x] Paginated (cursor-based).
- [ ] Activity type: submitted a link (requires community links feature).
- [x] Activity type: started/finished a book (requires tracking shelf transitions).
