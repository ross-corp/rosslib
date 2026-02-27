# Features

Backlog of small tasks for nephewbot to pick off. Each item should be self-contained and implementable without external coordination. Items are ordered by priority — nephewbot picks the top unchecked item.

## quick wins

## stats & data

## notifications & feed

## profile & social

## search & browse

- [ ] Add sort options to the browse-all-users page. In `webapp/src/app/users/page.tsx`, users are listed in registration order (`-created`). Add a sort dropdown with options: "Newest", "Most books", "Most followers". Pass the sort parameter to the `GET /users/search` endpoint (may need to add sort support to the API handler in `api/handlers/users.go` `SearchUsers`).

## book detail & discovery

- [ ] Link author names to their author pages on the book detail page. In `api/handlers/books.go` (lines 254–269), the `GetBookDetail` handler fetches each author's data from OL (via `ol.get(key + ".json")`) and collects names into `authorNames []string`. The author key is already available in the loop at line 258 (e.g. `/authors/OL23919A`). Change `authorNames` from `[]string` to `[]map[string]string` and collect both `name` and `key` (strip the `/authors/` prefix). Update the response JSON on line 305: `"authors": authorNames` now returns `[{"name": "Author", "key": "OL23919A"}, ...]`. In the local fallback (line 278), return `[{"name": "Author", "key": null}, ...]` since local records don't store author keys. In `webapp/src/app/books/[workId]/page.tsx`, update the `BookDetail` type's `authors` from `string[]` to `{name: string, key: string | null}[]`. Render each author as `<Link href={"/authors/" + a.key}>` when `key` is non-null, plain text otherwise. The `/authors/[authorKey]` page already exists.

- [ ] Show book subjects as tags on the book detail page. First, add `subjects` to the `GetBookDetail` response in `api/handlers/books.go`. The OL work response (`workData`) contains a `"subjects"` array of strings — extract up to 10 and include as `"subjects": [...]` in the response JSON (line 302). Also fall back to `localBooks[0].GetString("subjects")` (comma-separated) if OL data is unavailable. In `webapp/src/app/books/[workId]/page.tsx`, add `subjects: string[]` to the `BookDetail` type. Render them as small pill/chip tags (rounded, muted background, small text) below the book description. Each subject can link to `/genres/{slug}` if it matches a known genre, or to `/search?subject={encoded}` otherwise. These are read-only OL-sourced tags, distinct from user-assigned tags.

- [ ] Replace the raw OL work ID input in the "Suggest Link" form with a book search dropdown. In `webapp/src/components/book-link-list.tsx` (~lines 205–213), the related books form asks for a raw Open Library work ID string. Replace this with a search-as-you-type input that queries `/books/search`, shows matching results in a dropdown, and uses the selected book's OL key. This makes the feature actually usable by normal users.

- [ ] Add pagination to the author works grid. In `webapp/src/app/authors/[authorKey]/page.tsx`, all works are rendered at once with no pagination. For prolific authors this can be hundreds of entries. Add a "Show more" button or paginated grid (e.g. show first 24, load more on click).

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
