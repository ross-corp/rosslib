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
├── search/page.tsx                 book + user search
├── users/page.tsx                  browse all users
├── books/[workId]/page.tsx         single book page
├── settings/
│   ├── page.tsx                    profile settings
│   ├── import/page.tsx             Goodreads CSV import
│   └── tags/page.tsx               label category management
├── scan/page.tsx                   ISBN barcode scanner
├── library/compare/page.tsx        compare lists (set operations)
├── notifications/page.tsx          notification center
├── admin/page.tsx                 admin panel (moderator only)
├── [username]/
│   ├── page.tsx                    public profile
│   ├── shelves/[slug]/page.tsx     shelf page (owner gets library manager)
│   ├── tags/[...path]/page.tsx     tag browsing page
│   └── labels/[keySlug]/[...valuePath]/page.tsx   label browsing page (nested)
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
    ├── shelves/[shelfId]/books/route.ts
    ├── shelves/[shelfId]/books/[olId]/route.ts    ← GET, PATCH, DELETE
    ├── books/[workId]/links/route.ts              ← GET, POST community links
    ├── links/[linkId]/route.ts                    ← DELETE community link
    ├── links/[linkId]/vote/route.ts               ← POST, DELETE vote on link
    ├── links/[linkId]/edits/route.ts              ← POST propose link edit
    ├── admin/users/route.ts                       ← GET admin user list
    ├── admin/users/[userId]/moderator/route.ts    ← PUT grant/revoke moderator
    ├── admin/link-edits/route.ts                  ← GET list link edits
    ├── admin/link-edits/[editId]/route.ts         ← PUT approve/reject link edit
    ├── books/scan/route.ts                            ← POST barcode scan
    ├── books/lookup/route.ts                          ← GET ISBN lookup
    ├── books/[workId]/genre-ratings/route.ts         ← GET aggregate genre ratings
    ├── me/books/[olId]/genre-ratings/route.ts       ← GET, PUT user genre ratings
    ├── me/account/route.ts                         ← GET account info (has_password, has_google)
    ├── me/account/data/route.ts                   ← DELETE all user data
    ├── me/password/route.ts                        ← PUT set/change password
    ├── me/notifications/route.ts                  ← GET list notifications
    ├── me/notifications/unread-count/route.ts     ← GET unread count
    ├── me/notifications/read-all/route.ts         ← POST mark all read
    ├── me/notifications/[notifId]/read/route.ts   ← POST mark one read
    └── users/[username]/
        ├── tags/[...path]/route.ts
        ├── labels/[keySlug]/[...valuePath]/route.ts   ← catch-all for nested label paths
        └── shelves/[slug]/route.ts                ← GET (for client-side shelf switching)
```

---

## Key components

### `Nav` (`components/nav.tsx`)

Top navigation bar. Statically imported by each page — not in the root layout.

### `LibraryManager` (`components/library-manager.tsx`)

Full-page library manager rendered for shelf owners. Replaces the simple shelf grid on `[username]/shelves/[slug]` when `isOwner` is true.

Layout: `h-screen flex flex-col overflow-hidden` on the page, then inside LibraryManager:

```
┌─────────────────────────────────────────────────────┐
│ Nav                                                  │
├──────────┬──────────────────────────────────────────┤
│          │ top bar (shelf name / bulk action toolbar)│
│ sidebar  ├──────────────────────────────────────────┤
│          │                                          │
│ Shelves  │   book cover grid (scrollable)           │
│ Custom   │                                          │
│ Tags     │                                          │
│ Labels   │                                          │
│          │                                          │
└──────────┴──────────────────────────────────────────┘
```

**Sidebar** — clicking a shelf fetches its books client-side via `GET /api/users/:username/shelves/:slug`. Clicking a tag collection fetches via `GET /api/users/:username/tags/:path`. Clicking a label value fetches via `GET /api/users/:username/labels/:keySlug/*valuePath` (includes sub-values). Nested label values are indented by depth in the sidebar, showing only the last path segment as the display name.

**Top bar** — shows the current shelf name and book count when nothing is selected. Transforms into the bulk action toolbar when one or more books are checked:
- Rate — sets rating on all selected books via `PATCH /api/shelves/:shelfId/books/:olId`
- Move to shelf — moves via `POST /api/shelves/:targetId/books`, then refreshes the current shelf
- Labels — applies or clears a label value across all selected books via `PUT/DELETE /api/me/books/:olId/tags/:keyId`
- Remove — removes from current shelf via `DELETE /api/shelves/:shelfId/books/:olId`

Rate, Move, and Remove require a shelf context (disabled in tag-filtered views). Labels work in both shelf and tag views since they only need the `open_library_id`.

**Book grid** — cover images with a checkbox in the top-left. Checkboxes are invisible until hover or until at least one book is selected (at which point all checkboxes become visible). When books are selected, clicking a cover toggles selection instead of navigating to the book page.

### `ShelfBookGrid` (`components/shelf-book-grid.tsx`)

Simpler read-only-ish grid used on non-owner shelf views and the tag browsing page. Supports individual book removal (owner only) and the per-book `BookTagPicker`.

### `BookTagPicker` (`components/book-tag-picker.tsx`)

Dropdown for managing label assignments on a single book. Lazily loads current assignments on first open. Supports toggling predefined values and adding free-form values.

### `ShelfPicker` (`components/shelf-picker.tsx`)

Dropdown for adding/moving/removing a single book from shelves. Used on search results and book pages.

### `BookLinkList` (`components/book-link-list.tsx`)

Client component for community links (related books) on book detail pages. Shows links grouped by relationship type (sequel, prequel, companion, similar, etc.), sorted by upvote count. Logged-in users can upvote/unvote links, suggest new ones via an inline form, and propose edits to existing links (edit pencil icon). Proposed edits are submitted for moderator review. Target book is specified by Open Library work ID.

### `AdminUserList` (`components/admin-user-list.tsx`)

Client component for the `/admin` page. Provides a searchable, paginated table of all users with inline moderator toggle buttons. Moderators see a filled "Moderator" button; non-moderators see a "Grant" button. Search is debounced (300ms) and queries by username, display name, or email.

### `AdminLinkEdits` (`components/admin-link-edits.tsx`)

Client component for the `/admin` page. Displays proposed community link edits with status filter tabs (pending/approved/rejected). Each edit shows the proposer, book pair, current vs. proposed values (type and note) side by side, and approve/reject buttons for pending edits. Reviewed edits show the reviewer name, date, and optional comment.

### `ReviewText` (`components/review-text.tsx`)

Renders review text with wikilink and markdown link support. Parses two inline link syntaxes:
- `[[Book Title]]` — rendered as a link to `/search?q=Book%20Title`
- `[Book Title](/books/OL123W)` — rendered as a direct link to the book page

Used on book detail pages (community reviews), user reviews pages, recent reviews on profiles, and the collapsed review view in the book review editor. The companion `BookReviewEditor` component provides `[[` autocomplete that searches books and inserts markdown links.

### `PasswordForm` (`components/password-form.tsx`)

Client component rendered on the settings page below the profile form. Fetches `GET /api/me/account` on mount to determine whether the user has a password and/or Google linked, then shows the appropriate form: "Set password" for OAuth-only users, or "Change password" (with current password verification) for users who already have one. Calls `PUT /api/me/password`.

### `DeleteDataForm` (`components/delete-data-form.tsx`)

Client component rendered on the settings page below the password form in a "Danger zone" section. Shows a red "Delete all my data" button. Clicking it reveals a confirmation form where the user must type "delete my data" to proceed. Calls `DELETE /api/me/account/data` which removes all user-owned records (books, reviews, tags, shelves, follows, threads, notifications, etc.) but keeps the account. On success, redirects to the home page.

### `NotificationBell` (`components/notification-bell.tsx`)

Client component rendered in the nav bar for authenticated users. Polls `GET /api/me/notifications/unread-count` every 60 seconds and displays a bell icon with a red badge when unread notifications exist. Links to `/notifications`.

### `GenreRatingEditor` (`components/genre-rating-editor.tsx`)

Client component for genre dimension ratings on book detail pages. Shows aggregate community ratings as horizontal bar charts (genre name, progress bar, average/10, rater count). Logged-in users can expand an editor with 0–10 sliders for each of the 12 predefined genres (Fiction, Non-fiction, Fantasy, Science fiction, Mystery, Romance, Horror, Thriller, Biography, History, Poetry, Children). Setting a slider to 0 removes the rating. Saves via `PUT /api/me/books/:olId/genre-ratings` and refreshes aggregate data on save.

### `BookScanner` (`components/book-scanner.tsx`)

Client component for the `/scan` page. Three input modes: Camera (uses browser `BarcodeDetector` API for real-time EAN-13 scanning on supported devices), Upload (sends image to `POST /api/books/scan` for server-side barcode detection via gozxing), and Enter ISBN (manual input via `GET /api/books/lookup`). Detected books are displayed with cover, metadata, and a StatusPicker for quick library addition. Supports scanning multiple books in a session with a history list.

### `StarRating` (`components/star-rating.tsx`)

Read-only star display used on shelf book cards.

### `SetOperationForm` (`components/set-operation-form.tsx`)

Client component on `/library/compare` "My Lists" tab. Two collection dropdown selectors, operation picker (union/intersection/difference) with descriptions, compare button, result book grid with covers and ratings, and "Save as new list" form. Calls `POST /api/me/shelves/set-operation` to compute and `POST /api/me/shelves/set-operation/save` to persist.

### `CrossUserCompareForm` (`components/cross-user-compare-form.tsx`)

Client component on `/library/compare` "Compare with a Friend" tab. Select one of your lists, enter a friend's username, load their public shelves, pick one of their lists, choose an operation, compare. Result grid with book covers and star ratings. "Save as new list" option. Calls `GET /api/users/:username/shelves` to fetch friend's shelves, `POST /api/me/shelves/cross-user-compare` to compute, and `POST /api/me/shelves/cross-user-compare/save` to persist.

### `CompareTabs` (`components/compare-tabs.tsx`)

Client component providing tab navigation between "My Lists" and "Compare with a Friend" modes on the compare page.

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
