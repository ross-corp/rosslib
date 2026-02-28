# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Is

Rosslib is a Goodreads alternative. It has two services:

- `api/` — Go API built on PocketBase, runs on `:8090` (mapped to `:8091` on host)
- `webapp/` — Next.js 15 frontend on `:3000`

No external backing services needed — PocketBase bundles SQLite, auth, and file storage.

## Local Dev

```bash
task restart                # build & start both services via docker compose
docker compose logs -f      # watch logs
```

- API: `http://localhost:8091` (host) / `http://api:8090` (inter-container)
- PocketBase admin: `http://localhost:8091/_/`
- Webapp: `http://localhost:3000`

Data is persisted in the `pb_data` Docker volume.

## Commands

### API (Go / PocketBase)

```bash
cd api
go build .        # build binary
go test ./...     # run all tests
```

The API runs inside Docker via `docker compose`. The Dockerfile builds a static binary and runs `./pocketbase serve --http=0.0.0.0:8090`.

### Webapp (Next.js)

```bash
cd webapp
npm run dev       # dev server
npm run build     # production build
npm run lint      # ESLint
npx tsc --noEmit  # typecheck
```

## Architecture

### API Structure (`api/`)

PocketBase-based API. Routes are registered in `main.go` with three groups:
- **Public** — book search/lookup, user profiles, threads (GET)
- **Authenticated** — user operations (account, books, tags, labels, follows, feed, export); uses `apis.RequireAuth()`
- **Admin** — moderator-only (ghost management, link edit review)

Handler files in `api/handlers/`:
- `auth.go` — login, registration
- `books.go` — search, lookup, details, editions, stats, ratings
- `userbooks.go` — user book management (add, update, delete, status)
- `users.go` — profiles, search, reviews, labels, activity
- `collections.go` — labels/collections CRUD
- `tags.go` — tag key/value management
- `threads.go` — discussion threads and comments
- `links.go` — book links and voting
- `imports.go` — Goodreads CSV import
- `notifications.go` — notification handling
- `activity.go` — activity tracking
- `genreratings.go` — genre-based ratings
- `ghosts.go` — ghost user seeding/simulation
- `middleware.go` — auth middleware
- `helpers.go` — utility functions

Migrations in `api/migrations/` define the PocketBase collections (schema).

### Webapp Structure (`webapp/src/`)

- `app/` — Next.js App Router pages
  - `[username]/` — public profile; `[username]/shelves/[slug]` for label pages
  - `api/` — Next.js route handlers that proxy to the Go API (forwarding the `token` cookie)
  - `search/`, `login/`, `register/`, `settings/`, `users/`, `books/`
- `components/` — shared React components
- `lib/auth.ts` — server-side JWT decode from cookie (`getUser()`, `getToken()`)

The webapp proxies API calls through its own Next.js route handlers (`app/api/`), so the browser never talks directly to the API.

### Data Model (key collections)

- `users` — PocketBase auth collection; extended with `bio`, `is_private`, `display_name`, `avatar`, etc.
- `books` — global catalog keyed by `open_library_id`
- `collections` — named labels per user; `is_exclusive` + `exclusive_group` enforce mutual exclusivity
- `collection_items` — books in collections; holds `rating`, `review_text`, `spoiler`, `date_read`
- `user_books` — per-user book state (rating, review, progress, dates)
- `follows` — asymmetric social graph (`active` / `pending`)
- `tag_keys` / `tag_values` / `book_tag_values` — user-defined tagging system
- `threads` / `thread_comments` — book discussion threads
- `book_links` / `book_link_votes` / `book_link_edits` — community book relationships

### Environment Variables

The webapp reads `API_URL` (server-side, default `http://api:8090`) and `NEXT_PUBLIC_API_URL` (browser-side, default `http://localhost:8091`). These are set in `docker-compose.yml`.

## Docs

`docs/` has design docs worth consulting before implementing features:
- `datamodel.md` — full schema including planned tables not yet implemented
- `sysdesign.md` — deployment architecture and local dev setup
- `TODO.md` — feature backlog with implementation status
- `catalog_apis.md` — Open Library API reference
