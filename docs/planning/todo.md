# Features

Backlog of small tasks for nephewbot to pick off. Each item should be self-contained and implementable without external coordination.

## imports

- [ ] Add a "Create label" option to the Goodreads import shelf mapping. Currently the import preview (`import-form.tsx`) only offers "Tag" and "Skip" for each Goodreads shelf. Add a third option: "Create label" — when selected, show an inline form where the user types a label name. Multiple shelves can be assigned to the same label (each shelf becomes a value under that label key). Also add an "Add to existing label" option that shows a dropdown of the user's existing label keys. On import commit, create the label key (via tag_keys) and assign the appropriate tag_values, then tag each imported book with its shelf's label value.

## edition handling

- [ ] Add a "Change edition" button on the book detail page (visible only when the user has the book in their library). Clicking it opens a modal/panel showing the available editions (reuse the existing `EditionList` data from `GET /books/:workId/editions`). Each edition shows its cover thumbnail, format, publisher, and ISBN. Selecting an edition updates the user's `user_books` record with a new `selected_edition_key` field (Open Library edition key like `/books/OL123M`). The profile and shelf pages should then display the selected edition's cover instead of the default work cover. Requires: adding `selected_edition_key` column to `user_books` (PocketBase migration), a `PATCH /me/books/:olId` field for it, and frontend logic to prefer the edition cover URL when rendering book covers.

## UI improvements

- [ ] Reorganize the navbar into dropdown menus. Replace the flat link list with two dropdown menus: **Browse** (Search books, Genres, Scan ISBN) and **Community** (Browse users, My feed). Keep the notification bell and user avatar/settings as standalone items. Use a simple CSS dropdown or headless UI popover — no external library needed. The dropdowns should work on both desktop (hover or click) and mobile (click).

- [ ] Add a "Computed lists" section to the user profile page. Currently computed/continuous lists only appear in the shelves list and are not prominent. Add a dedicated card or section on the profile page (below Favorites, above Tags) that shows the user's computed lists with their operation type badge (union/intersection/difference) and a "Live" indicator for continuous lists. Link each to its shelf detail page.

- [ ] Fix book interaction buttons when viewing books on another user's page. When browsing another user's shelf or profile page and seeing their books, there's no quick way to add a book to your own library. Add a small "Want to read" button on book cover cards in shelf grids and profile book lists (only shown when viewing another user's page, not your own). Clicking it adds the book to the viewer's "Want to Read" status via `POST /me/books`. Include a small dropdown on the button with options: "Want to read", "Currently reading", "Add to label...", and "Rate & review" (navigates to the book detail page).

## bug reports

- [ ] Add bug report and feature request forms. Create a `/feedback` page with two tabs: "Bug Report" and "Feature Request". Bug report fields: title, description (textarea), steps to reproduce (textarea), severity dropdown (low/medium/high). Feature request fields: title, description (textarea). API endpoint `POST /feedback` stores submissions in a new `feedback` PocketBase collection (fields: `user` relation, `type` enum bug/feature, `title`, `description`, `steps_to_reproduce`, `severity`, `status` enum open/closed, `created`). Admin page (`/admin`) gets a new "Feedback" tab showing submissions with status toggle. Add a "Feedback" link in the nav or footer.

## caching

- [ ] Add a server-side response cache for Open Library API calls. Create a simple in-memory TTL cache (sync.Map or similar) keyed by the full OL request URL. Cache successful responses for 24 hours. Apply it to the shared OL HTTP client (the rate-limited client in `api/handlers/helpers.go` or the `olhttp` package). This avoids redundant lookups when multiple users search for or view the same book. Log cache hit/miss rates at startup interval (e.g. every hour). No frontend changes needed.

## Pending PRs

<!-- nephewbot moves tasks here when it opens a PR. Move to docs/planning/completed.md after merging. -->
- [Add delete-all-data endpoint and settings UI](https://github.com/ross-corp/rosslib/pull/37) — DELETE /me/account/data removes all user data; settings page gets Danger Zone with typed confirmation
