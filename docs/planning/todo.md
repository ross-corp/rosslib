# Features

Backlog of small tasks for nephewbot to pick off. Each item should be self-contained and implementable without external coordination.

## import improvements

## Pending PRs

<!-- nephewbot moves tasks here when it opens a PR. Move to docs/planning/completed.md after merging. -->
- [Add Google Books fallback for Goodreads import](https://github.com/ross-corp/rosslib/pull/50) — Google Books API client as step 7 in import lookup chain; maps results back to OL by re-searching with Google's title/author
- [Add delete-all-data endpoint and settings UI](https://github.com/ross-corp/rosslib/pull/37) — DELETE /me/account/data removes all user data; settings page gets Danger Zone with typed confirmation
- [Add create/existing label options to import shelf mapping](https://github.com/ross-corp/rosslib/pull/38) — Goodreads import configure step gains "Create label" and "Add to existing label" actions alongside Tag and Skip
- [Add edition picker for book covers](https://github.com/ross-corp/rosslib/pull/42) — Edition selector modal on book detail page; selected edition cover shown on profile/shelf pages
- [Reorganize navbar into dropdown menus](https://github.com/ross-corp/rosslib/pull/44) — Replace flat nav links with Browse and Community dropdown menus
- [Add computed lists section to user profile page](https://github.com/ross-corp/rosslib/pull/45) — Computed lists section on profile with operation badges and Live indicator; migration adds computed list fields to collections schema
- [Add quick-add button for books on other users' pages](https://github.com/ross-corp/rosslib/pull/46) — QuickAddButton overlay on book covers in shelf grids and profile book rows for visitors; one-click "Want to Read" with dropdown for other statuses
- [Add bug report and feature request forms](https://github.com/ross-corp/rosslib/pull/47) — /feedback page with tabbed form; feedback PocketBase collection; admin Feedback section with status toggle
- [Add in-memory TTL cache for Open Library API](https://github.com/ross-corp/rosslib/pull/48) — 24h sync.Map cache on singleton OL client; hourly stats logging and expired entry eviction
- [Persist unmatched import rows for later resolution](https://github.com/ross-corp/rosslib/pull/49) — pending_imports collection stores unmatched Goodreads rows; GET/PATCH/DELETE endpoints; import form loads from server instead of localStorage
