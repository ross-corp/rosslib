# Features

Backlog of small tasks for nephewbot to pick off. Each item should be self-contained and implementable without external coordination. Items are ordered by priority — nephewbot picks the top unchecked item.

## BUGS

## stats & data

## notifications & feed

## profile & social

## search & browse

## book detail & discovery

## settings & account

- [ ] Add email change functionality to settings page. Currently users can change their password and display name in settings, but there's no way to change their email address. Add `PUT /me/email` endpoint in `api/handlers/auth.go` that accepts `{ new_email, current_password }`, verifies the password, checks the new email isn't already in use, updates `users.email`, sets `email_verified = false`, and sends a new verification email. Add corresponding proxy route and form component on the settings page below the password form.

## UX polish

- [ ] Add `not-found.tsx` for book and user pages. Create `webapp/src/app/books/[workId]/not-found.tsx` — display "Book not found" with a search link. Create `webapp/src/app/[username]/not-found.tsx` — display "User not found" with a link to `/users`. In the corresponding `page.tsx` files, call `notFound()` (from `next/navigation`) when the API returns 404 instead of showing a generic error.

- [ ] Add keyboard shortcut hint that works cross-platform. In `webapp/src/components/nav.tsx`, the search bar shows "⌘K" which is Mac-only. Detect the user's OS via `navigator.userAgent` or `navigator.platform` in a client component and show "Ctrl+K" on Windows/Linux and "⌘K" on Mac. Extract this into a small `KeyboardShortcutHint` client component that accepts a `keys` prop like `{ mac: "⌘K", other: "Ctrl+K" }`.

- [ ] Truncate long author bios on author pages. In `webapp/src/app/authors/[authorKey]/page.tsx` (or the component rendering the bio), if the bio text exceeds 500 characters, truncate it and show a "Read more" toggle button. Use a client component with `useState` to toggle between truncated (first 500 chars + "...") and full text. Apply `prose` class from Tailwind typography plugin for better bio formatting.

- [ ] Add responsive hamburger menu for mobile navigation. In `webapp/src/components/nav.tsx`, wrap the desktop nav links in a container that hides below `md:` breakpoint. Add a hamburger button (`☰`) visible only below `md:` that toggles a full-width dropdown panel with all nav links stacked vertically. Use a client component with `useState` for the open/close toggle. Close the menu when a link is clicked or when clicking outside.

- [ ] Add `aria-label` attributes to nav bar interactive elements. In `webapp/src/components/nav.tsx`, none of the nav links, buttons, or icons have `aria-label` attributes. Add `aria-label` to: the search input ("Search books"), the notification bell icon ("Notifications"), the user avatar dropdown ("User menu"), the "Browse" dropdown ("Browse menu"), the "Community" dropdown ("Community menu"), and the sign-out button ("Sign out"). This improves screen reader accessibility.

- [ ] Add `role="alert"` to form validation error messages. Across login (`webapp/src/app/login/page.tsx`), register (`webapp/src/app/register/page.tsx`), forgot-password (`webapp/src/app/forgot-password/page.tsx`), and reset-password (`webapp/src/app/reset-password/page.tsx`), error messages are displayed as styled `<p>` elements but are not announced to screen readers. Add `role="alert"` and `aria-live="assertive"` to the error message containers so screen readers announce validation failures immediately.

- [ ] Add Escape key handler to close dropdowns. In `webapp/src/components/shelf-picker.tsx`, `webapp/src/components/book-tag-picker.tsx`, and `webapp/src/components/nav-dropdown.tsx`, dropdowns close when clicking outside but don't respond to the Escape key. Add a `useEffect` that listens for `keydown` events and closes the dropdown when `event.key === "Escape"`. This is a standard accessibility pattern for dropdown menus.

- [ ] Add relative timestamps to feed activity items. In `webapp/src/app/feed/page.tsx` (or the feed components), activity timestamps are shown as full dates. Add a relative time display (e.g. "2 hours ago", "3 days ago") using a simple helper function that calculates the difference between now and the activity's `created_at`. Show the relative time as the primary display and the full date as a `title` attribute on hover.

- [ ] Add user library book count to profile page header. The profile page at `webapp/src/app/[username]/page.tsx` shows `books_read` (finished books) and `currently_reading_count`, but doesn't show the total number of books in the user's library (across all statuses). Add a "Library" stat to the stats row that shows the total count. This can be derived from the shelves response: sum of `item_count` across all read-status shelves (want-to-read + currently-reading + read), or add a `total_books` field to the `GET /users/:username` profile response in `api/handlers/users.go`.

## data integrity

- [ ] Add `self_link` check to `CreateBookLink`. In `api/handlers/links.go`, the `CreateBookLink` handler does not verify that `from_book_id != to_book_id`. A user can create a link from a book to itself, which is meaningless. Add a check after resolving both book IDs: if they're the same, return 400 with `"cannot link a book to itself"`.

- [ ] Add length validation to thread title and comment body. In `api/handlers/threads.go`, the `CreateThread` handler accepts `title` and `body` from the request but the API docs say title is max 500 chars and body is max 10,000 chars. Verify these limits are enforced in the handler code. If not, add validation: `if len(title) > 500 { return 400 "title must be 500 characters or fewer" }` and `if len(body) > 10000 { return 400 "body must be 10,000 characters or fewer" }`. Same for `AddComment`: body max 5,000 chars.

- [ ] Add rate limit to recommendation sending. In `api/handlers/recommendations.go`, `SendRecommendation` has no rate limiting — a user could spam another user with recommendations. Add a check that a user cannot send more than 10 recommendations in a 24-hour window. Query `SELECT COUNT(*) FROM recommendations WHERE sender = :userId AND created >= datetime('now', '-1 day')`. Return 429 with `"too many recommendations, try again later"` if the limit is exceeded.

## blocked

## API gaps

- [ ] Add `GET /books/{workId}/similar-threads?title=` endpoint handler. The webapp proxy routes exist (`webapp/src/app/api/books/[workId]/similar-threads/route.ts` and `webapp/src/app/api/threads/[threadId]/similar/route.ts`) and `completed.md` references this as done, but NO handler function exists in `api/handlers/threads.go` and NO route is registered in `api/main.go`. Create `SimilarThreads` handler in `api/handlers/threads.go` that queries threads on the same book with `similarity(title, :title) > 0.3` (using pg_trgm) and returns up to 5 matches. Register `se.Router.GET("/books/{workId}/similar-threads", handlers.SimilarThreads(app))`. Also create `GetSimilarThreads` handler for `GET /threads/{threadId}/similar` and register it.

- [ ] Add `GET /me/books/{olId}/editions` convenience endpoint. Currently to select an edition, the frontend must call `GET /books/{workId}/editions` (which proxies to Open Library) and then `PATCH /me/books/{olId}` to save the selection. Add a new endpoint that returns the user's currently selected edition alongside the full editions list. Register `authed.GET("/me/books/{olId}/editions", handlers.GetMyBookEditions(app))` in main.go. The handler should return `{ "selected_edition_key": "...", "selected_edition_cover_url": "...", "editions": [...] }` by combining the user_books selection with the OL editions response.

- [ ] Add `GET /books/:workId/followers/count` public endpoint for book follow counts. Currently the only way to see how many users follow a book is to be authenticated and check your own follow status. Add a public endpoint in `api/handlers/books.go` that returns `{ "follower_count": N }` by counting rows in `book_follows WHERE book_id = :bookId`. Register as `se.Router.GET("/books/{workId}/followers/count", handlers.GetBookFollowerCount(app))` in main.go. Display this count on the book detail page next to the follow button as "N followers".

- [ ] Add `blocked_at` timestamp to block list response. In `api/handlers/blocks.go`, the `GetBlockedUsers` handler returns blocked user info but doesn't include when the block was created. Include the `created` timestamp from the `blocks` record in the response as `blocked_at`. In `webapp/src/app/settings/blocked/page.tsx`, display the block date next to each blocked user (e.g. "Blocked on Jan 15, 2026").

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