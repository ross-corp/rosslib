# COMPLETED

Items completed and moved out of todo.md.

## Profile
- [x] Public profile at `/{username}` with display name, byline, member since; edit at `/settings`.
- [x] Status labels (Want to Read, Owned to Read, Currently Reading, Finished, DNF) on profile via `user_books` + `book_tag_values`.
- [x] Shelf pages at `/{username}/shelves/{slug}` — cover grid, inline remove for owners.
- [x] Library manager (owner only) — sidebar with status/tags/labels, dense grid, multi-select, bulk actions (Rate, Change status, Labels, Tags, Remove).
- [x] Label value navigation + nested labels (`/` hierarchy) in library manager sidebar and public label pages.
- [x] Avatar upload via MinIO (`POST /me/avatar`).
- [x] Recent activity, stats (books read, reviews, followers/following, books this year, avg rating), and recent reviews on profile.
- [x] Private profiles with follow approval.
- [x] Two-column profile redesign (main + activity sidebar).
- [x] Currently Reading spotlight and default Favorites tag on profile.
- [x] Tabbed shelf browser on profile.

## Search & Discovery
- [x] Search bar in nav, `/search` page with Books and People tabs.
- [x] `/users` page — browse all users, alphabetical, paginated.
- [x] StatusPicker on book search results for inline status changes.
- [x] Author tab in search (Open Library author search API).
- [x] Full-text book search via Meilisearch — local catalog indexed on startup/upsert, concurrent Meilisearch + Open Library queries, deduplicated by OL work ID.

## User Accounts
- [x] Registration & login — username + email + password (bcrypt), 30-day JWT httpOnly cookie.

## Social
- [x] Follow/unfollow (asymmetric), `follows` table with `(follower_id, followee_id)` PK + status.
- [x] Private accounts require follow approval (status = 'pending').
- [x] Followers/following counts + mutual-follow friends count on profile.

## Collections
- [x] `user_books` table for per-user book ownership (rating, review, dates); status via label system.
- [x] `books` table keyed by `open_library_id`; upserted on first add. Expanded with isbn13, authors, publication_year.
- [x] `collections` + `collection_items` retained for tag collections.
- [x] Full user_books API: POST, PATCH, DELETE, status endpoints, status-map, user books list with status filter.
- [x] `GET /users/:username/reviews` with `?limit=N`; profile stats (books_this_year, average_rating).
- [x] Custom collections — CRUD endpoints, non-exclusive by default, `EnsureShelf` helper, shown on profile/shelf pages.

## Book Search
- [x] Book title search via Open Library API.
- [x] ISBN-direct lookup via Open Library with local upsert.
- [x] Published year range filter (`year_min`/`year_max`).
- [x] Popularity-blended default ranking (OL read count, rating quality, edition count).
- [x] Faceted filters: genre (subject) and language. Predefined genre chips (Fiction, Fantasy, Science fiction, Mystery, Romance, Horror, Thriller, Biography, History, Poetry, etc.) and language pills (English, Spanish, French, German, etc.). Filters apply to both Meilisearch local catalog and Open Library external search. Subjects stored on `books.subjects` column and indexed in Meilisearch; language is pass-through to OL. "Clear all filters" link when any filter is active.
- [x] Genre pages: `/genres` index with card grid showing 12 predefined genres and local book counts; `/genres/:slug` detail page with paginated book list, StatusPicker for logged-in users, breadcrumb nav. API endpoints `GET /genres` (genre list with DB counts) and `GET /genres/:slug/books?page=&limit=` (Meilisearch browse with subject filter, DB fallback). Genres link in nav bar.

## Book Pages
- [x] Metadata, cover, description, aggregate rating from Open Library.
- [x] User's own status/rating/review with StatusPicker.
- [x] Community reviews from local DB with spoiler gating.
- [x] Publisher, year, page count from OL editions API.
- [x] Read count / want-to-read count from local DB.

## Author Pages
- [x] `/authors/:authorKey` — bio, dates, photo, external links, works grid with covers.

## Reviews & Ratings
- [x] Rating (1-5 stars) and review text on `user_books`; partial update via `PATCH /me/books/:olId`.
- [x] Rating alone valid; review text optional with spoiler flag; one per user per book, editable/deletable.
- [x] Reviews page at `/{username}/reviews`; star ratings on shelf item cards.
- [x] Reviews on book pages sorted by recency + follower relationships.

## Discussion Threads
- [x] Any user can open a thread on a book (title, body, optional spoiler flag).
- [x] One level of comment nesting; chronological sort; soft delete by author.

## Goodreads Import
- [x] Preview + commit endpoints; 5-worker goroutine pool for OL lookups.
- [x] CSV parsing (ISBN formula wrapper, quoted multi-line reviews), ISBN13 → title+author fallback.
- [x] Status mapping (read → Finished, etc.), bookshelves → custom shelf creation via `EnsureShelf`.
- [x] Import UI at `/settings/import` — file picker, grouped preview (matched/ambiguous/unmatched), edition dropdown, confirm, summary, unmatched persistence in localStorage.

## CSV Export
- [x] `GET /me/export/csv` — streams CSV with optional shelf filter. Export UI at `/settings/export`.

## Reading Progress
- [x] Update progress on a book (page number or %). Stored as `progress_pages` + `progress_percent` on `user_books`. Progress bar on book detail page when "Currently Reading"; progress bars under covers on profile.
- [x] Device page mapping — set custom page count for your edition/device (`device_total_pages` on `user_books`). Page-based progress updates calculate % from the custom total. Falls back to catalog `page_count` when unset.

## DNF Date
- [x] DNF date (`date_dnf` on `user_books`) — when a user marks a book as DNF, they can record the date they stopped reading. Mirrors `date_read` for Finished books. Exposed via `PATCH /me/books/:olId`, returned from status and review endpoints, displayed on book pages and review listings, included in CSV export.

## Feed
- [x] Chronological feed of followed-user activity (add to collection, review, rate, thread, follow, started/finished book).
- [x] Cursor-based pagination; no algorithmic ranking.

## Edition Handling
- [x] Edition listing on book detail pages. `GET /books/:workId` now returns `editions` array (up to 50) and `edition_count` from Open Library. Each edition includes key, title, publisher, publish date, page count, ISBN, cover, format, and language.
- [x] Standalone `GET /books/:workId/editions?limit=50&offset=0` endpoint for paginated edition browsing.
- [x] Editions section on book detail page with collapsible list — shows 5 by default, expandable to all loaded, with "load more" for works with 50+ editions. Each edition card shows cover thumbnail, title, format badge, publisher, date, page count, language, and ISBN.

## API Docs
- [x] OpenAPI 3.0 spec (`api/internal/docs/openapi.yaml`) documenting all ~60 routes with request/response schemas. Served at `GET /docs/openapi.yaml`. Interactive Swagger UI at `GET /docs` (CDN-hosted Swagger UI bundle). Spec embedded in the binary via `go:embed`.

## API Rate Limiting
- [x] Rate-limit upstream-proxied routes to avoid getting banned from Open Library. All outbound OL requests share a single rate-limited HTTP client (`api/internal/olhttp`) using a token-bucket algorithm (5 rps, burst 15). Covers book search, ISBN lookup, book detail, editions, author search, author detail, and Goodreads import preview. Requests queue when the limit is hit rather than failing.

## Community Links
- [x] Community-submitted book-to-book connections on book detail pages. Link types: `sequel`, `prequel`, `companion`, `mentioned_in`, `similar`, `adaptation`. Optional note per link. Upvotable (sorted by votes on book pages). Soft-deleted by link author. DB tables: `book_links` + `book_link_votes`. API: `GET/POST /books/:workId/links`, `DELETE /links/:linkId`, `POST/DELETE /links/:linkId/vote`. Frontend: `BookLinkList` client component with grouped display, inline add-link form, and upvote toggle.
