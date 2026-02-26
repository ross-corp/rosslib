# Features

Larger features that need design work or external dependencies. nephewbot does NOT read from this file — safe to brainstorm here. Move items to `todo.md` (with full spec) when they're ready to be picked up.

## Genre predictions via embeddings

Full pipeline for predicting and displaying book genres using a local embeddings model.

1. Define a canonical genre list (~20 genres: Fiction, Non-fiction, Fantasy, Sci-fi, Mystery, Romance, Horror, Thriller, Biography, History, Poetry, Children's, Philosophy, Self-help, Science, True crime, Humor, Art, Religion, Travel)
2. Stand up a local embeddings endpoint (e.g. ollama with an embedding model)
3. For each book in a user's account, send `{title, author, description}` to the embeddings model and predict genre affinities
4. Present predictions to the user on the book detail page — show predicted genres as chips/tags with confidence scores
5. Let users adjust with sliders (0–5) for how much a book fits each genre
6. Store in DB: `predicted_genres` (model output), per-user genre ratings (user overrides), global genre scores (average across all users)
7. Use genre data to power recommendations and "similar books" features

Dependencies: local embeddings endpoint (ollama or similar), GPU access on the server. The existing `genre_ratings` table (0–10 manual ratings on 12 genres) could be extended or replaced.

## Integrations & external connections

### CLI tool
- `ross` CLI that talks to the API — search, add books, update status, view shelves from the terminal
- Auth via API token (new `api_tokens` table, generated from settings page)

### Calibre plugin
- Sync Calibre library with rosslib account — import books, sync read status
- Calibre has a plugin API in Python; plugin would call our REST API

### Kindle integration
- Import reading history from Kindle (likely screen scraping or Amazon API if available)
- Sync highlights and notes as review content

### Bookshelf batch scanning
- Scan a photo of an entire bookshelf (10-20 spines visible) and detect multiple ISBNs
- Extend the existing barcode scanner (`/scan` page) with a batch mode
- May need spine text OCR as a fallback when barcodes aren't visible
- Preprocessing: rotation correction, crop per-spine, try multiple barcode formats

## Community & discussion

### Spoiler-gated posts
- Posts on a book that are hidden until the reader has reached a certain page/percentage
- Ties into the existing reading progress system (`progress_pages`/`progress_percent` on `user_books`)
- UI: author sets a "visible after page X" threshold when creating a post; readers see a blurred card with "Read to page X to unlock"

### Book clubs
- Users create a club, invite members, pick a book, set a reading schedule
- Club page shows member progress, discussion threads scoped to the club
- Could reuse the existing threads system with a `club_id` scope

### Community wiki pages
- Objective, collaboratively edited pages about books (publication history, adaptations, awards, etc.)
- Separate from reviews (which are subjective). More like Wikipedia infoboxes.
- Needs edit history and moderation (reuse existing moderator system)

## User-submitted works

- Support for non-ISBN entries: fanfics, web serials, blog posts, self-published works
- Submission form with title, author, description, cover image, external URL
- Moderation queue (reuse existing admin panel pattern) — community votes or moderator approval before the work appears in search
- Submitted works get their own `books` entries with a `source = "user_submitted"` flag

## Short stories / anthology support

- First-class support for short stories as independent works that live inside anthologies
- Data model: `works` (a story) can belong to multiple `books` (anthologies/collections)
- Users can rate/review individual stories, not just the anthology
- Anthology page shows table of contents with per-story ratings

## AI features

- Embeddings-powered "similar books" recommendations (see genre predictions above)
- Reading taste fingerprint — vector representation of a user's reading preferences based on their ratings
- "If you liked X, try Y" powered by user-book embedding similarity rather than collaborative filtering
- Agent "users" that participate naturally (ghost users already exist — extend with embeddings-driven book recommendations in feed)

## Infrastructure

### CI/CD
- GitHub Actions: lint, typecheck, `go test`, and `npm run build` on every PR
- CD: auto-deploy to production on main merge (docker compose pull + restart on the server)
- Switch GHA runners to homelab machines to avoid burning minutes

### Testing
- Playwright or Cypress for webapp E2E tests — at minimum: login, search, add book, view profile
- Go integration tests against a test PocketBase instance

### Database
- Automated backups of the PocketBase SQLite DB (cron + rclone to cloud storage)
- Backup verification: periodic restore-and-check

### Nephewbot improvements
- nephewbot pulls from GitHub Issues instead of (or in addition to) `todo.md`
- In-app "Report bug" button creates a GitHub Issue via the API, nephewbot picks it up and submits a PR

## Business

- Premium tier (what would users pay for? export formats, advanced stats, unlimited shelves, early features?)
- Giveaways and publisher partnerships
- Branding: tealgardens.com is available
