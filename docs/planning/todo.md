# Features

Backlog of small tasks for nephewbot to pick off. Each item should be self-contained and implementable without external coordination. Items are ordered by priority — nephewbot picks the top unchecked item.

## BUGS

## stats & data

## notifications & feed

- [ ] Add pagination to `GET /users/{username}/activity`. Currently `api/handlers/activity.go` `GetUserActivity` hard-limits to 30 results with no pagination. Add cursor-based pagination matching the pattern used in `GetFeed` — accept a `cursor` query param (RFC3339Nano timestamp), add `AND a.created < :cursor` to the WHERE clause, and return `next_cursor` from the last result's timestamp. On the frontend (`webapp/src/app/[username]/page.tsx`), add a "Load more" button that passes the cursor.

- [ ] Add notification deletion endpoint. Add `DELETE /me/notifications/{notifId}` to `api/main.go` (authed group) with a handler in `api/handlers/notifications.go` that finds the notification by ID, verifies `user = auth.Id`, and deletes it. Return 200 `{ "ok": true }` or 404. Add corresponding proxy route `webapp/src/app/api/me/notifications/[notifId]/route.ts` exporting DELETE. Add a dismiss/delete button (small X icon) on each notification card in `webapp/src/app/notifications/page.tsx`.

## profile & social

- [ ] Add blocked users list page. Create `GET /me/blocks` endpoint — add route `authed.GET("/me/blocks", handlers.GetBlockedUsers(app))` in `api/main.go`. The handler in `api/handlers/blocks.go` should query `SELECT u.id, u.username, u.display_name, u.avatar FROM blocks b JOIN users u ON b.blocked = u.id WHERE b.blocker = :userId ORDER BY b.created DESC` and return the list with avatar URLs. Add proxy route `webapp/src/app/api/me/blocks/route.ts`. Create settings page `webapp/src/app/settings/blocked/page.tsx` with a list of blocked users, each with an "Unblock" button that calls `DELETE /api/users/:username/block`. Add "Blocked users" link to `SettingsNav` component in `webapp/src/components/settings-nav.tsx`.

- [ ] Add followed authors list page at `/settings/followed-authors`. Create `webapp/src/app/settings/followed-authors/page.tsx` — server component that fetches `GET ${API_URL}/me/followed-authors` with the auth token. Display each author as a card with name and an "Unfollow" button calling `DELETE /api/authors/:authorKey/follow`. Add "Followed authors" link to `SettingsNav` in `webapp/src/components/settings-nav.tsx`.

## search & browse

- [ ] Add pagination to book search results. Currently `api/handlers/books.go` `SearchBooks` returns at most 20 results with no pagination. Add `page` query param (default 1). For local results, use `OFFSET = (page-1)*20`. For OL results, add `&offset=` to the OL search URL. On the frontend (`webapp/src/app/search/page.tsx`), add "Next page" / "Previous page" buttons below search results, passing `page` as a URL query param.

- [ ] Add search within user's library. Add `GET /users/{username}/books/search?q=` endpoint — register in `api/main.go` under the optional-auth group. The handler in `api/handlers/userbooks.go` should query `user_books JOIN books WHERE user = :userId AND (b.title LIKE '%q%' OR b.authors LIKE '%q%')`, respecting privacy via `canViewProfile`. Return same format as `GetUserBooks`. On the frontend, add a search input above the book grid in `webapp/src/components/library-manager.tsx` that calls `GET /api/users/:username/books/search?q=` and filters the displayed books client-side or fetches from the API with debounce (400ms).

## book detail & discovery

- [ ] Show series navigation on book detail page. On `webapp/src/app/books/[workId]/page.tsx`, when the book has series memberships (the `series` array from `GET /books/:workId`), display a horizontal row of other books in the same series below the series badge. Fetch the series detail from `GET /api/series/:seriesId` (already exists) and render up to 6 book covers with title, position number, and a link to each book's detail page. Highlight the current book. Only show for the first series if the book belongs to multiple.

- [ ] Add description to series pages. Add `description` text field to the `series` PocketBase collection in a new migration file `api/migrations/`. Update `GET /series/{seriesId}` handler in `api/handlers/series.go` to return the `description` field. Add `PUT /series/{seriesId}` endpoint (auth required) to update the description — register in `main.go` authed group. On the series page (`webapp/src/app/series/[seriesId]/page.tsx`), display the description below the series name. Add an inline edit button for logged-in users that reveals a textarea to update the description via `PUT /api/series/:seriesId`.

- [ ] Add genre chips to book search results. In `api/handlers/books.go` `SearchBooks`, for local books that have a `subjects` column, include a `subjects` array (split comma-separated string, take first 3) in each result object. On the frontend search results (`webapp/src/app/search/page.tsx`), render up to 3 small genre/subject chips below each book result's title and author. Each chip links to `/genres/:slug` (slugify the subject name).

## settings & account

- [ ] Add "sent recommendations" tab to recommendations page. Add `GET /me/recommendations/sent` endpoint — register `authed.GET("/me/recommendations/sent", handlers.GetSentRecommendations(app))` in `api/main.go`. Handler in `api/handlers/recommendations.go` queries `SELECT r.*, u.username, u.display_name, u.avatar, b.title, b.open_library_id, b.cover_url FROM recommendations r JOIN users u ON r.recipient = u.id JOIN books b ON r.book = b.id WHERE r.sender = :userId ORDER BY r.created DESC`. Add proxy route `webapp/src/app/api/me/recommendations/sent/route.ts`. On `webapp/src/app/recommendations/page.tsx`, add tabs "Received" / "Sent" — the Sent tab shows recommendations the user has sent, with recipient name, book cover, note, and status badge.

- [ ] Add account deletion endpoint. Add `DELETE /me/account` (auth required) to `api/main.go` with handler in `api/handlers/userdata.go`. The handler should first call the existing `DeleteAllData` logic to remove all user-owned data, then delete the user record itself from the `users` collection. On the settings page danger zone (`webapp/src/components/delete-data-form.tsx`), add a second button "Delete my account permanently" below the existing "Delete all my data" button, with a confirmation that requires typing "delete my account". This calls `DELETE /api/me/account` and clears the auth cookie, redirecting to the home page.

## UX polish

- [ ] Add `loading.tsx` skeleton files for the four highest-traffic pages. Create `webapp/src/app/search/loading.tsx` (grid of 8 skeleton cards with pulsing placeholder for cover, title line, author line), `webapp/src/app/books/[workId]/loading.tsx` (skeleton with cover placeholder, title bar, description lines, review placeholders), `webapp/src/app/[username]/loading.tsx` (avatar circle, name bar, stats row, book grid skeletons), and `webapp/src/app/notifications/loading.tsx` (list of 5 notification row skeletons). Each skeleton should use Tailwind's `animate-pulse` with `bg-surface-2 rounded` placeholder divs matching the approximate layout of the real page.

- [ ] Add `not-found.tsx` for book and user pages. Create `webapp/src/app/books/[workId]/not-found.tsx` — display "Book not found" with a search link. Create `webapp/src/app/[username]/not-found.tsx` — display "User not found" with a link to `/users`. In the corresponding `page.tsx` files, call `notFound()` (from `next/navigation`) when the API returns 404 instead of showing a generic error.

- [ ] Add keyboard shortcut hint that works cross-platform. In `webapp/src/components/nav.tsx`, the search bar shows "⌘K" which is Mac-only. Detect the user's OS via `navigator.userAgent` or `navigator.platform` in a client component and show "Ctrl+K" on Windows/Linux and "⌘K" on Mac. Extract this into a small `KeyboardShortcutHint` client component that accepts a `keys` prop like `{ mac: "⌘K", other: "Ctrl+K" }`.

- [ ] Truncate long author bios on author pages. In `webapp/src/app/authors/[authorKey]/page.tsx` (or the component rendering the bio), if the bio text exceeds 500 characters, truncate it and show a "Read more" toggle button. Use a client component with `useState` to toggle between truncated (first 500 chars + "...") and full text. Apply `prose` class from Tailwind typography plugin for better bio formatting.

- [ ] Add responsive hamburger menu for mobile navigation. In `webapp/src/components/nav.tsx`, wrap the desktop nav links in a container that hides below `md:` breakpoint. Add a hamburger button (`☰`) visible only below `md:` that toggles a full-width dropdown panel with all nav links stacked vertically. Use a client component with `useState` for the open/close toggle. Close the menu when a link is clicked or when clicking outside.

## blocked

## API gaps

- [ ] Add `GET /books/{workId}/similar-threads?title=` endpoint. The API docs document this for finding similar threads before creating a new one, and the completed.md says it was implemented with `pg_trgm`, but the route is not registered in `api/main.go` and no handler function exists. Create `SimilarThreads` handler in `api/handlers/threads.go` that queries `SELECT id, title, username, comment_count, similarity(title, :title) as sim FROM threads WHERE book = :book AND deleted_at IS NULL AND similarity(title, :title) > 0.3 ORDER BY sim DESC LIMIT 5`. Register as `se.Router.GET("/books/{workId}/similar-threads", handlers.SimilarThreads(app))` in main.go (public route).

- [ ] Add `GET /threads/{threadId}/similar` endpoint. Similar to the above but for an existing thread — finds threads on the same book with similar titles. Create handler in `api/handlers/threads.go` that loads the thread, gets its book_id and title, then queries other threads on the same book with `similarity(title, :title) > 0.3`. Register as `se.Router.GET("/threads/{threadId}/similar", handlers.GetSimilarThreads(app))` in main.go (public route). On the thread detail page component, add a "Similar Discussions" section below the comments.

- [ ] Add `GET /me/books/{olId}/editions` convenience endpoint. Currently to select an edition, the frontend must call `GET /books/{workId}/editions` (which proxies to Open Library) and then `PATCH /me/books/{olId}` to save the selection. Add a new endpoint that returns the user's currently selected edition alongside the full editions list. Register `authed.GET("/me/books/{olId}/editions", handlers.GetMyBookEditions(app))` in main.go. The handler should return `{ "selected_edition_key": "...", "selected_edition_cover_url": "...", "editions": [...] }` by combining the user_books selection with the OL editions response.

## series improvements

- [ ] Add series deletion endpoint for empty series. Add `DELETE /series/{seriesId}` (auth required) to `api/main.go`. The handler in `api/handlers/series.go` should check that the series has zero `book_series` links (no books), then delete it. Return 400 if the series still has books, 200 on success. This prevents orphaned series records from accumulating.

- [ ] Add series edit endpoint. Add `PATCH /series/{seriesId}` (auth required) to `api/main.go`. The handler in `api/handlers/series.go` should accept `{ "name": "...", "description": "..." }` and update the series record. Only the `name` and `description` fields should be updatable. Return 200 with the updated series. On the series page (`webapp/src/app/series/[seriesId]/page.tsx`), add an edit button for logged-in users that opens an inline form to rename the series or edit its description.

## import improvements

- [ ] LLM-powered fuzzy matching for failed import lookups. When standard book lookups fail to find a match during import, fall back to a "power mode" that uses an LLM to generate title/author permutations (alternate spellings, subtitle variations, series name removal, etc.) and retry searches with each permutation until possible matches are found. Present the candidate matches to the user for confirmation.

## Pending PRs

- [Fix search page popular section to match active tab](https://github.com/ross-corp/rosslib/pull/120) — show popular authors/people instead of always showing books when on Authors/People tabs
- [Fix: flatten simultaneous finish + rate into one feed event](https://github.com/ross-corp/rosslib/pull/119) — merge near-simultaneous finished_book + rated activities into a single "finished and rated" feed event
- [Fix library sidebar counts not updating on re-label](https://github.com/ross-corp/rosslib/pull/118) — optimistic count updates in sidebar when moving books between statuses or removing them
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
- [Add recently viewed books (client-side)](https://github.com/ross-corp/rosslib/pull/104) — useRecentlyViewed hook with localStorage, RecordRecentlyViewed on book detail, recently viewed row on search page
- [Add saved searches](https://github.com/ross-corp/rosslib/pull/105) — saved_searches collection, CRUD API endpoints, SavedSearches component with chips on search page
- [Add review sorting options on book detail page](https://github.com/ross-corp/rosslib/pull/106) — sort query param on GET /books/:workId/reviews and sort dropdown on book detail page
- [Add 'more by this author' section on book detail page](https://github.com/ross-corp/rosslib/pull/107) — "More by {Author}" section with up to 6 book covers, reuses existing author works endpoint
- [Add book quotes/highlights](https://github.com/ross-corp/rosslib/pull/108) — book_quotes collection, CRUD API endpoints, BookQuoteList component on book detail page with public/private toggle
- [Add thread locking for moderators](https://github.com/ross-corp/rosslib/pull/109) — locked_at column on threads, lock/unlock endpoints, locked banner and mod toggle on frontend
- [Add review comments / discussion under reviews](https://github.com/ross-corp/rosslib/pull/110) — review_comments collection, GET/POST/DELETE endpoints, ReviewComments component on book detail page with notifications
- [Add label descriptions](https://github.com/ross-corp/rosslib/pull/111) — description column on collections, create/edit UI in library manager, render on visitor shelf detail page
- [Add LibraryThing CSV import](https://github.com/ross-corp/rosslib/pull/112) — LibraryThing TSV import with preview/commit endpoints, tab-delimited parsing, author name reversal, collections/tags as labels
- [Add API token generation for integrations](https://github.com/ross-corp/rosslib/pull/113) — api_tokens collection, CRUD endpoints, Bearer token auth in middleware, settings UI
- [Add loading skeleton components](https://github.com/ross-corp/rosslib/pull/114) — Skeleton base component, composed skeletons, loading.tsx files for feed/profile/book detail/search
- [Add keyboard shortcuts with help overlay](https://github.com/ross-corp/rosslib/pull/115) — useKeyboardShortcuts hook, shortcuts overlay modal, hint badge for logged-in users
- [Add dark mode with light/dark/system theme toggle](https://github.com/ross-corp/rosslib/pull/116) — theme infrastructure (CSS variables, data-theme attribute, FOUC prevention), theme toggle in nav, theme API endpoint, semantic color token conversion across ~45 files
- [Add empty state illustrations for zero-data pages](https://github.com/ross-corp/rosslib/pull/117) — reusable EmptyState component with consistent CTA links on feed, notifications, library, and label pages
- [Auto-predict label names from shelf naming patterns](https://github.com/ross-corp/rosslib/pull/121) — detect common prefix and year-based patterns in Goodreads/StoryGraph shelf names and auto-group them into labels during import