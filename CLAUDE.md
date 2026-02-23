# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Is

Rosslib is a Goodreads alternative. It has two services:

- `api/` — Go REST API (Gin framework) on `:8080`
- `webapp/` — Next.js 15 frontend on `:3000`

Backing services: PostgreSQL, Redis, Meilisearch. Managed infrastructure on a single AWS EC2 instance via Terraform (`infra/`).

## Local Dev

```bash
cp .env.example .env
docker compose up          # starts postgres, redis, meilisearch
cd api && go run .         # Go API on :8080 (or use `air` for hot reload)
cd webapp && npm run dev   # Next.js on :3000
```

The `docker-compose.yml` also includes `api` and `webapp` services if you want to run everything in Docker. For active development, run the API and webapp directly.

## Commands

### API (Go)

```bash
cd api
go run .          # run the server
go test ./...     # run all tests
go build .        # build binary
```

Hot reload with `air` (config at `api/.air.toml`):
```bash
cd api && air
```

### Webapp (Next.js)

```bash
cd webapp
npm run dev       # dev server
npm run build     # production build
npm run lint      # ESLint
npx tsc --noEmit  # typecheck
```

## Architecture

### API Structure (`api/internal/`)

Each package follows a handler pattern — `NewHandler(pool)` returns a struct with methods registered as Gin route handlers. The `pool` is a `pgxpool.Pool` (pgx v5) passed through from `main`.

- `auth/` — registration, login, JWT issuance (30-day `httpOnly` cookie named `token`)
- `books/` — Open Library search/lookup proxy; upserts into local `books` table
- `collections/` — shelves CRUD; enforces mutual exclusivity within `exclusive_group`
- `imports/` — Goodreads CSV import (preview + commit); 5-worker goroutine pool for OL lookups
- `users/` — profile, follow/unfollow, user search
- `middleware/` — `Auth(secret)` (required) and `OptionalAuth(secret)` Gin middlewares
- `db/` — `db.go` (connection pool), `schema.go` (idempotent `CREATE TABLE IF NOT EXISTS` + `ALTER TABLE ADD COLUMN IF NOT EXISTS` run at startup via `db.Migrate`)
- `server/` — route registration in `NewRouter`

Routes are split into public and `authed` groups (the latter uses `middleware.Auth`).

### Webapp Structure (`webapp/src/`)

- `app/` — Next.js App Router pages
  - `[username]/` — public profile; `[username]/shelves/[slug]` for shelf pages
  - `api/` — Next.js route handlers that proxy to the Go API (forwarding the `token` cookie)
  - `search/`, `login/`, `register/`, `settings/`, `users/`, `books/`
- `components/` — shared React components (`nav.tsx`, `import-form.tsx`, `shelf-book-grid.tsx`, `shelf-picker.tsx`, etc.)
- `lib/auth.ts` — server-side JWT decode from cookie (`getUser()`, `getToken()`)

The webapp proxies API calls through its own Next.js route handlers (`app/api/`), so the browser never talks directly to `:8080`.

### Data Model (key tables)

- `users` — accounts; bcrypt passwords; soft-delete via `deleted_at`
- `books` — global catalog keyed by `open_library_id` (bare OL work ID, e.g. `OL82592W`); upserted on first add
- `collections` — named lists per user; `is_exclusive` + `exclusive_group` enforce mutual exclusivity (the three default shelves share `exclusive_group = 'read_status'`)
- `collection_items` — books in collections; holds `rating`, `review_text`, `spoiler`, `date_read`, `date_added`
- `follows` — asymmetric social graph; `status` is `'active'` or `'pending'`

Schema is applied idempotently at API startup via `db.Migrate`. No migration tool — new columns use `ALTER TABLE ... ADD COLUMN IF NOT EXISTS`.

### Environment Variables

Defined in `.env` (copy from `.env.example`):

| Variable | Purpose |
|---|---|
| `POSTGRES_USER/PASSWORD/DB` | PostgreSQL credentials |
| `MEILI_MASTER_KEY` | Meilisearch master key |
| `JWT_SECRET` | Signs JWTs |

The API reads `DATABASE_URL`, `REDIS_URL`, `PORT`, and `JWT_SECRET` from environment. The webapp reads `API_URL` (server-side) and `NEXT_PUBLIC_API_URL` (browser-side).

## CI/CD

`.github/workflows/deploy.yml` runs on push to `main`:
1. `test` — `go test ./...`, `tsc --noEmit`, `npm run lint` (also on PRs)
2. `build-push` — builds images, pushes to GHCR tagged with commit SHA + `latest`
3. `deploy` — copies `docker-compose.prod.yml` to EC2, pulls images, restarts

## Docs

`docs/` has design docs worth consulting before implementing features:
- `datamodel.md` — full schema including planned tables not yet implemented
- `sysdesign.md` — deployment architecture and local dev setup
- `TODO.md` — feature backlog with implementation status
- `catalog_apis.md` — Open Library API reference
