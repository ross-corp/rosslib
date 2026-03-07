# Webapp Architecture

The webapp is a Next.js 15 app using the App Router. It runs on `:3000` and proxies all Go API calls through its own route handlers so the browser never talks directly to `:8080`.

---

## Key conventions

### Proxy pattern

Every client-side API call goes through a Next.js route handler in `webapp/src/app/api/`. These handlers:

1. Pull the `token` cookie from the request
2. Forward it to the Go API as `Authorization: Bearer <token>`
3. Return the response as-is

```
Browser → POST /api/shelves/:id/books
       → Next.js handler (adds auth header)
       → Go API POST /shelves/:id/books
```

This keeps auth cookies httpOnly and the Go API URL server-side only.

When adding a new Go API call that client components need to make, add a corresponding Next.js route handler. Match the path structure of the Go API where possible.

### Server vs client components

Server components (no `"use client"`) handle data fetching and pass data as props:

```tsx
// Server component — fetches data at request time
export default async function ShelfPage({ params }) {
  const shelf = await fetchShelf(username, slug);   // direct fetch to Go API
  return <LibraryManager initialBooks={shelf.books} ... />;
}
```

Client components (`"use client"`) handle interaction, state, and browser API calls:

```tsx
"use client";
// Receives initial data from server, manages local state, calls /api/* routes
export default function LibraryManager({ initialBooks, ... }) {
  const [books, setBooks] = useState(initialBooks);
  // ...
}
```

Server components call `${process.env.API_URL}` directly (server-side env var). Client components call `/api/...` proxy routes (relative URL, works in browser).

### Auth helpers

`webapp/src/lib/auth.ts` exports two server-side helpers:

- `getUser()` — decodes the JWT cookie and returns `{ user_id, username }` or null
- `getToken()` — returns the raw JWT string or null

These are only usable in server components and route handlers (they use `next/headers`).

---

## Page structure

```
webapp/src/app/
├── layout.tsx                      root layout (Nav is NOT here — included per-page)
├── page.tsx                        home / landing
├── login/page.tsx
├── register/page.tsx
├── forgot-password/page.tsx          forgot password (request reset link)
├── reset-password/page.tsx           set new password (from email link)
├── search/page.tsx                 book + user + author search (shows popular books when no query)
├── feed/page.tsx                   activity feed with filter chips (All, Reviews, Ratings, Status Updates, Threads, Social)
├── users/page.tsx                  browse all users (sort by newest/books/followers)
├── books/[workId]/page.tsx         single book page (series badges, series navigation row, review sort dropdown)
├── books/[workId]/not-found.tsx   "Book not found" with search link
├── series/[seriesId]/page.tsx     series detail — ordered book list with covers, reading progress & status picker
├── settings/
│   ├── page.tsx                    profile settings
│   ├── import/page.tsx             CSV import (Goodreads, StoryGraph & LibraryThing tabs)
│   ├── imports/pending/page.tsx   review unmatched imports
│   ├── export/page.tsx             CSV export
│   ├── tags/page.tsx               label category management
│   ├── ghost-activity/page.tsx     ghost user controls
│   ├── feedback/page.tsx           view and delete submitted feedback
│   ├── follow-requests/page.tsx    pending follow requests
│   ├── followed-books/page.tsx     manage followed books
│   ├── followed-authors/page.tsx   manage followed authors
│   └── blocked/page.tsx            manage blocked users
├── scan/page.tsx                   ISBN barcode scanner
├── library/compare/page.tsx        compare lists (set operations)
├── notifications/page.tsx          notification center
├── recommendations/page.tsx       book recommendations (Received / Sent tabs)
├── feedback/page.tsx              bug report & feature request form
├── admin/page.tsx                 admin panel (moderator only)
├── [username]/
│   ├── not-found.tsx               "User not found" with link to /users
│   ├── page.tsx                    public profile (incl. computed lists, followed authors sidebar, favorite genre chips)
│   ├── stats/page.tsx              detailed reading statistics
│   ├── shelves/[slug]/page.tsx     label page (owner gets library manager)
│   ├── followers/page.tsx          followers list
│   ├── following/page.tsx          following list
│   ├── tags/[...path]/page.tsx     tag browsing page
│   ├── labels/[keySlug]/[...valuePath]/page.tsx   label browsing page (nested)
│   ├── reviews/page.tsx           paginated reviews list (?page=N)
│   ├── timeline/page.tsx          reading timeline (books by month/year)
│   └── year-in-review/page.tsx   year-in-review summary (stats, top books, genres)
└── api/                            Next.js proxy route handlers
    ├── auth/login/route.ts
    ├── auth/register/route.ts
    ├── auth/logout/route.ts
    ├── auth/google/route.ts              ← GET redirects to Google consent screen
    ├── auth/google/callback/route.ts     ← GET exchanges code, calls API, sets cookies
    ├── auth/forgot-password/route.ts       ← POST request reset email
    ├── auth/reset-password/route.ts        ← POST reset password with token
    ├── users/me/route.ts
    ├── users/[username]/follow/route.ts
    ├── me/shelves/route.ts
    ├── me/shelves/set-operation/route.ts
    ├── me/shelves/set-operation/save/route.ts
    ├── me/shelves/cross-user-compare/route.ts
    ├── me/shelves/cross-user-compare/save/route.ts
    ├── users/[username]/shelves/route.ts
    ├── me/tag-keys/route.ts
    ├── me/tag-keys/[keyId]/route.ts
    ├── me/tag-keys/[keyId]/values/route.ts
    ├── me/tag-keys/[keyId]/values/[valueId]/route.ts
    ├── me/books/[olId]/tags/route.ts
    ├── me/books/[olId]/tags/[keyId]/route.ts
    ├── me/books/[olId]/tags/[keyId]/values/[valueId]/route.ts
    ├── me/import/goodreads/preview/route.ts
    ├── me/import/goodreads/commit/route.ts
    ├── me/import/storygraph/preview/route.ts
    ├── me/import/storygraph/commit/route.ts
    ├── me/imports/pending/route.ts               ← GET list pending imports
    ├── me/imports/pending/[id]/route.ts           ← PATCH resolve/dismiss, DELETE
    ├── shelves/[shelfId]/books/route.ts
    ├── shelves/[shelfId]/books/[olId]/route.ts    ← GET, PATCH, DELETE
    ├── books/[workId]/series/route.ts              ← GET, POST series memberships
    ├── series/[seriesId]/route.ts                 ← GET series detail, PUT update description
    ├── books/[workId]/readers/route.ts                ← GET friends reading this book
    ├── books/[workId]/reviews/[userId]/like/route.ts ← POST toggle, GET check review like
    ├── books/[workId]/links/route.ts              ← GET, POST community links
    ├── links/[linkId]/route.ts                    ← DELETE community link
    ├── links/[linkId]/vote/route.ts               ← POST, DELETE vote on link
    ├── links/[linkId]/edits/route.ts              ← POST propose link edit
    ├── feedback/route.ts                           ← POST submit feedback
    ├── reports/route.ts                            ← POST submit content report
    ├── admin/feedback/route.ts                    ← GET list feedback (admin)
    ├── admin/feedback/[feedbackId]/route.ts       ← PATCH toggle feedback status (admin)
    ├── admin/reports/route.ts                     ← GET list reports (admin)
    ├── admin/reports/[reportId]/route.ts          ← PATCH review/dismiss report (admin)
    ├── admin/users/route.ts                       ← GET admin user list
    ├── admin/users/[userId]/moderator/route.ts    ← PUT grant/revoke moderator
    ├── admin/link-edits/route.ts                  ← GET list link edits
    ├── admin/link-edits/[editId]/route.ts         ← PUT approve/reject link edit
    ├── books/scan/route.ts                            ← POST barcode scan
    ├── books/lookup/route.ts                          ← GET ISBN lookup
    ├── books/[workId]/followers/count/route.ts        ← GET book follower count (public)
    ├── books/[workId]/genre-ratings/route.ts         ← GET aggregate genre ratings
    ├── me/books/[olId]/genre-ratings/route.ts       ← GET, PUT user genre ratings
    ├── me/account/route.ts                         ← GET account info (has_password, has_google)
    ├── me/account/data/route.ts                   ← DELETE all user data
    ├── me/avatar/route.ts                           ← POST upload avatar
    ├── me/banner/route.ts                           ← POST upload banner
    ├── me/password/route.ts                        ← PUT set/change password
    ├── me/notifications/route.ts                  ← GET list notifications
    ├── me/notifications/unread-count/route.ts     ← GET unread count
    ├── me/notifications/read-all/route.ts         ← POST mark all read
    ├── me/notifications/[notifId]/route.ts         ← DELETE delete notification
    ├── me/notifications/[notifId]/read/route.ts   ← POST mark one read
    ├── me/followed-authors/route.ts              ← GET followed authors list
    ├── me/notification-preferences/route.ts     ← GET, PUT notification prefs
    ├── me/recommendations/route.ts               ← GET, POST recommendations
    ├── me/recommendations/sent/route.ts          ← GET sent recommendations
    ├── me/recommendations/[recId]/route.ts       ← PATCH update recommendation status
    ├── me/saved-searches/route.ts                ← GET, POST saved searches
    ├── me/saved-searches/[id]/route.ts           ← DELETE saved search
    ├── users/route.ts                             ← GET search users
    └── users/[username]/
        ├── followers/route.ts                     ← GET followers list
        ├── following/route.ts                     ← GET following list
        ├── stats/route.ts                         ← GET reading statistics
        ├── tags/[...path]/route.ts
        ├── labels/[keySlug]/[...valuePath]/route.ts   ← catch-all for nested label paths
        ├── books/search/route.ts                   ← GET search within user's library
        ├── shelves/[slug]/route.ts                ← GET (for client-side label switching)
        ├── timeline/route.ts                      ← GET reading timeline
        └── year-in-review/route.ts                ← GET year-in-review summary
```

---

## Key components

### `Nav` (`components/nav.tsx`)

Top navigation bar. Server component that fetches the current user. Links are organized into two dropdown menus: **Browse** (Search books, Genres, Scan ISBN) and **Community** (Browse users, My feed). Notification bell, admin link, user avatar, and sign out remain as standalone items. On desktop (`md:` and above), the search bar and nav links are shown inline. Below `md:`, the search bar and desktop nav are hidden and replaced by the `MobileNav` hamburger menu.

### `MobileNav` (`components/mobile-nav.tsx`)

Client component rendered inside `Nav`, visible only below the `md:` breakpoint. Shows a hamburger button (`☰`) that toggles a full-width dropdown panel with all nav links stacked vertically, grouped under "Browse" and "Community" section headings. Includes notification, admin, profile, and sign out links for authenticated users, or sign in/sign up for guests. Closes when a link is clicked or when clicking outside.

### `ThemeToggle` (`components/theme-toggle.tsx`)

Client component in the nav bar that cycles through `system` → `light` → `dark` themes. Stores the preference in `localStorage` (key: `rosslib-theme`) and applies it via `data-theme` attribute on `<html>`. For logged-in users, also persists the choice to the API via `PUT /api/me/theme`. Shows a sun icon in light mode, moon icon in dark mode, with an "auto" badge when set to system.

### `NavDropdown` (`components/nav-dropdown.tsx`)

Client component used by `Nav` for dropdown menus. Opens on hover (desktop) or click (mobile). Closes when clicking outside. Takes a `label` string and an array of `{ href, label }` items.

### `Pagination` (`components/pagination.tsx`)

Shared server component for consistent pagination across pages. Takes `prevHref` (string or null), `nextHref` (string or null), and an optional `label` string (e.g., "Page 2 of 5"). Renders Previous/Next links with matching border+rounded styling. Used on search, users, and reviews pages.

### `LibraryManager` (`components/library-manager.tsx`)

Full-page library manager rendered for label owners. Replaces the simple label grid on `[username]/shelves/[slug]` when `isOwner` is true.

Layout: `h-screen flex flex-col overflow-hidden` on the page, then inside LibraryManager:

```
┌─────────────────────────────────────────────────────┐
│ Nav                                                  │
├──────────┬──────────────────────────────────────────┤
│          │ top bar (label name / bulk action toolbar)│
│ sidebar  ├──────────────────────────────────────────┤
│          │                                          │
│ Labels   │   book cover grid (scrollable)           │
│ Custom   │                                          │
│ Tags     │                                          │
│          │                                          │
│          │                                          │
└──────────┴──────────────────────────────────────────┘
```

**Sidebar** — clicking a label fetches its books client-side via `GET /api/users/:username/shelves/:slug`. Clicking a tag collection fetches via `GET /api/users/:username/tags/:path`. Clicking a label value fetches via `GET /api/users/:username/labels/:keySlug/*valuePath` (includes sub-values). Nested label values are indented by depth in the sidebar, showing only the last path segment as the display name. A "+ New label" button at the bottom opens an inline form to create a custom label with a name and optional description (max 1000 chars).

**Search** — a search input in the top bar lets users search within the current library by title or author. Typing triggers a debounced (400ms) API call to `GET /api/users/:username/books/search?q=`. Results replace the displayed book grid while the search is active. Clearing the search restores the original view. Search state is also cleared when navigating to a different sidebar filter.

**Top bar** — shows the current label name, book count, description (if set), and sort options when nothing is selected. Custom labels (non-status, non-default) show an "Edit description" / "Add description" button. Sort options: Date added (default), Title, Author, Rating. Changing the sort re-fetches the current view from the API with the `sort` query param. Transforms into the bulk action toolbar when one or more books are checked:
- Rate — sets rating on all selected books via `PATCH /api/shelves/:shelfId/books/:olId`
- Move to label — moves via `POST /api/shelves/:targetId/books`, then refreshes the current label
- Labels — applies or clears a label value across all selected books via `PUT/DELETE /api/me/books/:olId/tags/:keyId`
- Remove — removes from current label via `DELETE /api/shelves/:shelfId/books/:olId`

Rate, Move, and Remove require a label context (disabled in tag-filtered views). Labels work in both label and tag views since they only need the `open_library_id`.

**Book grid** — cover images with a checkbox in the top-left. Checkboxes are invisible until hover or until at least one book is selected (at which point all checkboxes become visible). When books are selected, clicking a cover toggles selection instead of navigating to the book page.

### `ShelfBookGrid` (`components/shelf-book-grid.tsx`)

Simpler read-only-ish grid used on non-owner label views and the tag browsing page. Supports individual book removal (owner only) and the per-book `BookTagPicker`.

### `BookTagPicker` (`components/book-tag-picker.tsx`)

Dropdown for managing label assignments on a single book. Lazily loads current assignments on first open. Supports toggling predefined values and typing a free-form value.

### `QuickAddButton` (`components/quick-add-button.tsx`)

Compact overlay button shown on book covers in `ShelfBookGrid` and `BookCoverRow` when a logged-in user is viewing another user's profile or label page. Appears on hover in the bottom-right corner of the book cover. The main button triggers "Want to read" (adds via `POST /api/me/books` with the want-to-read status), and a dropdown arrow reveals all status options (Want to Read, Currently Reading, Finished, etc.) plus a "Rate & review" link that navigates to the book detail page. Requires `statusValues` and `statusKeyId` props (fetched from the viewer's `GET /me/tag-keys`).

### `ShelfPicker` (`components/shelf-picker.tsx`)

Dropdown for adding/moving/removing a single book from labels. Used on search results and book pages.

### `BookLinkList` (`components/book-link-list.tsx`)

Client component for community links (related books) on book detail pages. Shows links grouped by relationship type (sequel, prequel, companion, similar, etc.), sorted by upvote count. Logged-in users can upvote/unvote links, suggest new ones via an inline form, and propose edits to existing links (edit pencil icon). Proposed edits are submitted for moderator review. Target book is selected via a search-as-you-type dropdown that queries `/api/books/search` with 400ms debounce.

### `AdminUserList` (`components/admin-user-list.tsx`)

Client component for the `/admin` page. Provides a searchable, paginated table of all users with inline moderator toggle buttons. Moderators see a filled "Moderator" button; non-moderators see a "Grant" button. Search is debounced (300ms) and queries by username, display name, or email.

### `AdminFeedback` (`components/admin-feedback.tsx`)

Client component for the `/admin` page. Displays user-submitted bug reports and feature requests with status filter tabs (open/closed). Each item shows the type badge (bug/feature), severity badge for bugs, title, description, steps to reproduce (for bugs), and a close/reopen button. Calls `GET /api/admin/feedback` and `PATCH /api/admin/feedback/:id`.

### `AdminReports` (`components/admin-reports.tsx`)

Client component for the `/admin` page. Displays content reports submitted by users with status filter tabs (pending/reviewed/dismissed). Each report shows the content type badge (review/thread/comment/link), reason badge, reporter info, optional details, and a content preview. Pending reports have Review and Dismiss buttons. Calls `GET /api/admin/reports` and `PATCH /api/admin/reports/:id`.

### `ReportButton` (`components/report-button.tsx`)

Client component that renders a small flag icon button. When clicked, opens a `ReportModal`. Takes `contentType` ("review", "thread", "comment", or "link") and `contentId` props. Shown on reviews (book detail page), community links (BookLinkList), and thread comments (ThreadComments) for logged-in users viewing other users' content.

### `ReportModal` (`components/report-modal.tsx`)

Client component that renders a modal overlay for submitting content reports. Shows a reason radio group (spam, harassment, inappropriate, other) and an optional details textarea. Calls `POST /api/reports`. Displays success message and auto-closes on completion.

### `FeedbackForm` (`components/feedback-form.tsx`)

Client component for the `/feedback` page. Two-tab form for submitting bug reports or feature requests. Bug report tab includes title, description, steps to reproduce, and severity dropdown. Feature request tab includes title and description. Calls `POST /api/feedback`.

### `FeedbackList` (`app/settings/feedback/feedback-list.tsx`)

Client component for the `/settings/feedback` page. Displays the user's submitted bug reports and feature requests in a list. Each item shows type badge (Bug/Feature), status badge (Open/Closed), severity badge (for bugs), title, description preview (2-line clamp), and submission date. Each item has a "Delete" button that calls `DELETE /api/me/feedback/:id` and optimistically removes the item from the list.

### `AdminLinkEdits` (`components/admin-link-edits.tsx`)

Client component for the `/admin` page. Displays proposed community link edits with status filter tabs (pending/approved/rejected). Each edit shows the proposer, book pair, current vs. proposed values (type and note) side by side, and approve/reject buttons for pending edits. Reviewed edits show the reviewer name, date, and optional comment.

### `EditionPicker` (`components/edition-picker.tsx`)

Client component for selecting a specific edition of a book. Shown on the book detail page below the cover image when the user has the book in their library. Opens a modal listing all available editions (reusing the editions data from `GET /books/:workId/editions`) with cover thumbnails, format badges, publisher, and ISBN. Selecting an edition saves the edition key and cover URL to the user's `user_books` record via `PATCH /api/me/books/:olId`. The selected edition's cover is then displayed on the book detail page, profile pages, and shelf views instead of the default work cover.

### `AuthorBio` (`components/author-bio.tsx`)

Client component for the author detail page. Displays the author's bio text with automatic truncation. Bios longer than 500 characters are truncated with an ellipsis and a "Read more" button that toggles to show the full text. Short bios are displayed in full with no toggle.

### `AuthorWorksGrid` (`components/author-works-grid.tsx`)

Client component for the author detail page. Displays a paginated grid of the author's works (24 per page) with cover images and titles. A "Show more" button fetches the next page via `GET /api/authors/:authorKey/works?limit=24&offset=N`. Receives initial works and total count from the server component.

### `EditionList` (`components/edition-list.tsx`)

Read-only list of book editions shown in the "Editions" section of the book detail page. Displays edition cover thumbnails, format, publisher, publish date, page count, language, and ISBN. Supports "Show all" expansion and "Load more" pagination.

### `ReviewText` (`components/review-text.tsx`)

Renders review text with wikilink and markdown link support. Parses two inline link syntaxes:
- `[[Book Title]]` — rendered as a link to `/search?q=Book%20Title`
- `[Book Title](/books/OL123W)` — rendered as a direct link to the book page

Used on book detail pages (community reviews), user reviews pages, recent reviews on profiles, and the collapsed review view in the book review editor. The companion `BookReviewEditor` component provides `[[` autocomplete that searches books and inserts markdown links.

### `SettingsNav` (`components/settings-nav.tsx`)

Client component providing pill-style navigation across settings sub-pages. Uses `usePathname()` to highlight the active section. Rendered on all settings pages (Profile, Import, Export, Ghost Activity). The active pill uses `bg-accent text-white`; inactive pills use `bg-surface-2`. Fetches pending follow request count on mount and displays a red numeric badge on the "Follow requests" pill when count > 0.

### `PasswordForm` (`components/password-form.tsx`)

Client component rendered on the settings page below the profile form. Fetches `GET /api/me/account` on mount to determine whether the user has a password and/or Google linked, then shows the appropriate form: "Set password" for OAuth-only users, or "Change password" (with current password verification) for users who already have one. Calls `PUT /api/me/password`.

### `EmailForm` (`components/email-form.tsx`)

Client component rendered on the settings page below the password form. Fetches `GET /api/me/account` on mount to display the current email. Provides a form with "New email" and "Current password" fields. Calls `PUT /api/me/email` to change the email address. On success, the account's `email_verified` is set to false and the email verification banner will reappear.

### `DeleteDataForm` (`components/delete-data-form.tsx`)

Client component rendered on the settings page below the password form in a "Danger zone" section. Contains two destructive actions:

1. **Delete all my data** — red button that reveals a confirmation form requiring the user to type "delete my data". Calls `DELETE /api/me/account/data` to remove all user-owned records (books, reviews, tags, labels, follows, threads, notifications, etc.) while keeping the account. Redirects to home on success.
2. **Delete my account permanently** — second red button below, requiring the user to type "delete my account". Calls `DELETE /api/me/account` to remove all data **and** the user account itself. The proxy clears the auth cookie, and the user is redirected to the home page.

### `PendingImportsManager` (`components/pending-imports-manager.tsx`)

Client component for the `/settings/imports/pending` page. Displays unmatched import rows from previous Goodreads imports with title, author, ISBN, shelf, and rating info. Each row has three actions: "Search & Link" opens a modal with debounced book search — selecting a result calls `PATCH /api/me/imports/pending/:id` with `action: "resolve"` and the OL work ID. "Dismiss" marks the row as resolved without importing (`action: "dismiss"`). "Delete" permanently removes the row via `DELETE /api/me/imports/pending/:id`. Shows an empty state when all items are resolved.

### `RecommendButton` (`components/recommend-button.tsx`)

Client component rendered on book detail pages for logged-in users. Shows a "Recommend" button that opens a modal. The modal has two steps: (1) search for a user by username or display name (debounced, uses `GET /api/users?q=`), and (2) add an optional note before sending. Calls `POST /api/me/recommendations` with the selected username, book OL ID, and optional note. Shows a success confirmation after sending.

### `NotificationBell` (`components/notification-bell.tsx`)

Client component rendered in the nav bar for authenticated users. Polls `GET /api/me/notifications/unread-count` every 60 seconds and displays a bell icon with a red badge when unread notifications exist. Links to `/notifications`.

### `NotificationPreferences` (`components/notification-preferences.tsx`)

Client component rendered on the settings page. Displays toggle switches for each notification type (new publications, new threads, new links, new reviews, review likes, mentions, recommendations). Fetches current preferences on mount via `GET /api/me/notification-preferences` and updates individual toggles via `PUT /api/me/notification-preferences` with optimistic UI updates.

### `ReviewLikeButton` (`components/review-like-button.tsx`)

Client component for liking/unliking reviews. Shows a heart icon with like count. Toggling calls `POST /api/books/:workId/reviews/:userId/like`. Disabled for the user's own reviews. Used on the book detail page's community reviews section. Unauthenticated users see a static like count (no button).

### `GenreRatingEditor` (`components/genre-rating-editor.tsx`)

Client component for genre dimension ratings on book detail pages. Shows aggregate community ratings as horizontal bar charts (genre name, progress bar, average/10, rater count). Logged-in users can expand an editor with 0–10 sliders for each of the 12 predefined genres (Fiction, Non-fiction, Fantasy, Science fiction, Mystery, Romance, Horror, Thriller, Biography, History, Poetry, Children). Setting a slider to 0 removes the rating. Saves via `PUT /api/me/books/:olId/genre-ratings` and refreshes aggregate data on save.

### `BookScanner` (`components/book-scanner.tsx`)

Client component for the `/scan` page. Three input modes: Camera (uses browser `BarcodeDetector` API for real-time EAN-13 scanning on supported devices), Upload (sends image to `POST /api/books/scan` for server-side barcode detection via gozxing), and Enter ISBN (manual input via `GET /api/books/lookup`). Detected books are displayed with cover, metadata, and a StatusPicker for quick library addition. Supports scanning multiple books in a session with a history list.

### `EmptyState` (`components/empty-state.tsx`)

Reusable component for zero-data states across the app. Renders a centered message with an optional call-to-action link. Used on the feed page, notifications page, library pages (owner and visitor), and shelf/label views. Keeps empty state styling consistent: centered `py-16` container, `text-sm` message text, bordered button link.

### `KeyboardShortcuts` (`components/keyboard-shortcuts.tsx`)

Client component rendered in the root layout. Registers global keyboard shortcuts via the `useKeyboardShortcuts` hook: `/` focuses the search input, `Escape` closes any open modal or blurs the focused element, and `?` toggles a shortcuts help overlay. All shortcuts except `Escape` are suppressed when an input or textarea is focused. Shows a "Press ? for shortcuts" hint in the bottom-right corner for logged-in users.

### `KeyboardShortcutsOverlay` (`components/keyboard-shortcuts-overlay.tsx`)

Client component rendered by `KeyboardShortcuts` when the help overlay is open. Lists all available keyboard shortcuts in a modal with `kbd` badges. Closes on backdrop click or `Escape`.

### `StarRating` (`components/star-rating.tsx`)

Read-only star display used on label book cards.

### `SetOperationForm` (`components/set-operation-form.tsx`)

Client component on `/library/compare` "My Lists" tab. Two collection dropdown selectors, operation picker (union/intersection/difference) with descriptions, compare button, result book grid with covers and ratings, and "Save as new list" form. Calls `POST /api/me/shelves/set-operation` to compute and `POST /api/me/shelves/set-operation/save` to persist. ("Shelves" in the API path is a legacy name — the user-facing term is "labels".)

### `CrossUserCompareForm` (`components/cross-user-compare-form.tsx`)

Client component on `/library/compare` "Compare with a Friend" tab. Select one of your lists, enter a friend's username, load their public labels, pick one of their lists, choose an operation, compare. Result grid with book covers and star ratings. "Save as new list" option. Calls `GET /api/users/:username/shelves` to fetch friend's labels, `POST /api/me/shelves/cross-user-compare` to compute, and `POST /api/me/shelves/cross-user-compare/save` to persist.

### `CompareTabs` (`components/compare-tabs.tsx`)

Client component providing tab navigation between "My Lists" and "Compare with a Friend" modes on the compare page.

### `ReadingGoalCard` (`components/reading-goal-card.tsx`)

Client component displaying a reading goal progress bar on user profile pages. Shows "{X} of {Y} books read in {year}" with a visual progress bar.

### `ReadingGoalForm` (`components/reading-goal-form.tsx`)

Client component on the settings page for setting an annual reading goal. Calls `PUT /api/me/goals/:year` to create or update the goal.

### `Toast` (`components/toast.tsx`)

Global toast notification system. `ToastProvider` wraps the app in `layout.tsx` and renders fixed-position banners in the bottom-right corner. The `useToast()` hook exposes `toast.success(message)` and `toast.error(message)`. Toasts auto-dismiss after 4 seconds with a slide-in animation. Used across the app for feedback on user actions: book added/moved/removed, follow/unfollow, review saved, import complete, profile updated, export download, block/unblock, reading progress updated, and bulk library operations.

### `ReadingProgress` (`components/reading-progress.tsx`)

Client component on the book detail page (shown when status is "currently-reading"). Lets the user update progress by page number or percentage. Displays a progress bar and, when enough data is available (at least 1 day of reading, a known page count, and `progress_pages > 0`), shows an estimated reading pace ("~X pages/day") and estimated finish date below the progress bar. Uses `date_started` (or `date_added` as fallback) for the elapsed time calculation.

### `BlockButton` (`components/block-button.tsx`)

Client component on user profile pages. Shows "Block" button that opens an inline confirmation prompt. After blocking, the page reloads to show the restricted view. When already blocked, shows "Unblock" button instead. Calls `POST /api/users/:username/block` and `DELETE /api/users/:username/block`.

### `SavedSearches` (`components/saved-searches.tsx`)

Client component rendered on the search page for logged-in users. Shows saved searches as clickable chips above the search bar — clicking a chip populates the query and filters. Each chip has a small "x" button to delete the saved search. When filters or a query are active, shows a "Save this search" link that reveals a name input. Max 20 saved searches per user. Calls `GET /api/me/saved-searches`, `POST /api/me/saved-searches`, and `DELETE /api/me/saved-searches/:id`.

### `UserActivityList` (`components/user-activity-list.tsx`)

Client component that renders a list of activity items with cursor-based "Load more" pagination. Used on the user profile page sidebar. Receives initial activities and cursor from the server component, then fetches additional pages via `GET /api/users/:username/activity?cursor=`.

### `SeriesBookList` (`components/series-book-list.tsx`)

Client component that renders the book list on the series detail page. Each book shows position number, cover (with `BookCoverPlaceholder` fallback), title, author, and an interactive `StatusPicker` for logged-in users. Receives `books`, `statusValues`, `statusKeyId`, and `bookStatusMap` as props from the server component.

### `SeriesDescription` (`components/series-description.tsx`)

Client component for displaying and inline-editing a series name and description. Receives `seriesId`, `initialName`, `initialDescription`, and `isLoggedIn` as props. Renders the series `<h1>` title and description text. Logged-in users see an "Edit series" button that reveals an inline form with a name input and description textarea. Saves via `PATCH /api/series/:seriesId`.

---

## Adding a new page

1. Create `webapp/src/app/<path>/page.tsx`
2. If it needs auth: call `getUser()` and `getToken()` at the top of the server component
3. Fetch data server-side using `${process.env.API_URL}` with the token header if needed
4. Pass data to a client component for any interactive parts
5. Add any needed proxy routes under `webapp/src/app/api/`

## Adding a new API call from the client

1. Add a route handler in `webapp/src/app/api/` that extracts the `token` cookie and proxies to the Go API
2. Call the `/api/...` path from the client component using `fetch`
3. Match the Go API's HTTP method(s) — export named functions `GET`, `POST`, `PATCH`, `DELETE` etc. from the route handler

---

## Environment variables

| Variable | Where used | Purpose |
|---|---|---|
| `API_URL` | Server-side only | Go API base URL (e.g. `http://api:8090`) |
| `NEXT_PUBLIC_API_URL` | Client-side | Not currently used; reserved |
| `NEXT_PUBLIC_GOOGLE_CLIENT_ID` | Client-side (build-time) | Google OAuth client ID; when set, shows "Continue with Google" button on login/register pages |
| `GOOGLE_CLIENT_SECRET` | Server-side only | Google OAuth client secret; used by the callback route to exchange authorization codes for tokens |
| `NEXT_PUBLIC_URL` | Server-side | Public base URL of the webapp (e.g. `http://localhost:3000`); used to construct Google OAuth redirect URIs and post-auth redirects |
