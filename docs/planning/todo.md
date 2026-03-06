# Features

Backlog of small tasks for nephewbot to pick off. Each item should be self-contained and implementable without external coordination. Items are ordered by priority — nephewbot picks the top unchecked item.

## profile & social

## search & browse

- [ ] Add genre filter chips to genre detail page: the genre detail page `webapp/src/app/genres/[slug]/page.tsx` shows books for a single genre but has no way to further filter. Add a sort dropdown (title A-Z, rating, year) that passes the `sort` query param to `GET /genres/:slug/books`. The API already supports `sort=title|rating|year` per the docs.

## book detail & discovery

- [ ] Add series navigation links on book detail page: on `webapp/src/app/books/[workId]/page.tsx`, when a book belongs to a series, show "Previous in series" and "Next in series" navigation links. The `GET /series/:seriesId` response already includes all books with positions. Fetch the series data (already available from `GET /books/:workId` response's `series` array), find adjacent positions, and render small cover thumbnails with links for prev/next books.
- [ ] Add follower count display on book detail page: the `GET /books/:workId/followers/count` endpoint exists and returns `{ follower_count: N }` but no frontend component displays it. On `webapp/src/app/books/[workId]/page.tsx`, fetch the follower count server-side and display it near the existing read/want-to-read counts (e.g., "12 people following this book"). The API is public (no auth needed).

## settings & account

- [ ] Add "Followed books" link to settings nav: the page `webapp/src/app/settings/followed-books/page.tsx` exists but `webapp/src/components/settings-nav.tsx` may not include a link to it. Verify that SettingsNav includes "Followed books" and "Followed authors" in its pill list. If missing, add them with hrefs `/settings/followed-books` and `/settings/followed-authors`.
- [ ] Add email change form to settings page: the `PUT /me/email` API endpoint exists (documented in api.md) with proper validation (requires current password, checks duplicates, resets email_verified). The `EmailForm` component exists at `webapp/src/components/email-form.tsx`. Verify it is rendered on the settings page `webapp/src/app/settings/page.tsx` — if not, import and render it below the PasswordForm in the account section.

## UX polish

- [ ] Add `role="dialog"` and `aria-modal="true"` to modal overlays: several components render modal overlays without proper ARIA dialog semantics. Add `role="dialog"`, `aria-modal="true"`, and `aria-labelledby` (pointing to the modal title) to the modal container `<div>` in: `webapp/src/components/pending-imports-manager.tsx` (~line 233), `webapp/src/components/recommend-button.tsx` (~line 109), `webapp/src/components/report-modal.tsx` (modal container), and `webapp/src/components/edition-picker.tsx` (modal container).
- [ ] Add `aria-label` to icon-only buttons in BookLinkList: in `webapp/src/components/book-link-list.tsx`, the upvote, edit (pencil), and delete (trash) icon buttons lack `aria-label` attributes. Add descriptive labels like `aria-label="Upvote link"`, `aria-label="Edit link"`, `aria-label="Delete link"` to each icon button so screen readers can identify their purpose.
- [ ] Add `role="tab"` and `aria-selected` to CompareTabs: in `webapp/src/components/compare-tabs.tsx` (lines 17-36), the "My Lists" and "Compare with a Friend" tab buttons lack proper ARIA tab semantics. Add `role="tablist"` to the parent, `role="tab"` and `aria-selected={isActive}` to each button, and `role="tabpanel"` with `aria-labelledby` to the content area.
- [ ] Add `aria-label` and `aria-valuenow` to genre rating sliders: in `webapp/src/components/genre-rating-editor.tsx` (~lines 183-191), the range input sliders for genre ratings lack `aria-label` (should be the genre name, e.g. "Fiction rating") and `aria-valuenow` (current slider value). Add both attributes so screen readers announce the genre being rated and its current value.

## data integrity

- [ ] Add input length validation to tag/label creation endpoints: `POST /me/tag-keys` in `api/handlers/tags.go` (~line 74) and `POST /me/tag-keys/:keyId/values` (~line 155) only check for empty names but have no maximum length check. Add a 100-character limit for tag key names and tag value names, returning 400 with "name must be 100 characters or fewer" if exceeded. Similarly, `POST /me/shelves` in `api/handlers/collections.go` (~line 92) should validate `name` length (max 255 characters to match the DB column).
- [ ] Add input length validation to feedback and report submissions: `POST /feedback` in `api/handlers/feedback.go` validates type and requires title/description but has no length limits. Add max lengths: title 500 chars, description 10000 chars, steps_to_reproduce 5000 chars. `POST /reports` in `api/handlers/reports.go` should limit `details` to 2000 chars. Return 400 with descriptive error messages when limits are exceeded.
- [ ] Add `reviewer_comment` and `reviewed_at` to admin link edit review: the `PUT /admin/link-edits/:editId` handler in `api/handlers/links.go` should already set `reviewer_comment` from the request body and `reviewed_at` to `time.Now()`. Verify these fields are being set (they depend on the migration fix in the BUGS section). On the frontend, `webapp/src/components/admin-link-edits.tsx` should display `reviewer_comment` and `reviewed_at` for reviewed edits — verify and add if missing.

## API gaps

- [ ] Add pagination to `GET /books/:workId/threads` endpoint: the thread listing in `api/handlers/threads.go` returns all threads for a book with no pagination. Add `page` and `limit` query params (default limit 20, max 100) with SQL `LIMIT` and `OFFSET`. Return `total` count in the response. Update the thread list on the book detail page `webapp/src/app/books/[workId]/page.tsx` (or the ThreadList component) to support "Load more" or page-based navigation.
- [ ] Add pagination to `GET /books/:workId/links` endpoint: community links in `api/handlers/links.go` returns all links for a book with no pagination. For books with many community links, add `limit` and `offset` query params (default limit 50). Return `total` count. Update `webapp/src/components/book-link-list.tsx` to show a "Show more" button when more links are available.
- [ ] Add `GET /me/feedback` and `DELETE /me/feedback/:feedbackId` proxy routes: the API endpoints exist in `api/handlers/feedback.go` and are registered in `api/main.go`, and the frontend page exists at `webapp/src/app/settings/feedback/page.tsx` with `FeedbackList` component. Verify the Next.js proxy routes exist at `webapp/src/app/api/me/feedback/route.ts` (GET) and `webapp/src/app/api/me/feedback/[feedbackId]/route.ts` (DELETE). If missing, create them following the standard proxy pattern (extract token cookie, forward as Bearer header to API_URL).

## series improvements

- [ ] Add series search: create a `GET /series/search?q=<name>` endpoint in `api/handlers/series.go` that searches the `series` collection by name using `LIKE` matching. Return `{ results: [{ id, name, description, book_count }] }` with up to 20 results. Add a corresponding Next.js proxy route at `webapp/src/app/api/series/search/route.ts`. This enables the "Add to series" flow on book detail pages to search existing series instead of only creating new ones by exact name match.
- [ ] Add series page link in nav or browse menu: the series detail page exists at `webapp/src/app/series/[seriesId]/page.tsx` but there's no way to browse or discover series from the navigation. On the book detail page, when a book has series memberships, make the series name a clickable link to `/series/:seriesId` (verify this is already done). On the author page, if any of the author's works belong to series, show a "Series" section listing unique series with links.

## import improvements

- [ ] Add retry button for unmatched imports on pending imports page: the `PendingImportsManager` component at `webapp/src/components/pending-imports-manager.tsx` shows unmatched imports with Search, Dismiss, and Delete actions. Add a "Retry lookup" button that calls `POST /me/import/goodreads/preview` with just the single row's title/author/ISBN to re-attempt the lookup chain (including LLM fuzzy matching if ANTHROPIC_API_KEY is set). If a match is found, auto-populate the resolve modal with the result.

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
