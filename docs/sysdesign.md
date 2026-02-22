# System Design

## Overview

Rosslib is a social reading site — a Goodreads/Letterboxd equivalent for books. Users track what they've read, organize books into flexible collections, discuss works, and share links between related content in a communal wiki style.

## Tech Stack

| Layer | Choice | Rationale |
|---|---|---|
| Webapp | Next.js (TypeScript) | SSR for SEO on book/profile pages; runs as a persistent Node.js process |
| Backend API | Go (Gin) | Fast, low memory footprint, single binary deploys cleanly on a VPS |
| Database | PostgreSQL | Relational model fits social graph + book catalog; self-hosted, no managed service needed |
| Cache | Redis | Session storage, feed caching, rate limiting; lightweight, easy to self-host |
| Search | Meilisearch | Full-text book/author search with faceting; single binary, ~50MB RAM at rest, VPS-friendly |
| File Storage | S3 | Cover images, avatars, CSV exports; kept as a managed service since object storage is hard to self-host reliably and S3 is cheap |
| Auth | JWT (access + refresh tokens) | Stateless; works well across web and future mobile clients |
| Reverse Proxy | nginx | TLS termination, routes traffic between webapp and API, serves static assets |
| Deployment | Docker Compose (VPS) | All services run as containers on one or two EC2/VPS instances |
| CI/CD | GitHub Actions | Build, test, push images, deploy via SSH on merge to main |

## Architecture

```
                    Internet
                        │ HTTPS :443
                ┌───────▼────────┐
                │     nginx      │  TLS termination (Let's Encrypt)
                └──┬─────────┬───┘
                   │         │
       /api/*      │         │  /*
        ┌──────────▼──┐   ┌──▼───────────────┐
        │  Go API     │   │  Next.js webapp  │
        │  :8080      │   │  :3000           │
        └──────┬──────┘   └────────┬─────────┘
               │                  │ API calls (server-side)
               └──────────┬───────┘
                          │
          ┌────────────────┼──────────────────┐
          │                │                  │
    ┌─────▼──────┐   ┌─────▼──────┐   ┌───────▼───────┐
    │ PostgreSQL │   │   Redis    │   │  Meilisearch  │
    │   :5432    │   │   :6379    │   │    :7700      │
    └────────────┘   └────────────┘   └───────────────┘

                        S3 (AWS)
               covers, avatars, CSV exports
               (accessed directly by Go API)
```

All services run in Docker Compose on the same host at MVP. PostgreSQL can be moved to a dedicated instance when needed.

## Core Components

### Webapp (Next.js)

Runs as a long-lived Node.js process. Responsibilities:

- Server-side renders book pages, profile pages, and collection pages for SEO.
- Handles all user-facing routing.
- Fetches data from the Go API on the server side (same host, no network latency).
- Static assets (JS, CSS) served by nginx or CDN.

Next.js is not a static export — it runs as a server because SSR is needed for dynamic per-user pages and good SEO on public book/profile pages.

### API Server (Go)

REST JSON API. Organized into modules:

- `auth` — registration, login, token refresh, password reset
- `users` — profiles, follow/friend graph
- `books` — catalog CRUD, search proxy to Meilisearch
- `collections` — user lists, list items, set operations
- `reviews` — ratings, text reviews, per-book aggregate stats
- `threads` — discussion threads on book pages
- `links` — community-submitted links between works (wiki)
- `import` — Goodreads CSV ingestion
- `export` — CSV generation, S3 upload, pre-signed URL response

The Go binary is compiled and run in a minimal Docker container. Single process, no orchestration overhead.

### Database (PostgreSQL)

See `datamodel.md` for full schema. Key design choices:

- Books and authors are shared/global (not per-user). Users reference them.
- Collections use an `exclusive_group` field to enforce mutual exclusivity (read/unread) at the application layer.
- Social graph is a `follows` join table (asymmetric follow).
- Soft deletes on most user-generated content (`deleted_at` timestamp).

At MVP, PostgreSQL runs in Docker Compose on the same host as the API. Persist data with a named Docker volume. Back up with `pg_dump` on a cron job to S3.

### Search (Meilisearch)

Single binary, minimal configuration, excellent out-of-the-box relevance. Books and authors are indexed on write.

Index fields:

- `title`, `author_names` (full-text)
- `isbn`, `isbn13` (filterable)
- `genres`, `tags` (filterable, faceted)
- `published_year` (filterable, sortable)

Sync strategy: write to Postgres first, then enqueue a background task to update Meilisearch. Accept eventual consistency for search. Meilisearch can be fully re-indexed from Postgres if needed.

### Cache (Redis)

Used for:

- Refresh token storage (with user_id → token mapping for revocation).
- Feed assembly cache per user (TTL 2 min).
- Follower list cache per user (TTL 5 min).
- Rate limiting counters on auth and write endpoints.

### File Storage (S3)

Kept as a managed service. Object storage is difficult to self-host reliably and S3 is inexpensive at this scale.

- Book cover images: `covers/{isbn}.jpg` (ingested from Open Library).
- User avatars: `avatars/{user_id}/{uuid}.jpg`.
- CSV exports: `exports/{user_id}/{timestamp}.csv`, returned as pre-signed URLs (TTL 1 hour).

Access controlled via IAM role on the EC2 instance — no credentials in the application.

## Deployment

### VPS / EC2 Setup

One instance to start (e.g. `t3.medium` or `t3.large`). All services run via Docker Compose:

```yaml
services:
  webapp:    # Next.js :3000
  api:       # Go :8080
  postgres:  # :5432, named volume
  redis:     # :6379
  meilisearch: # :7700, named volume
```

nginx runs on the host (not in Docker) to handle TLS via Let's Encrypt / Certbot cleanly.

### Scaling Path

When the single instance becomes a bottleneck:

1. Move PostgreSQL to its own instance (RDS standard or self-managed on a second VPS).
2. Move Redis to its own instance (ElastiCache basic tier or self-managed).
3. Add a second app server and put a load balancer (ALB) in front.
4. Move Meilisearch to its own instance.

No architectural changes required — just service relocation.

## External Data

Book metadata seeded from the Open Library API. Strategy:

- On first lookup of an ISBN not in the DB, fetch from Open Library and persist locally.
- Background goroutine runs periodic refresh of metadata for high-activity books.
- User corrections go through a `book_edits` review queue (wiki-style, future moderation).

## Auth Flow

1. `POST /auth/register` — creates user, returns access token (15 min) + refresh token (30 days, stored in httpOnly cookie).
2. `POST /auth/refresh` — validates refresh token in Redis, issues new access token.
3. Access token sent as `Authorization: Bearer <token>` on authenticated API requests.
4. Refresh tokens stored in Redis keyed by `refresh:{token}` → `user_id`, enabling revocation.

## Key Design Decisions

**Self-hosted over managed services** — Aurora Serverless, ElastiCache, and OpenSearch Service all carry significant baseline costs even at low traffic. PostgreSQL, Redis, and Meilisearch running in Docker Compose on a single VPS handle this workload at a fraction of the cost.

**Meilisearch over OpenSearch** — OpenSearch requires at minimum 2 nodes and gigabytes of heap. Meilisearch is a single binary using ~50MB at rest, has excellent relevance out of the box, and is trivially self-hosted. More than adequate for book/author search.

**S3 kept as managed** — Object storage is the exception: self-hosting (MinIO etc.) adds operational risk for a use case that S3 handles cheaply and reliably.

**Webapp as a server process** — Next.js runs SSR for SEO on book and profile pages. A static export would not work for authenticated, personalized, or dynamically generated pages.

**Relational DB** — Social graph, collection membership, and set operations are naturally relational. A document DB would complicate these significantly.

**Asymmetric follows** — Lower friction than mutual friends. A "mutual" concept can be surfaced in the UI later without schema changes.

**Set operations in application layer** — Intersection/union/difference of collections computed at read time. Simple to implement; revisit with materialized views if needed.

**No microservices at MVP** — Single Go binary and single Next.js process. Internal package structure allows splitting later.

## AWS Services Used

| Service | Use | Why AWS vs Self-Hosted |
|---|---|---|
| EC2 | Host all services | VPS-style, full control |
| S3 | File storage | Object storage is hard to self-host reliably |
| Route 53 | DNS | Convenient with EC2 |
| ACM / Let's Encrypt | TLS | Let's Encrypt via Certbot on nginx |
| GitHub Actions | CI/CD | Build + deploy via SSH |

Everything else (Postgres, Redis, Meilisearch) self-hosted on EC2 via Docker Compose.
