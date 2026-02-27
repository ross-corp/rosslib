# Features

Backlog of small tasks for nephewbot to pick off. Each item should be self-contained and implementable without external coordination. Items are ordered by priority — nephewbot picks the top unchecked item.

## quick wins

## stats & data

## notifications & feed

## profile & social

## search & browse

## book detail & discovery

## settings & account

- [ ] Improve settings page navigation. In `webapp/src/app/settings/page.tsx` (lines 43–56), links to Export and Ghost Activity are tucked into the header as small text. There's no link to Import at all. Add a horizontal nav bar below the "Settings" heading with links to: Profile (current page), Import (`/settings/import`), Export (`/settings/export`), and Ghost Activity (`/settings/ghost-activity`). Use a simple pill/tab style with the current section highlighted. This makes all settings sections discoverable from any settings page. Apply the same nav bar to each settings sub-page by extracting it into a shared `SettingsNav` component in `webapp/src/components/settings-nav.tsx`.

- [ ] Add a "Followed Books" management page. The API endpoint `GET /me/followed-books` (registered in `api/main.go` line 152) returns the user's followed books, and `DELETE /books/{workId}/follow` unfollows. But there is no webapp page to view or manage these. **Steps**: (1) Create proxy route `webapp/src/app/api/me/followed-books/route.ts` that forwards to `GET /me/followed-books`. (2) Create page at `webapp/src/app/settings/followed-books/page.tsx` — fetch the list, display each book with cover image, title, author, and an "Unfollow" button (calls `DELETE /api/books/{workId}/follow`). (3) Add a "Followed Books" link in the settings navigation (or at minimum link from the main settings page).

- [ ] Add a pending imports review page at `/settings/imports/pending`. The API has three endpoints in `api/handlers/pending_imports.go`: `GetPendingImports` (GET, returns list with title, author, isbn13, status), `ResolvePendingImport` (PATCH, links an unmatched row to a book), and `DeletePendingImport` (DELETE, dismisses a row). Create a new Next.js page at `webapp/src/app/settings/imports/pending/page.tsx`. It should: (1) fetch pending imports via `GET /api/me/imports/pending`, (2) show each row with title, author, ISBN, and status, (3) provide a "Search & Link" button per row that opens a book search modal — when a book is selected, call `PATCH /api/me/imports/pending/:id` with the OL work ID, (4) provide a "Dismiss" button that calls `DELETE /api/me/imports/pending/:id`. Add a link to this page from the import page (`/settings/import`). Add the Next.js proxy routes if they don't exist.

## UX polish

- [ ] Add color-coded genre cards on the genres page. In `webapp/src/app/genres/page.tsx`, all 12 genre cards are visually identical — just a name and count in a plain border box. Assign each genre a distinct background color (muted pastels or gradients) to make the page visually scannable and more engaging as a discovery surface.

## blocked

- [ ] Populate series data from Open Library during book lookup. **BLOCKED: depends on PR #60 (series metadata) being merged first.** Once the `series` and `book_series` collections exist, update `GetBookDetail` in `api/handlers/books.go` to auto-detect series data. The OL editions response (`/works/{workId}/editions.json`) includes a `series` array on some editions. For each edition entry, check for a `series` field. If found, find-or-create a `series` record by name and create a `book_series` link with the position number (if available). Also try the OL work's `subjects` array for series-like patterns (e.g. "Harry Potter" appearing as a subject). This is best-effort — not all OL works have series data. Log when series data is found vs. not for visibility into coverage.

## Pending PRs

<!-- nephewbot moves tasks here when it opens a PR. Move to docs/planning/completed.md after merging. -->
- [Replace raw OL ID input with book search dropdown](https://github.com/ross-corp/rosslib/pull/83) — Search-as-you-type dropdown in Suggest Link form using /api/books/search
- [Show book subjects as tags on book detail page](https://github.com/ross-corp/rosslib/pull/82) — Extract OL subjects and render as pill tags below description
- [Add review likes](https://github.com/ross-corp/rosslib/pull/51) — Toggle like on reviews with heart icon, notifications, and activity recording
- [Add user blocking](https://github.com/ross-corp/rosslib/pull/52) — Block/unblock users with review/search/feed filtering and profile UI
- [Add reading goals](https://github.com/ross-corp/rosslib/pull/53) — Annual reading goal with progress tracking, profile card, and settings form
- [Add @mention notifications in thread comments](https://github.com/ross-corp/rosslib/pull/54) — @username in comments creates thread_mention notifications and renders as profile links
- [Add recommend-to-friend feature](https://github.com/ross-corp/rosslib/pull/55) — Recommend books to users with modal, notifications, and /recommendations page
- [Add detailed reading statistics](https://github.com/ross-corp/rosslib/pull/56) — GET /users/:username/stats with books by year/month, rating distribution, and /[username]/stats page
- [Add reading timeline view](https://github.com/ross-corp/rosslib/pull/57) — GET /users/:username/timeline with year/month grouping and /[username]/timeline page
- [Add content reporting](https://github.com/ross-corp/rosslib/pull/58) — Reports collection with flag icons on reviews/comments/links, modal submission, and admin review panel
- [Add notification preferences](https://github.com/ross-corp/rosslib/pull/59) — Per-user notification toggle switches on settings page with GET/PUT API and ShouldNotify helper
- [Add series metadata](https://github.com/ross-corp/rosslib/pull/60) — Series and book_series collections, API endpoints, book detail series badges, /series/:id page, shelf grid position badges
- [Add StoryGraph CSV import](https://github.com/ross-corp/rosslib/pull/61) — StoryGraph preview/commit endpoints, status mapping, tag import, tabbed import page
- [Fix homepage feature grid responsiveness](https://github.com/ross-corp/rosslib/pull/62) — Add responsive breakpoints (1-col mobile, 2-col sm, 3-col lg) to feature grid
- [Add rating validation to AddBook and UpdateBook](https://github.com/ross-corp/rosslib/pull/64) — Validate rating is in range 1-5 with clear 400 error
- [Add max-length validation to threads and comments](https://github.com/ross-corp/rosslib/pull/65) — Title max 500 chars, body max 10k chars, comment max 5k chars with clear 400 errors
- [Add max-length validation to profile fields](https://github.com/ross-corp/rosslib/pull/66) — display_name max 100 chars, bio max 2000 chars with clear 400 errors
- [Populate friends_count on profile endpoint](https://github.com/ross-corp/rosslib/pull/67) — Count mutual follows (friends) instead of hardcoded 0
- [Populate books_this_year on profile endpoint](https://github.com/ross-corp/rosslib/pull/68) — Count finished books with date_read in current calendar year
- [Populate page_count and publisher from local book records](https://github.com/ross-corp/rosslib/pull/69) — Add migration for page_count/publisher columns and populate from local data in GetBookDetail
- [Return edition_count in GetBookDetail response](https://github.com/ross-corp/rosslib/pull/70) — Fetch edition count from OL editions endpoint and include in book detail JSON
- [Populate book stats for local search results](https://github.com/ross-corp/rosslib/pull/71) — Batch-fetch book_stats for local results to populate average_rating, rating_count, and already_read_count
- [Add distinct icons per notification type](https://github.com/ross-corp/rosslib/pull/72) — Type-specific SVG icons for notification types (book, chat, link, star) with bell fallback
- [Add click-to-mark-read on individual notifications](https://github.com/ross-corp/rosslib/pull/73) — Click unread dot to mark single notification as read with optimistic UI update
- [Show reading progress on other users' profiles](https://github.com/ross-corp/rosslib/pull/74) — Include progress_pages, progress_percent, page_count in GetUserBooks grouped status response
- [Add Want to Read section to profile page](https://github.com/ross-corp/rosslib/pull/75) — Render want-to-read book covers grid on user profiles between Currently Reading and Favorites
- [Add empty states for profile sections on own profile](https://github.com/ross-corp/rosslib/pull/76) — Show helpful empty state messages for Currently Reading, Recent Reviews, and Recent Activity on own profile
- [Add followers/following list pages](https://github.com/ross-corp/rosslib/pull/77) — Followers/following API endpoints with privacy checks, paginated list pages with user cards and follow buttons
- [Add user avatars to People search results](https://github.com/ross-corp/rosslib/pull/78) — Render avatar images with letter fallback on People search tab
- [Add empty/landing state to search page](https://github.com/ross-corp/rosslib/pull/79) — Show prompt message and popular books grid when no query is entered
- [Add sort options to browse-all-users page](https://github.com/ross-corp/rosslib/pull/80) — Sort dropdown (Newest, Most books, Most followers) on /users page
- [Link author names to author pages on book detail](https://github.com/ross-corp/rosslib/pull/81) — Author names link to /authors/{key} pages, with plain text fallback for local-only records
- [Add pagination to author works grid](https://github.com/ross-corp/rosslib/pull/84) — Paginated author works with Show more button (24 per page)
