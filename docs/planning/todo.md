# Features

Backlog of small tasks for nephewbot to pick off. Each item should be self-contained and implementable without external coordination. Items are ordered by priority — nephewbot picks the top unchecked item.


## BUGS

- [ ] Implement password reset and email verification backend: the webapp proxy routes exist (`POST /auth/forgot-password`, `POST /auth/reset-password`, `POST /auth/verify-email`, `POST /auth/resend-verification`) but the Go handler functions and backing DB tables (`password_reset_tokens`, `email_verification_tokens`) are entirely missing. The `email_verified` field exists on users. Requires: new migration for token tables, `api/internal/email` package for SMTP, four handler functions in `api/handlers/auth.go`, route registration in `api/main.go`. See `docs/planning/completed.md` "Password Reset" and "Email Verification" sections for full spec. **Large task — not suitable for nephewbot.**
- [ ] Implement set-operation and cross-user-compare backend: the webapp proxy routes exist (`POST /me/shelves/set-operation`, `POST /me/shelves/set-operation/save`, `POST /me/shelves/cross-user-compare`, `POST /me/shelves/cross-user-compare/save`) but no Go handler functions exist in `api/handlers/collections.go`. Requires: new `computed_collections` table migration, four handler functions, route registration. See `docs/planning/completed.md` "Computed Lists" sections for full spec. **Large task — not suitable for nephewbot.**

## performance

- [ ] Add compound index on `user_books (user, date_added DESC)` for timeline queries: the reading timeline endpoint (`GET /users/:username/timeline` in `api/handlers/timeline.go`) queries `user_books` filtering by `user` and ordering by `date_read`. While the unique index on `(user, book)` exists, a compound index on `(user, date_added DESC)` or `(user, date_read)` would speed up timeline and activity queries. Add the index in the next migration file in `api/migrations/`.
- [ ] Add compound index on `activities (user, activity_type, created DESC)` for feed filtering: the feed endpoint (`GET /me/feed` in `api/handlers/activity.go`) filters activities by followed users and orders by `created DESC`. The existing index on `(user, created DESC)` helps but adding `activity_type` to the compound index would optimize feed filtering when a type filter is added (see pending PR #98). Add in the next migration file.
- [ ] Add `cache: "no-store"` to all dynamic server-side fetches: some server components in the webapp may not set `cache: "no-store"` on their `fetch()` calls, which means Next.js could cache stale data. Audit all `fetch(${process.env.API_URL}...)` calls in server components and ensure they include `{ cache: "no-store" }` for user-specific or frequently-changing data. Specifically check: `webapp/src/app/[username]/page.tsx`, `webapp/src/app/[username]/stats/page.tsx`, `webapp/src/app/[username]/timeline/page.tsx`, and `webapp/src/app/notifications/page.tsx`.

## error handling

- [ ] Add `error.tsx` boundary for book detail page: `webapp/src/app/books/[workId]/` has a `not-found.tsx` but no `error.tsx`. Create `webapp/src/app/books/[workId]/error.tsx` as a client component (`"use client"`) that displays a "Something went wrong loading this book" message with a "Try again" button (calls `reset()`) and a link to search. Follow the Next.js error boundary pattern: `export default function Error({ error, reset })`.
- [ ] Add `error.tsx` boundary for user profile page: `webapp/src/app/[username]/` has a `not-found.tsx` but no `error.tsx`. Create `webapp/src/app/[username]/error.tsx` as a client component that displays "Something went wrong loading this profile" with a "Try again" button and a link to `/users`. This catches runtime errors in the profile page and its nested routes (stats, reviews, timeline, etc.).
- [ ] Add `error.tsx` boundary for admin page: the admin page at `webapp/src/app/admin/page.tsx` loads multiple data sources (users, feedback, reports, link edits) and any of them could fail. Create `webapp/src/app/admin/error.tsx` with a "Failed to load admin panel" message, "Try again" button, and a reminder to check moderator permissions.
- [ ] Standardize server-side fetch error handling: some server components return empty arrays on fetch failure (e.g., search page returns `[]`), others return `null` (e.g., profile page), and others call `notFound()` (e.g., book detail). Establish a consistent pattern: return `notFound()` for 404s, throw for 500s (caught by `error.tsx`), and only return empty arrays for list endpoints where "no data" is a valid state. Apply this consistently to `webapp/src/app/[username]/page.tsx`, `webapp/src/app/series/[seriesId]/page.tsx`, and `webapp/src/app/[username]/reviews/page.tsx`.

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
