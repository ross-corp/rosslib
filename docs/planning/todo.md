# Features

Backlog of small tasks for nephewbot to pick off. Each item should be self-contained and implementable without external coordination. Items are ordered by priority — nephewbot picks the top unchecked item.


## BUGS

- [ ] Implement password reset and email verification backend: the webapp proxy routes exist (`POST /auth/forgot-password`, `POST /auth/reset-password`, `POST /auth/verify-email`, `POST /auth/resend-verification`) but the Go handler functions and backing DB tables (`password_reset_tokens`, `email_verification_tokens`) are entirely missing. The `email_verified` field exists on users. Requires: new migration for token tables, `api/internal/email` package for SMTP, four handler functions in `api/handlers/auth.go`, route registration in `api/main.go`. See `docs/planning/completed.md` "Password Reset" and "Email Verification" sections for full spec. **Large task — not suitable for nephewbot.**
- [ ] Implement set-operation and cross-user-compare backend: the webapp proxy routes exist (`POST /me/shelves/set-operation`, `POST /me/shelves/set-operation/save`, `POST /me/shelves/cross-user-compare`, `POST /me/shelves/cross-user-compare/save`) but no Go handler functions exist in `api/handlers/collections.go`. Requires: new `computed_collections` table migration, four handler functions, route registration. See `docs/planning/completed.md` "Computed Lists" sections for full spec. **Large task — not suitable for nephewbot.**

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

## Loading States

- [ ] Add `loading.tsx` for the users browse page: create `webapp/src/app/users/loading.tsx` that shows a grid of placeholder cards (8 items, each with a gray circle for avatar and two gray bars for name/stats) matching the layout in `users/page.tsx`. Use simple `animate-pulse` div placeholders.
- [ ] Add `loading.tsx` for the feed page: create `webapp/src/app/feed/loading.tsx` that shows 5 placeholder activity cards (gray bars for avatar, name, and description) matching the `ActivityCard` layout. Use `animate-pulse` div placeholders.
- [ ] Add `loading.tsx` for the stats page: create `webapp/src/app/[username]/stats/loading.tsx` that shows placeholder bar charts (5 summary cards at top with gray squares, then placeholder horizontal bars) matching the stats page layout. Use `animate-pulse` div placeholders.
- [ ] Add `loading.tsx` for the timeline page: create `webapp/src/app/[username]/timeline/loading.tsx` that shows a placeholder year heading and 3 month sections each with 4 gray book cover rectangles. Use `animate-pulse` div placeholders.

## UX Polish

- [ ] Add a "View timeline" link to the profile page: on `webapp/src/app/[username]/page.tsx`, add a link below the "Reading Stats →" section heading that says "Reading Timeline →" and links to `/${username}/timeline`. Currently the timeline page exists but is only discoverable via direct URL — there's no link to it anywhere in the UI.
- [ ] Add a "View stats" link to the timeline page: on `webapp/src/app/[username]/timeline/page.tsx`, add a link next to the back arrow (line 75) that says "Stats" and links to `/${username}/stats`. These two pages are complementary but neither links to the other.
- [ ] Show followed authors on visitor profile pages: on `webapp/src/app/[username]/page.tsx`, the "Followed Authors" sidebar section (lines 678-702) is only shown when `isOwnProfile` is true. Make it visible to visitors too by fetching the target user's followed authors via `GET /users/${username}/followed-authors` (a new public endpoint that returns author_key + author_name for any user, respecting privacy). Add the Go handler in `api/handlers/users.go` registered as `GET /users/{username}/followed-authors` (no auth required, respects `is_private`), and a webapp proxy route at `webapp/src/app/api/users/[username]/followed-authors/route.ts`.
- [ ] Show reading goal on visitor profile pages: on `webapp/src/app/[username]/page.tsx`, the `ReadingGoalCard` (line 607-611) is already shown for visitors because `fetchGoal` uses the public endpoint `GET /users/:username/goals/:year`. However, there's no link to set a goal visible on the own-profile view aside from settings. Add a small "Set a reading goal" link in the Reading Stats section (near line 615) when `isOwnProfile && !readingGoal` that links to `/settings#reading-goal`.
- [ ] Add page title metadata to the feed page: `webapp/src/app/feed/page.tsx` has no `metadata` export. Add `export const metadata = { title: "Feed" }` so the browser tab shows "Feed" instead of the default app title.
- [ ] Add page title metadata to the genres index page: `webapp/src/app/genres/page.tsx` has no `metadata` export. Add `export const metadata = { title: "Genres" }` so the browser tab shows "Genres".
- [ ] Add page title metadata to the notifications page: `webapp/src/app/notifications/page.tsx` — check if it has a `metadata` export and add `export const metadata = { title: "Notifications" }` if missing.
- [ ] Add page title metadata to the recommendations page: `webapp/src/app/recommendations/page.tsx` — add `export const metadata = { title: "Recommendations" }` if missing.
- [ ] Add `aria-label` attributes to icon-only buttons in `ShelfPicker` (`webapp/src/components/shelf-picker.tsx`): the dropdown toggle button and the remove button inside the shelf picker likely use icon-only rendering without accessible labels. Add `aria-label="Change reading status"` to the main toggle and `aria-label="Remove from library"` to the remove action. Also add `aria-haspopup="listbox"` and `aria-expanded={isOpen}` to the toggle button.

## API & Performance

- [ ] Add an index on `thread_comments.user` column: create a new PocketBase migration in `api/migrations/` that adds an index on the `user` column of the `thread_comments` collection. This column is queried when looking up a user's comment history but has no index. Use the pattern from existing migrations — `collection.Indexes` append and `app.Save(collection)`.
- [ ] Add an index on `book_links.user` column: create a new PocketBase migration in `api/migrations/` that adds an index on the `user` column of the `book_links` collection. This supports queries for "links submitted by this user" but currently has no index.
- [ ] Add an index on `follows.followee` column: create a new PocketBase migration in `api/migrations/` that adds a simple index on the `followee` column of the `follows` collection. The existing unique index on `(follower, followee)` doesn't efficiently serve reverse lookups (e.g., "who follows this user?") because `followee` is the second column. A standalone `followee` index speeds up follower list queries and follower count.
- [ ] Add pagination to `GET /users/:username/activity` response: the endpoint in `api/handlers/activity.go` already supports cursor-based pagination, but the profile page in `webapp/src/app/[username]/page.tsx` fetches all activities without a limit. Pass `?limit=10` in the `fetchRecentActivity` call (line 154) to avoid loading the entire activity history for users with many activities.

## Profile & Social

- [ ] Show followed books on user profile sidebar: on `webapp/src/app/[username]/page.tsx`, add a "Followed Books" section below the "Followed Authors" sidebar section (after line 702). For the profile owner, fetch via `GET /me/followed-books?limit=5` (already exists). Show up to 5 book covers as small thumbnails with titles, and a "View all N →" link to `/settings/followed-books` when there are more than 5. This surfaces a feature that currently exists but is buried in settings.
- [ ] Show label/tag collections on visitor profile pages as a browsable sidebar: on `webapp/src/app/[username]/page.tsx`, the custom tag collections (line 272-274) and label keys (line 280) are computed but only shown as a count in the library summary link (line 481-495). Add a "Collections" section below the main content area that lists tag collections (name + count) and label keys (name + values count) as clickable links to `/${username}/tags/${slug}` and `/${username}/labels/${keySlug}/${valueSlug}` respectively, so visitors can browse the user's organizational system.

## Search & Browse

- [ ] Add book count next to genre name on genre detail page breadcrumb: on `webapp/src/app/genres/[slug]/page.tsx`, the page heading shows the genre name but the total book count is only shown in the pagination info. Add the total count next to the heading, e.g. "Fiction (142 books)", matching how the author page shows work count next to the author name.
- [ ] Add sort options to the genre detail page: on `webapp/src/app/genres/[slug]/page.tsx`, the API endpoint `GET /genres/:slug/books` already supports `sort=title|rating|year` query params (documented in `docs/documentation/api.md`), but the frontend page doesn't expose sort controls. Add a sort dropdown with options "Title (A-Z)", "Highest Rated", and "Newest" that updates the URL's `sort` query param, matching the pattern used on the users browse page (`webapp/src/app/users/page.tsx` sort dropdown).

## Book Detail & Discovery

- [ ] Add "Back to search" link on book detail page when arrived from search: on `webapp/src/app/books/[workId]/page.tsx`, there is no back-navigation to return to search results. Add a breadcrumb-style `← Back to search` link at the top of the page when the referrer is `/search`. Use `searchParams` to accept an optional `from=search&q=<query>` parameter, and when present, show a link back to `/search?q=<query>`. Update search result links in `webapp/src/app/search/page.tsx` to append `?from=search&q=${query}` to book detail hrefs.
- [ ] Show series position badge on book covers throughout the app: when a book has `series_position` data, display a small numbered badge (e.g., "#3") on the cover thumbnail in `BookCoverRow` (`webapp/src/components/book-cover-row.tsx`). The component already receives `series_position` in the book data (line 24 of `[username]/page.tsx` type definition). Add a small absolute-positioned badge in the top-right corner of the cover when `book.series_position` is set, using `text-xs bg-surface-2 border border-border rounded px-1`.
