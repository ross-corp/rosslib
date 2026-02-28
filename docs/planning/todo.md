# Features

Backlog of small tasks for nephewbot to pick off. Each item should be self-contained and implementable without external coordination. Items are ordered by priority — nephewbot picks the top unchecked item.

## stats & data

## notifications & feed

## profile & social

## search & browse

- [ ] Add recently viewed books (client-side). In the webapp, create a `useRecentlyViewed` hook in `webapp/src/lib/recently-viewed.ts` that stores the last 10 viewed book IDs + titles + covers in `localStorage`. Update the book detail page (`webapp/src/app/books/[workId]/page.tsx`) to call `addToRecentlyViewed(book)` on mount. On the search page, when no query is entered, show a "Recently Viewed" row above trending/popular books if the list is non-empty. Render as a book cover row with titles. No API changes needed — purely client-side state.

- [ ] Add saved searches. Create a `saved_searches` table via migration: `id` (uuid PK), `user_id` (FK → users, cascade), `name` (varchar(100)), `query` (text), `filters` (jsonb — stores genre, language, year_min, year_max, sort, tab), `created_at`. API endpoints in a new `api/handlers/searches.go`: `GET /me/saved-searches` (list), `POST /me/saved-searches` (create, max 20 per user), `DELETE /me/saved-searches/:id`. On the search page, when filters are active, show a "Save this search" button that opens a name input. Show saved searches as chips above the search bar. Clicking a chip populates the query and filters. Webapp proxy routes: `GET/POST/DELETE /api/me/saved-searches`.

## book detail & discovery

- [ ] Add review sorting options on book detail page. In `GET /books/:workId/reviews` in `api/handlers/books.go`, accept a `sort` query param: `newest` (default — current behavior), `oldest`, `highest` (rating DESC), `lowest` (rating ASC), `most_liked` (LEFT JOIN `review_likes` GROUP BY count DESC — depends on PR #51 being merged, fall back to `newest` if review_likes table doesn't exist). On the book detail page, add a sort dropdown above the reviews section. Forward the `sort` param through the webapp proxy route.

- [ ] Add "more by this author" section on book detail page. On the book detail page (`webapp/src/app/books/[workId]/page.tsx`), below the main book info, show a "More by {Author}" section. Extract the first author's OL author key from the book detail response (the `authors` field may include keys if they were fetched from OL). If an author key is available, call `GET /authors/:authorKey` and display up to 6 book covers in a horizontal row (exclude the current book). If no author key is available, search local books by author name via `GET /books/search?q={authorName}` and filter out the current book. Link "See all" to the author page.

- [ ] Add book quotes/highlights. Create a `book_quotes` table via migration: `id` (uuid PK), `user_id` (FK → users, cascade), `book_id` (FK → books, cascade), `text` (text, max 2000 chars), `page_number` (integer, nullable), `note` (text, nullable, max 500 chars), `is_public` (boolean, default true), `created_at`. API endpoints in `api/handlers/quotes.go`: `GET /books/:workId/quotes` (public quotes for a book, paginated), `GET /me/books/:olId/quotes` (user's own quotes), `POST /me/books/:olId/quotes`, `DELETE /me/quotes/:quoteId`. On the book detail page, add a "Quotes" tab or section showing public quotes with attribution (username, page number). On the user's book view, show an "Add quote" button. Webapp proxy routes for all endpoints.

- [ ] Add thread locking for moderators. Add a `locked_at` column (timestamptz, nullable) to `threads` via migration. Add `POST /threads/:threadId/lock` and `POST /threads/:threadId/unlock` endpoints in `api/handlers/threads.go` — moderator-only (check `is_moderator` from auth). When locked, `POST /threads/:threadId/comments` returns 403 with "This thread is locked." In the frontend, show a lock icon on locked threads and hide the comment form. Moderators see a lock/unlock toggle button on the thread detail page.

- [ ] Add review comments / discussion under reviews. Create a `review_comments` table via migration: `id` (uuid PK), `user_id` (FK → users, cascade), `book_id` (FK → books, cascade), `review_user_id` (FK → users, cascade — the review author), `body` (text, max 2000 chars), `created_at`, `deleted_at` (timestamptz, nullable — soft delete). API endpoints in `api/handlers/reviewcomments.go`: `GET /books/:workId/reviews/:userId/comments` (list comments on a review, paginated), `POST /books/:workId/reviews/:userId/comments` (add comment), `DELETE /review-comments/:commentId` (author or moderator). Generate a `review_comment` notification for the review author. On the book detail page, show a comment count per review and a toggle to expand/collapse the comment thread under each review. Webapp proxy routes for all endpoints.

## settings & account

- [ ] Add label descriptions. Add a `description` column (text, nullable, max 1000 chars) to `collections` via migration. Update `POST /me/shelves` and `PATCH /me/shelves/:id` in `api/handlers/collections.go` to accept and persist `description`. Return it in `GET /users/:username/shelves` and label detail responses. On the label detail page, render the description below the label name. On the create/edit label form, add a textarea for description. Don't add descriptions to the three default status labels (Want to Read, Currently Reading, Read).

- [ ] Add LibraryThing CSV import. Add `POST /me/import/librarything/preview` and `POST /me/import/librarything/commit` endpoints in `api/handlers/imports.go`, following the same pattern as Goodreads and StoryGraph imports. LibraryThing CSV format: columns include `Title`, `Author (First, Last)`, `ISBN`, `Rating`, `Review`, `Date Read`, `Collections` (semicolon-separated), `Tags` (comma-separated). Map LT collections and tags to labels. Status mapping: "Currently Reading" → Currently Reading, "To Read" / "Wishlist" → Want to Read, "Read but unowned" / books with a Date Read → Finished. On the import page, add a "LibraryThing" tab alongside Goodreads and StoryGraph (extends PR #61's tabbed UI).

- [ ] Add API token generation for integrations. Create an `api_tokens` table via migration: `id` (uuid PK), `user_id` (FK → users, cascade), `name` (varchar(100) — user-chosen label like "CLI" or "Calibre"), `token_hash` (text — SHA-256 hash), `last_used_at` (timestamptz nullable), `created_at`. API endpoints: `GET /me/api-tokens` (list tokens — return name, created_at, last_used_at, never the raw token), `POST /me/api-tokens` (create — return the raw token once, hash and store), `DELETE /me/api-tokens/:id`. In auth middleware, also check for `Authorization: Bearer <token>` against `api_tokens` (hash the incoming token and look up). Add a "Developer" or "API Tokens" section to settings page with create/revoke UI. Max 5 tokens per user.

## UX polish

- [ ] Add loading skeleton components. Create a `Skeleton` component in `webapp/src/components/skeleton.tsx` — a pulsing gray placeholder block that accepts `width`, `height`, and `variant` ("text", "circular", "rectangular") props. Create composed skeletons: `BookGridSkeleton` (grid of cover-sized rectangles), `ProfileSkeleton` (avatar circle + text lines), `ReviewSkeleton` (star row + text block). Use React Suspense boundaries in the main pages (feed, profile, book detail, search results) with these skeletons as fallbacks. No external dependency — CSS animation with `@keyframes`.

- [ ] Add keyboard shortcuts. Create a `useKeyboardShortcuts` hook in `webapp/src/lib/keyboard-shortcuts.ts` that registers global keydown listeners. Shortcuts: `/` focuses the search input (prevent typing "/" in the box), `Escape` closes any open modal/dropdown, `?` shows a shortcuts help overlay. Only register when no input/textarea is focused (except Escape). Show a small "Press ? for shortcuts" hint at the bottom of the page for logged-in users. The shortcuts overlay is a simple modal listing all available shortcuts.

- [ ] Add dark mode. Add a `theme` preference to the user — either in `localStorage` for unauthenticated users or as a `theme` column on `users` (varchar(10), default `'system'`, values `'light'`, `'dark'`, `'system'`). In `webapp/src/app/layout.tsx`, read the theme preference and apply a `data-theme` attribute to `<html>`. Define CSS custom properties for all colors in `globals.css` under `[data-theme="light"]` and `[data-theme="dark"]` selectors. For `system`, use `prefers-color-scheme` media query. Add a theme toggle (sun/moon icon) in the nav bar. Convert existing hardcoded colors to use the CSS custom properties. This is a larger task — start with the infrastructure (toggle, CSS variables, layout attribute) and convert colors page by page.

- [ ] Add empty state illustrations for zero-data pages. For pages that show nothing when a user is new (feed, library, labels, notifications), add friendly empty state messages with call-to-action links. Feed: "Your feed is empty. Follow some readers to see their activity." with link to `/users`. Library: "No books yet. Search for a book to get started." with link to `/search`. Notifications: "No notifications. You're all caught up!" Labels: "This label is empty. Browse books to add some." Check each page for existing empty handling and add where missing. Use the same visual style (centered text, muted color, optional icon).

## blocked

- [ ] Populate series data from Open Library during book lookup. **BLOCKED: depends on PR #60 (series metadata) being merged first.** Once the `series` and `book_series` collections exist, update `GetBookDetail` in `api/handlers/books.go` to auto-detect series data. The OL editions response (`/works/{workId}/editions.json`) includes a `series` array on some editions. For each edition entry, check for a `series` field. If found, find-or-create a `series` record by name and create a `book_series` link with the position number (if available). Also try the OL work's `subjects` array for series-like patterns (e.g. "Harry Potter" appearing as a subject). This is best-effort — not all OL works have series data. Log when series data is found vs. not for visibility into coverage.

## BUGS

- [ ] numbers on left sidebar of library do not update if I re-label books. for ex, marking something from in progress to read should subtract 1 and add 1 to the other

- [ ] in the feed, when a user finishes and rates something at the same time, those should be flattened into one event. "user finished X and rated it Y stars"

- [ ] http://huey.taila415c.ts.net:3000/search?type=authors on the "popular on rosslib", this always shows books, but when I've selected authors I should see authors, and so for people

## import improvements

- [ ] Auto-predict label names from Goodreads shelf naming patterns. If a user has Goodreads shelves with a common prefix (e.g. `genre-scifi`, `genre-romance`, `genre-fantasy`), detect the prefix and suggest it as the label name (e.g. "genre") with the suffix as the value. Same logic for year-based Goodreads shelves (e.g. `read-2023`, `read-2024`) — detect the year pattern and suggest "read" as the label with the year as the value.
  - [ ] Same pattern detection for year-based Goodreads shelf names (e.g. `read-2023`, `read-2024` → label "read", values "2023", "2024")

- [ ] LLM-powered fuzzy matching for failed import lookups. When standard book lookups fail to find a match during import, fall back to a "power mode" that uses an LLM to generate title/author permutations (alternate spellings, subtitle variations, series name removal, etc.) and retry searches with each permutation until possible matches are found. Present the candidate matches to the user for confirmation.

## Pending PRs

- [Add sort options to owner library view and tag/label endpoints](https://github.com/ross-corp/rosslib/pull/90) — sort dropdown in LibraryManager + sort param on tags/labels API endpoints
- [Add confirmation dialog when removing books from library](https://github.com/ross-corp/rosslib/pull/91) — reusable ConfirmDialog component, applied to shelf grid, library manager bulk remove, and shelf picker
- [Wire toast notifications into all user actions](https://github.com/ross-corp/rosslib/pull/92) — extend existing toast system to cover import, quick-add, bulk library ops, settings, export, block, and reading progress
- [Use BookCoverPlaceholder consistently across all book cover fallbacks](https://github.com/ross-corp/rosslib/pull/93) — replace plain div fallbacks with BookCoverPlaceholder in 18 files
- [Add reading pace calculation for currently-reading books](https://github.com/ross-corp/rosslib/pull/94) — show pages/day and estimated finish date below reading progress bar
- [Add re-read tracking with reading sessions](https://github.com/ross-corp/rosslib/pull/95) — reading_sessions collection, CRUD API endpoints, ReadingHistory component on book detail page
- [Add year-in-review summary page](https://github.com/ross-corp/rosslib/pull/96) — year-in-review API endpoint and page with stats, top books, genres, and month-by-month grid
- [Add periodic book_stats backfill every 24 hours](https://github.com/ross-corp/rosslib/pull/97) — bookstats package with BackfillAll + StartPoller goroutine running on startup and every 24h
- [Add feed filtering by activity type](https://github.com/ross-corp/rosslib/pull/98) — type query param on GET /me/feed and filter chips on feed page
- [Add friends reading this on book detail page](https://github.com/ross-corp/rosslib/pull/99) — GET /books/:workId/readers endpoint and avatar row on book detail page
- [Add follow suggestions on feed page](https://github.com/ross-corp/rosslib/pull/100) — GET /me/suggested-follows endpoint and FollowSuggestions component on feed page
- [Add profile banner image](https://github.com/ross-corp/rosslib/pull/101) — banner file field on users, POST /me/banner endpoint, banner display on profile page, upload in settings
- [Add favorite genres display on user profile](https://github.com/ross-corp/rosslib/pull/102) — top 5 genre chips derived from finished books' subjects, displayed below bio on profile page
- [Add trending books section to search landing page](https://github.com/ross-corp/rosslib/pull/103) — GET /books/trending endpoint and horizontal scrollable row on search page