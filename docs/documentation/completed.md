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
- [x] Moderator soft-delete for community links. Added `is_moderator` boolean column on `users` table (default false). `is_moderator` is included in JWT claims and extracted by auth middleware. `DELETE /links/:linkId` now allows both the link author and moderators to soft-delete any link. Frontend shows delete button for own links and for moderators (with distinct tooltip). Moderator status is set directly in the DB (no admin UI yet).

## Admin
- [x] Admin UI to grant/revoke moderator status. Added `RequireModerator()` middleware that gates all `/admin/*` routes (403 for non-moderators). API endpoints: `GET /admin/users` (paginated user list with search by username/display name/email, includes `is_moderator` flag) and `PUT /admin/users/:userId/moderator` (set `is_moderator` true/false). Existing ghost admin routes (`/admin/ghosts/*`) now also require moderator access. Frontend: `/admin` page with searchable user table and inline moderator toggle buttons. "Admin" nav link visible only to moderators. Changes take effect on the target user's next login (JWT re-issue required).

## Reviews
- [x] Wikilinks to other books in review text. `ReviewText` component (`components/review-text.tsx`) parses `[[Book Title]]` wikilinks (linking to search) and `[Title](/books/OLID)` markdown links (direct book page links). `BookReviewEditor` provides `[[` autocomplete that searches books via the API and inserts markdown links. `ReviewText` is now used consistently across all review display locations: book detail page community reviews, user reviews page (`/[username]/reviews`), and recent reviews on profiles.

## Feed
- [x] Activity type: submitted a link. When a user submits a community link between two books, a `created_link` activity is recorded and shown in the feed. Displays as "submitted a [type] link on [from_book] to [to_book]" with links to both book pages. Metadata includes `link_type`, `to_book_ol_id`, and `to_book_title`. OpenAPI spec updated.

## Social — Author Follows
- [x] Follow/unfollow authors on author pages. DB table `author_follows` keyed by `(user_id, author_key)` stores followed authors with cached `author_name`. API endpoints: `POST /authors/:authorKey/follow` (follow, records `followed_author` activity with metadata), `DELETE /authors/:authorKey/follow` (unfollow), `GET /authors/:authorKey/follow` (check status), `GET /me/followed-authors` (list followed authors). Frontend: `AuthorFollowButton` client component on author detail pages (visible to logged-in users). Feed displays "followed author [Name]" with link to author page. OpenAPI spec updated with all new endpoints and `followed_author` activity type.

## Notifications — New Author Publications
- [x] New publication notifications for followed authors. Background poller (`notifications.StartPoller`) runs every 6 hours, checking Open Library for work count changes on all followed authors. DB tables: `author_works_snapshot` (tracks known work counts per author key) and `notifications` (per-user notification rows with `notif_type`, `title`, `body`, JSONB `metadata`, `read` flag). On first poll, snapshots are seeded without generating notifications; subsequent polls create `new_publication` notifications fanned out to all followers when an author's work count increases. Newest work titles are included in the notification body. API endpoints: `GET /me/notifications` (paginated, cursor-based), `GET /me/notifications/unread-count`, `POST /me/notifications/:notifId/read`, `POST /me/notifications/read-all`. Frontend: `NotificationBell` client component in nav bar polls unread count every 60s and shows a red badge; `/notifications` page lists all notifications with unread dot indicator, "Mark all read" button, and author page links. OpenAPI spec updated with all new endpoints and `Notification` schema.

## Social — Book Follows
- [x] Follow/unfollow books on book detail pages. DB table `book_follows` keyed by `(user_id, book_id)` stores subscriptions. API endpoints: `POST /books/:workId/follow` (follow, records `followed_book` activity), `DELETE /books/:workId/follow` (unfollow), `GET /books/:workId/follow` (check status), `GET /me/followed-books` (list followed books with metadata). Notifications: when a new thread, community link, or review is posted on a followed book, all followers (except the actor) receive a notification (`book_new_thread`, `book_new_link`, `book_new_review` types). Notifications are fire-and-forget via goroutines. Frontend: `BookFollowButton` client component on book detail pages (visible to logged-in users). Notification page shows "View book" links for book-related notifications. Feed displays "followed [Book Title]" with link to book page. OpenAPI spec updated with all new endpoints and `followed_book` activity type.

## Community Link Edit Queue
- [x] Edit queue for community links. Any authenticated user can propose edits (type and/or note changes) to existing book-to-book links via a pencil icon on book pages. Proposals are stored in `book_link_edits` table with `pending`/`approved`/`rejected` status. One pending edit per user per link. Moderators review edits from the admin panel (`/admin`) with side-by-side current vs. proposed values, approve/reject buttons, and status filter tabs. Approved edits are applied within a DB transaction. API: `POST /links/:linkId/edits` (propose), `GET /admin/link-edits?status=` (list), `PUT /admin/link-edits/:editId` (review). Frontend: inline edit form in `BookLinkList`, `AdminLinkEdits` component on admin page.

## User Accounts — Set/Change Password
- [x] Allow Google OAuth users to set a password (link accounts both ways). API endpoints: `GET /me/account` returns `{has_password, has_google}` to indicate account status, `PUT /me/password` accepts `{new_password, current_password?}` — if the user already has a password, `current_password` is required and verified via bcrypt; OAuth-only users only need `new_password` (min 8 chars). New password is bcrypt-hashed and stored in `password_hash`. Webapp: proxy routes `GET /api/me/account` and `PUT /api/me/password`. `PasswordForm` client component on the settings page fetches account info on mount, shows "Set password" form for OAuth-only users or "Change password" form (with current password field) for users who already have one. Includes confirm-password field with client-side match validation. OpenAPI spec updated with both new endpoints.

## User Accounts — Password Reset
- [x] Password reset via email link. DB table `password_reset_tokens` stores SHA-256 hashed tokens with 1-hour expiry and single-use flag. API endpoints: `POST /auth/forgot-password` (accepts email, generates token, sends reset email via SMTP — always returns 200 to avoid leaking account info) and `POST /auth/reset-password` (validates token, sets new password, marks token as used in a transaction). Previous unused tokens are invalidated when a new one is requested. Email sending uses Go's `net/smtp` via a new `email` package (`api/internal/email`). SMTP configuration is optional — when `SMTP_HOST` is not set, reset requests are logged but no email is sent. Webapp: `/forgot-password` page with email form and success message, `/reset-password?token=...` page with new password + confirm password form. Next.js proxy routes: `POST /api/auth/forgot-password` and `POST /api/auth/reset-password`. Environment variables: `SMTP_HOST`, `SMTP_PORT` (default 587), `SMTP_USER`, `SMTP_PASSWORD`, `SMTP_FROM`, `WEBAPP_URL` (default `http://localhost:3000`). OpenAPI spec updated. Login page already linked to `/forgot-password`.

## Social — Author Badges
- [x] Represent users who are also authors with profile badges. DB change: `author_key` column on `users` (varchar(50), nullable) stores an Open Library author ID (e.g. `OL23919A`). API: `GET /users/:username` profile response includes `author_key`; admin endpoint `PUT /admin/users/:userId/author` sets or clears the author key. Admin user list (`GET /admin/users`) includes `author_key` in the response. Webapp: profile page shows an amber "Author" badge (with book icon) next to the display name when `author_key` is set, linking to `/authors/:authorKey`. Admin user list has inline author key editor — click "Set author" to enter an OL author key, or click the existing badge to edit/clear it. Webapp proxy route: `PUT /api/admin/users/:userId/author`.

## Genre Ratings
- [x] Rate a book on genre dimensions (0–10 scale). DB table `genre_ratings` keyed by `(user_id, book_id, genre)` with CHECK constraint on rating 0–10. 12 predefined genres matching the existing genre browse list: Fiction, Non-fiction, Fantasy, Science fiction, Mystery, Romance, Horror, Thriller, Biography, History, Poetry, Children. API: `GET /books/:workId/genre-ratings` (public aggregate: average + rater count per genre), `GET /me/books/:olId/genre-ratings` (user's own ratings), `PUT /me/books/:olId/genre-ratings` (upsert/delete batch). Frontend: `GenreRatingEditor` client component on book detail pages shows aggregate ratings as horizontal bar charts and provides 0–10 slider editor for logged-in users. Webapp proxy routes: `GET /api/books/:workId/genre-ratings`, `GET/PUT /api/me/books/:olId/genre-ratings`. OpenAPI spec updated.

## Computed Lists — Set Operations
- [x] Set operations on collections (union, intersection, difference). API endpoints: `POST /me/shelves/set-operation` (compute union/intersection/difference between two of the user's collections, returns result book list) and `POST /me/shelves/set-operation/save` (compute and save result as a new shelf). Both endpoints verify collection ownership. Union uses `DISTINCT ON` across both collections; intersection joins `collection_items` on matching `book_id`; difference uses `NOT EXISTS` subquery. Frontend: `/library/compare` page with `SetOperationForm` client component — two collection dropdowns, operation selector with descriptions, compare button, result grid (reuses existing book cover grid pattern with ratings), and "Save as new list" form that creates a new shelf from the result. "Compare lists" link added to settings page. Webapp proxy routes: `POST /api/me/shelves/set-operation`, `POST /api/me/shelves/set-operation/save`. OpenAPI spec updated.

## Computed Lists — Cross-User Set Operations
- [x] Cross-user set operations (compare your list with a friend's). API endpoints: `POST /me/shelves/cross-user-compare` (compute union/intersection/difference between one of your collections and another user's public collection, identified by username + shelf slug) and `POST /me/shelves/cross-user-compare/save` (compute and save result as a new shelf). Both endpoints resolve the other user's collection by username and slug, respecting profile privacy (returns 403 for private profiles the viewer doesn't follow). Uses `privacy.CanViewProfile` for access control. Frontend: `/library/compare` page updated with `CompareTabs` component providing "My Lists" (existing) and "Compare with a Friend" tabs. `CrossUserCompareForm` client component: select your list, enter a friend's username, load their shelves, pick their list, choose operation, compare. Results grid with book covers and ratings, "Save as new list" option. Webapp proxy routes: `POST /api/me/shelves/cross-user-compare`, `POST /api/me/shelves/cross-user-compare/save`, `GET /api/users/:username/shelves` (new proxy for client-side shelf fetching). OpenAPI spec updated.

## Computed Lists — Continuous (Live) Lists
- [x] Continuous vs. one-time computed lists. DB table `computed_collections` stores the operation definition (operation type, source collection IDs, cross-user source username/slug, `is_continuous` flag, `last_computed_at` timestamp) linked to the result collection via `collection_id` FK. Both save endpoints (`POST /me/shelves/set-operation/save` and `POST /me/shelves/cross-user-compare/save`) accept optional `is_continuous` boolean (default false). When false (one-time), behavior is unchanged — a static snapshot shelf is created. When true (continuous/live), the operation definition is stored alongside the snapshot. On each view of a continuous shelf (`GET /users/:username/shelves/:slug`), the set operation is re-executed against the current source collections and fresh results are returned; `last_computed_at` is updated. Cross-user continuous lists resolve the other user's collection by stored username+slug on each view, respecting privacy. Falls back to static data if recomputation fails. `GetUserShelves` response includes `is_continuous` flag per shelf. Shelf detail response includes `computed` object with `operation`, `is_continuous`, and `last_computed_at`. Frontend: "Keep updated" checkbox on both `SetOperationForm` and `CrossUserCompareForm` save sections. Shelf detail page shows "Live" badge with operation type for continuous lists. Owner view shows a blue info banner. OpenAPI spec updated with `is_continuous` on both save request schemas, `computed` object on `ShelfDetailResponse`.

## User Accounts — Email Verification
- [x] Email verification before full access. DB changes: `email_verified` boolean column on `users` (default false), `email_verification_tokens` table (UUID id, user_id FK, token_hash, expires_at, used, created_at) with index on user_id. Existing Google OAuth users are marked verified on migration. API: `POST /auth/register` now sends a verification email asynchronously after creating the account (token expires in 24 hours); JWT includes `email_verified` claim. `POST /auth/verify-email` validates the token (SHA-256 hashed in DB), marks user verified, returns fresh JWT with `email_verified: true`. `POST /auth/resend-verification` (authed) generates a new token and re-sends the email; no-ops if already verified. `GET /me/account` now includes `email_verified` in response. Google OAuth users (`POST /auth/google`) are auto-verified — new accounts set `email_verified = true`, existing accounts linking Google also get verified. `POST /auth/login` reads `email_verified` from DB and includes it in JWT. Email client: new `SendVerification(to, verifyURL)` method matching existing `SendPasswordReset` pattern; gracefully skips when SMTP is not configured. Webapp: registration page shows "Check your email" message after successful signup instead of redirecting to home. `/verify-email?token=` page auto-verifies on load and sets fresh JWT cookie. `EmailVerificationBanner` client component on settings page fetches `/api/me/account`, shows amber banner with "Resend verification email" button if unverified. `lib/auth.ts` `AuthUser` type includes `email_verified`. Webapp proxy routes: `POST /api/auth/verify-email` (sets JWT cookie on success), `POST /api/auth/resend-verification`. OpenAPI spec updated with both new endpoints and `email_verified` field on `/me/account`.

## Discussion Threads — Similar Thread Suggestions
- [x] Similar thread detection and linking. DB changes: `pg_trgm` extension enabled; GIN trigram index on `threads.title` for fast similarity lookups. API endpoints: `GET /books/:workId/similar-threads?title=...` (find similar threads by title before creating a new one, returns up to 5 matches above 0.3 similarity threshold) and `GET /threads/:threadId/similar` (find similar threads on the same book for an existing thread). Both return thread data plus a `similarity` float score, sorted by similarity descending. Frontend: thread creation form in `ThreadList` component debounce-searches for similar threads as the user types a title (400ms delay, minimum 5 chars), displaying matches in an amber suggestion box with links to existing discussions. `SimilarThreads` client component on thread detail pages fetches and displays related threads under a "Similar Discussions" section. Webapp proxy routes: `GET /api/books/:workId/similar-threads` and `GET /api/threads/:threadId/similar`. OpenAPI spec updated with both endpoints and `SimilarThreadResponse` schema.

## Book Stats — Precomputed Aggregates
- [x] Precomputed aggregate stats per book to avoid expensive multi-join COUNT/AVG queries on hot paths. DB table `book_stats` keyed by `book_id` (FK → books) stores `reads_count`, `want_to_read_count`, `rating_sum`, `rating_count`, `review_count`, and `updated_at`. Stats are refreshed asynchronously (fire-and-forget goroutines) whenever a user changes a book's status, rating, or review — hooked into `userbooks.AddBook`, `userbooks.UpdateBook`, `userbooks.RemoveBook`, `userbooks.setStatusLabel`, `tags.SetBookTag`, `tags.UnsetBookTag`, `tags.UnsetBookTagValue`, and `imports.commitRow`. Backfilled for all existing books at API startup via `bookstats.BackfillAll`. `GET /books/:workId` now reads `local_reads_count` and `local_want_to_read_count` from the `book_stats` table (simple PK + LEFT JOIN) instead of running the previous 6-way JOIN with COUNT(*) FILTER across `user_books`, `books`, `users`, `tag_keys`, `book_tag_values`, and `tag_values`. New API endpoint: `GET /books/:workId/stats` returns all cached stats (reads_count, want_to_read_count, average_rating, rating_count, review_count). Webapp proxy route: `GET /api/books/:workId/stats`. OpenAPI spec updated with `BookStats` schema and `/books/{workId}/stats` endpoint.

## Book Scanning — ISBN Barcode Scanner
- [x] Scan a book's ISBN barcode to look it up and add to library. API endpoint: `POST /books/scan` accepts a `multipart/form-data` image upload, decodes it, detects EAN-13 barcodes using gozxing (pure Go ZXing port), calls existing `LookupBookByISBN` to resolve and upsert the book, and returns `{isbn, book}`. Returns 422 with hint if no barcode detected, 404 if ISBN found but no book matches. Frontend: `/scan` page with `BookScanner` client component offering three input modes: (1) Camera — uses browser `BarcodeDetector` API for real-time scanning on supported devices (Chrome/Android), with live video feed and auto-detection every 500ms; (2) Upload — sends photo to `POST /api/books/scan` for server-side detection; (3) Enter ISBN — manual input via `GET /api/books/lookup`. Detected books display with cover, metadata, and `StatusPicker` for instant library addition. Supports scanning multiple books per session with a history list showing all scanned books with their own `StatusPicker`. "Scan" link added to nav bar (authed users) and settings page. Webapp proxy routes: `POST /api/books/scan`, `GET /api/books/lookup`. OpenAPI spec updated with `POST /books/scan` endpoint.

## User Accounts — Google OAuth
- [x] OAuth via Google for sign-in and registration. DB changes: `google_id` column on `users` (varchar(255), unique partial index), `password_hash` made nullable for OAuth-only accounts. API endpoint: `POST /auth/google` accepts `{google_id, email, name}` from the webapp after it exchanges a Google authorization code for tokens. Three flows: (1) existing Google user found by `google_id` — issue JWT, (2) existing email user — link `google_id` and issue JWT, (3) new user — auto-derive username from email prefix (with numeric suffix if taken), set `display_name` from Google profile, create default shelves, issue JWT. Password login returns a specific error for Google-only accounts. Webapp routes: `GET /api/auth/google` redirects to Google consent screen, `GET /api/auth/google/callback` exchanges code for tokens via Google's token endpoint, fetches user info, calls `POST /auth/google`, sets JWT cookie, redirects to home. Frontend: "Continue with Google" / "Sign up with Google" buttons on login and register pages (conditionally shown when `NEXT_PUBLIC_GOOGLE_CLIENT_ID` is set). Google error messages displayed on the login page for OAuth failures. Environment variables: `NEXT_PUBLIC_GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `NEXT_PUBLIC_URL`. OpenAPI spec updated.
