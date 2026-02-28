# Features

Backlog of small tasks for nephewbot to pick off. Each item should be self-contained and implementable without external coordination. Items are ordered by priority — nephewbot picks the top unchecked item.

## stats & data

## notifications & feed

## profile & social

## search & browse

## book detail & discovery


## settings & account

## UX polish

## blocked

- [ ] Populate series data from Open Library during book lookup. **BLOCKED: depends on PR #60 (series metadata) being merged first.** Once the `series` and `book_series` collections exist, update `GetBookDetail` in `api/handlers/books.go` to auto-detect series data. The OL editions response (`/works/{workId}/editions.json`) includes a `series` array on some editions. For each edition entry, check for a `series` field. If found, find-or-create a `series` record by name and create a `book_series` link with the position number (if available). Also try the OL work's `subjects` array for series-like patterns (e.g. "Harry Potter" appearing as a subject). This is best-effort — not all OL works have series data. Log when series data is found vs. not for visibility into coverage.

## BUGS

- [ ] http://huey.taila415c.ts.net:3000/search?type=authors on the "popular on rosslib", this always shows books, but when I've selected authors I should see authors, and so for people

## import improvements

- [ ] Auto-predict label names from Goodreads shelf naming patterns. If a user has Goodreads shelves with a common prefix (e.g. `genre-scifi`, `genre-romance`, `genre-fantasy`), detect the prefix and suggest it as the label name (e.g. "genre") with the suffix as the value. Same logic for year-based Goodreads shelves (e.g. `read-2023`, `read-2024`) — detect the year pattern and suggest "read" as the label with the year as the value.
  - [ ] Same pattern detection for year-based Goodreads shelf names (e.g. `read-2023`, `read-2024` → label "read", values "2023", "2024")

- [ ] LLM-powered fuzzy matching for failed import lookups. When standard book lookups fail to find a match during import, fall back to a "power mode" that uses an LLM to generate title/author permutations (alternate spellings, subtitle variations, series name removal, etc.) and retry searches with each permutation until possible matches are found. Present the candidate matches to the user for confirmation.

## Pending PRs

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