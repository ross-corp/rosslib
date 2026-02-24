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
  - [x] Default shelves (Want to Read, Currently Reading, Read) shown on profile with item counts; cards link to shelf pages.
  - [x] Shelf pages at `/{username}/shelves/{slug}` — cover grid with title, owner can remove books inline.
  - [x] Library manager at `/{username}/shelves/{slug}` (owner only) — full-page layout with sidebar (shelves, custom, tags, labels), dense cover grid with multi-select checkboxes, and bulk action toolbar (Rate, Move to shelf, Labels, Remove). Non-owners see the original read-only grid.
  - [ ] Avatar upload.
  - [ ] Recent activity (reviews, threads, list updates) on profile.
  - [ ] Stats: books read (done), reviews written (needs reviews feature), followers/following count (done).
  - [ ] Profiles can be set to private; followers must be approved.

- [ ] Search & Discovery
  - [x] Search bar in nav — submits GET to /search.
  - [x] `/search` page — Books tab searches by title via Open Library API (up to 20 results with cover, authors, year). People tab searches users by username or display name.
  - [x] `/users` page — browse all users, alphabetical, paginated (20/page).
  - [x] Tab selector on `/search` to filter between Books and People.
  - [x] "Add to shelf" picker on each book search result — logged-in users can add/move/remove books across their 3 default shelves inline.
  - [ ] Author tab in search.
  - [ ] Full-text book/author search via Meilisearch (will replace Open Library as primary search backend).

- [ ] Social
  - [x] Follow / unfollow users (asymmetric). Follow button on profile page; `is_following` returned from profile endpoint.
  - [x] `follows` table with `(follower_id, followee_id)` PK and `status` field.
  - [ ] Private accounts require follow approval (status = 'pending').
  - [x] Followers / following counts on profile.
  - [x] "Friends" (mutual follows) surfaced in UI — friends_count (mutual follows) returned from profile endpoint and shown on profile page.
  - [ ] Follow authors, see new publications.
  - [ ] Follow works, see sequels / new discussions / links.
  - [ ] A way to represent users who are also authors
    - [ ] badges

## Data Model

- [ ] Collections
  - [x] 3 default collections for all users: Want to Read, Currently Reading, Read — created on registration (or lazily on first `/me/shelves` call for existing users).
  - [x] `books` table — global catalog keyed by `open_library_id`; upserted when a user adds a book to a shelf.
  - [x] `collections` + `collection_items` tables with `is_exclusive` / `exclusive_group` for mutual exclusivity enforcement.
  - [x] API: `GET /users/:username/shelves`, `GET /users/:username/shelves/:slug`, `GET /me/shelves`, `POST /shelves/:shelfId/books`, `DELETE /shelves/:shelfId/books/:olId`.
  - [x] Expand `books` table: add `isbn13 VARCHAR(13)`, `authors TEXT`, `publication_year INT` — needed for import matching and book pages.
  - [x] Expand `collection_items` table: add `rating SMALLINT` (0–5, 0 = unrated), `review_text TEXT`, `spoiler BOOLEAN DEFAULT false`, `date_read TIMESTAMPTZ`, `date_added TIMESTAMPTZ` — needed before import can store review/rating data.
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
  - [ ] Faceted filters: genre, published year range, language.
  - [ ] Results ranked by relevance, with popular books surfaced higher.
- [ ] Book pages
  - [ ] Metadata: title, author(s), cover, description, publisher, year, page count.
  - [ ] Aggregate stats: average rating, read count, want-to-read count.
  - [ ] User's own status (added to which collection, their rating/review).
  - [ ] Community reviews and discussion threads.
  - [ ] Community links to related works.
- [ ] Author page
- [ ] Genre pages

- [ ] Edition handling

## Reviews & Ratings

- [x] Rating and review text live on `collection_items` (one per user per book). Fields: `rating`, `review_text`, `spoiler`, `date_read`, `date_added`.
- [x] `PATCH /shelves/:shelfId/books/:olId` — partial update of rating, review_text, spoiler, date_read. Uses `map[string]json.RawMessage` to distinguish absent fields from explicit nulls; dynamically builds SET clause.
- [ ] A user can rate a book 1–5 stars (integers; half stars are a future enhancement).
- [ ] A rating alone (no review text) is valid.
- [ ] Review text is optional; can include a spoiler flag.
- [ ] One review per user per book; can be edited or deleted.
- [ ] `GET /users/:username/reviews` — list all reviews by a user (for profile "Reviews" tab).
- [ ] Reviews are shown on book pages sorted by recency and follower relationships (reviews from people you follow shown first).
- [x] Display star rating on shelf item cards (profile shelf pages).

## Discussion Threads

- [ ] Any user can open a thread on a book's page.
- [ ] Thread has a title, body, and optional spoiler flag.
- [ ] Threads get reccommended for union if they're similar enough
  - [ ] link to similar comments if similarity score > some percentage
- [ ] Threaded comments support one level of nesting (reply to a comment, not reply to a reply).
- [ ] No upvotes at MVP; chronological sort only.
- [ ] Author can delete their own thread or comments; soft delete.

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
    - [x] Map `Exclusive Shelf` column: `read` → Read, `currently-reading` → Currently Reading, `to-read` → Want to Read; any other value (e.g. `owned-to-read`, `dnf`) treated as a new custom exclusive shelf in the `read_status` group.
    - [x] Map `Bookshelves` column: each tag → proposed as a non-exclusive custom shelf, reusing existing ones by slug via `EnsureShelf`.
    - [x] Per-row payload includes: title, author, isbn13, matched OL ID, target shelf slugs, rating, review_text, date_read, date_added.
  - [x] `POST /me/import/goodreads/commit` — accepts the confirmed preview payload; writes to DB sequentially.
    - [x] Upserts books into `books` table (isbn13, authors, publication_year populated from OL data).
    - [x] Creates any new custom shelves (exclusive or non-exclusive) via `EnsureShelf`.
    - [x] Adds books to collections; sets `rating`, `review_text`, `spoiler`, `date_read`, `date_added` on `collection_items`. Removes book from other exclusive shelves in the same group.
    - [x] Skips rows the user excluded in the preview.
    - [x] Returns a summary: imported count, failed count, error list (null-safe in UI).
  - [x] Import UI at `/settings/import` (`app/settings/import/page.tsx`):
    - [x] File picker for `.csv` upload.
    - [x] Calls preview endpoint; shows three groups: Matched (collapsed), Ambiguous (expanded, edition dropdown), Unmatched (info-only, no checkboxes).
    - [x] User can choose an edition for ambiguous rows or uncheck matched rows to skip them.
    - [x] "Import N books" confirm button calls commit endpoint.
    - [x] Shows import summary (imported/failed counts, expandable error list) on completion.
    - [x] Unmatched books saved to `localStorage` (`rosslib:import:unmatched`) after commit; shown as a persistent "Not found" panel on the idle screen with per-book search links and dismiss controls.
- [ ]  CSV Export
  - [ ] Export any collection (or all collections) to CSV.
  - [ ] Columns: title, author, ISBN, date added, rating, review, collection name.
  - [ ] Generated server-side and made available via a pre-signed S3 URL.
- [ ] kindle integration

## Feed

- [ ] Chronological feed of activity from users you follow.
- [ ] Activity types surfaced: added to collection, wrote a review, started/finished a book, created a thread, submitted a link, followed a new user.
- [ ] No algorithmic ranking at MVP; pure chronological.
- [ ] Paginated (cursor-based).
