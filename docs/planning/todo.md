# Features

Backlog of small tasks for nephewbot to pick off. Each item should be self-contained and implementable without external coordination.

## stats & data

## moderation & safety

## data quality

- [ ] Populate series data from Open Library during book lookup. When a book is fetched from Open Library (in `books.go` `GetBookDetail`), check the OL work response for a `subject_places`, `subjects`, or — more usefully — the `/works/{workId}/editions.json` entries for `series` fields, or query `/search.json?q=series:{title}` to find related works. A simpler approach: parse the OL work's `links` and `subject_people` for series indicators. If a series is detected, auto-create the `series` and `book_series` records. This is best-effort — not all OL works have series data. Log when series data is found vs. not for visibility into coverage.

## import improvements

## Pending PRs

<!-- nephewbot moves tasks here when it opens a PR. Move to docs/planning/completed.md after merging. -->
- [Add delete-all-data endpoint and settings UI](https://github.com/ross-corp/rosslib/pull/37) — DELETE /me/account/data removes all user data; settings page gets Danger Zone with typed confirmation
- [Add create/existing label options to import shelf mapping](https://github.com/ross-corp/rosslib/pull/38) — Goodreads import configure step gains "Create label" and "Add to existing label" actions alongside Tag and Skip
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
