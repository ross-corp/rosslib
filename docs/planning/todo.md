# Features

Backlog of small tasks for nephewbot to pick off. Each item should be self-contained and implementable without external coordination.

## UI improvements

- [ ] Fix book interaction buttons when viewing books on another user's page. When browsing another user's shelf or profile page and seeing their books, there's no quick way to add a book to your own library. Add a small "Want to read" button on book cover cards in shelf grids and profile book lists (only shown when viewing another user's page, not your own). Clicking it adds the book to the viewer's "Want to Read" status via `POST /me/books`. Include a small dropdown on the button with options: "Want to read", "Currently reading", "Add to label...", and "Rate & review" (navigates to the book detail page).

## bug reports

- [ ] Add bug report and feature request forms. Create a `/feedback` page with two tabs: "Bug Report" and "Feature Request". Bug report fields: title, description (textarea), steps to reproduce (textarea), severity dropdown (low/medium/high). Feature request fields: title, description (textarea). API endpoint `POST /feedback` stores submissions in a new `feedback` PocketBase collection (fields: `user` relation, `type` enum bug/feature, `title`, `description`, `steps_to_reproduce`, `severity`, `status` enum open/closed, `created`). Admin page (`/admin`) gets a new "Feedback" tab showing submissions with status toggle. Add a "Feedback" link in the nav or footer.

## caching

- [ ] Add a server-side response cache for Open Library API calls. Create a simple in-memory TTL cache (sync.Map or similar) keyed by the full OL request URL. Cache successful responses for 24 hours. Apply it to the shared OL HTTP client (the rate-limited client in `api/handlers/helpers.go` or the `olhttp` package). This avoids redundant lookups when multiple users search for or view the same book. Log cache hit/miss rates at startup interval (e.g. every hour). No frontend changes needed.

## import improvements

- [ ] Add a fallback catalog API (Google Books or similar) for Goodreads import. The current import uses Open Library exclusively. After the recent title/author cleaning improvements, the remaining misses are entries genuinely absent from OL: periodical/magazine issues (e.g. *Destinies* magazine), niche regional textbooks, and academic journals. Adding Google Books as a fallback after all OL attempts fail would cover many of these. The challenge is mapping Google Books results back to OL work IDs (since rosslib uses OL IDs as canonical identifiers) — either search OL by the title/author Google returns, or support books without OL IDs as local-only entries. Requires: a `googleBooksLookup(isbn, title, author)` helper in `api/handlers/helpers.go`, an optional `GOOGLE_BOOKS_API_KEY` env var (free tier: 1,000 req/day), and integration into the import preview lookup chain as a new step between the OL ISBN search and the title+author fallback.

- [ ] Persist unmatched import rows so users can retry or manually resolve them later. Currently, books that fail to match during Goodreads import are silently dropped. Instead, store unmatched rows in a `pending_imports` PocketBase collection (fields: `user` relation, `source` "goodreads", `title`, `author`, `isbn13`, `exclusive_shelf`, `custom_shelves` json, `rating`, `review_text`, `date_read`, `date_added`, `status` "unmatched"/"resolved", `created`). Show a "Failed imports" section on the import results page and a persistent banner/table in the library UI listing unmatched books with options to: search & manually match, retry lookup, or dismiss. API endpoints: `GET /me/imports/pending`, `PATCH /me/imports/pending/:id` (to resolve with a chosen OL ID or dismiss), `DELETE /me/imports/pending/:id`.

## Pending PRs

<!-- nephewbot moves tasks here when it opens a PR. Move to docs/planning/completed.md after merging. -->
- [Add delete-all-data endpoint and settings UI](https://github.com/ross-corp/rosslib/pull/37) — DELETE /me/account/data removes all user data; settings page gets Danger Zone with typed confirmation
- [Add create/existing label options to import shelf mapping](https://github.com/ross-corp/rosslib/pull/38) — Goodreads import configure step gains "Create label" and "Add to existing label" actions alongside Tag and Skip
- [Add edition picker for book covers](https://github.com/ross-corp/rosslib/pull/42) — Edition selector modal on book detail page; selected edition cover shown on profile/shelf pages
- [Reorganize navbar into dropdown menus](https://github.com/ross-corp/rosslib/pull/44) — Replace flat nav links with Browse and Community dropdown menus
- [Add computed lists section to user profile page](https://github.com/ross-corp/rosslib/pull/45) — Computed lists section on profile with operation badges and Live indicator; migration adds computed list fields to collections schema
