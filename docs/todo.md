# Todo

Organized by area. MVP items are unmarked; post-MVP items are labeled `[post-MVP]`.

---

## Infrastructure

- [x] Initialize Go API project (module, directory structure, linter config)
- [x] Initialize Next.js frontend project (TypeScript, Tailwind)
- [x] Set up PostgreSQL locally (Docker Compose for dev)
- [x] Set up Redis locally (Docker Compose)
- [x] Write Docker Compose file (webapp, api, postgres, redis, meilisearch)
- [x] Set up Meilisearch (Docker Compose service)
- [ ] Write initial DB migration tooling (golang-migrate or similar)
- [ ] Set up GitHub Actions CI pipeline (lint, test, build)
- [ ] Create AWS accounts / IAM roles for deployment
- [ ] Set up nginx on host with Let's Encrypt (Certbot)
- [ ] Provision EC2 instance (t3.medium or t3.large)
- [ ] Set up S3 bucket for file storage (covers, avatars, exports)
- [ ] Configure IAM role on EC2 for S3 access (no credentials in app)
- [ ] Configure Secrets Manager for credentials
- [ ] Set up Route 53 + ACM for domain

---

## Database

- [ ] Write migrations for all core tables (see `datamodel.md`)
  - [ ] users
  - [ ] books, authors, book_authors
  - [ ] genres, book_genres
  - [ ] collections, collection_items
  - [ ] reviews
  - [ ] follows
  - [ ] threads, comments
  - [ ] links
  - [ ] activity
  - [ ] book_stats
  - [ ] book_edits
- [ ] Seed script: import sample book data from Open Library

---

## Auth

- [ ] POST `/auth/register`
- [ ] POST `/auth/login`
- [ ] POST `/auth/refresh`
- [ ] POST `/auth/logout`
- [ ] POST `/auth/password-reset/request`
- [ ] POST `/auth/password-reset/confirm`
- [ ] JWT middleware (access token validation)
- [ ] `[post-MVP]` OAuth via Google

---

## Users

- [ ] GET `/users/:username` — public profile
- [ ] PATCH `/users/me` — update profile (display name, bio, avatar)
- [ ] POST `/users/me/avatar` — upload avatar to S3
- [ ] GET `/users/me/feed` — activity feed (cursor-paginated)
- [ ] POST `/users/:username/follow`
- [ ] DELETE `/users/:username/follow`
- [ ] GET `/users/:username/followers`
- [ ] GET `/users/:username/following`
- [ ] `[post-MVP]` Private account follow approval flow

---

## Books

- [ ] Integrate Open Library API client
- [ ] GET `/books/search?q=` — full-text search via Meilisearch
- [ ] GET `/books/:id` — book detail page data
- [ ] Background job: index books to Meilisearch on create/update
- [ ] Background job: refresh book_stats aggregates
- [ ] `[post-MVP]` POST `/books/:id/edits` — submit metadata correction
- [ ] `[post-MVP]` Admin moderation queue for book edits
- [ ] `[post-MVP]` Edition grouping (link multiple ISBNs as same work)

---

## Collections

- [ ] Seed default collections on user registration (Read, Currently Reading, Want to Read)
- [ ] GET `/users/:username/collections` — list user's public collections
- [ ] GET `/users/:username/collections/:slug` — collection detail + items
- [ ] POST `/users/me/collections` — create collection
- [ ] PATCH `/users/me/collections/:id` — rename, update description/visibility
- [ ] DELETE `/users/me/collections/:id`
- [ ] POST `/users/me/collections/:id/items` — add book (enforce exclusive_group logic)
- [ ] DELETE `/users/me/collections/:id/items/:bookId`
- [ ] PATCH `/users/me/collections/:id/items/:bookId` — update notes, sort_order
- [ ] GET `/users/me/collections/set-op?op=union|intersection|difference&a=:id&b=:id` — set operations
- [ ] `[post-MVP]` Save set operation result as new collection
- [ ] `[post-MVP]` Sub-labels / hierarchical tags on collection items

---

## Reviews & Ratings

- [ ] POST `/books/:id/reviews` — create/update review + rating
- [ ] DELETE `/books/:id/reviews` — delete own review
- [ ] GET `/books/:id/reviews` — list reviews (followers first, then recency)
- [ ] GET `/users/:username/reviews` — all reviews by user

---

## Threads & Comments

- [ ] GET `/books/:id/threads` — list threads for a book
- [ ] POST `/books/:id/threads` — create thread
- [ ] GET `/threads/:id` — thread detail with comments
- [ ] DELETE `/threads/:id` — soft delete own thread
- [ ] POST `/threads/:id/comments` — add comment or reply
- [ ] DELETE `/comments/:id` — soft delete own comment

---

## Community Links

- [ ] GET `/books/:id/links` — outbound links (sorted by upvotes)
- [ ] POST `/books/:id/links` — submit link to another book
- [ ] POST `/links/:id/upvote`
- [ ] DELETE `/links/:id/upvote`
- [ ] `[post-MVP]` Soft delete / moderator removal of links

---

## Import / Export

- [ ] POST `/import/goodreads` — upload CSV, parse, return preview
- [ ] POST `/import/goodreads/confirm` — commit import after user reviews
- [ ] GET `/export/collections` — generate CSV export, return pre-signed S3 URL

---

## Frontend

- [ ] Auth pages: login, register, forgot password
- [ ] User profile page (`/@username`)
- [ ] Collection page (`/@username/:collection-slug`)
- [ ] Book page (`/books/:id`)
- [ ] Search results page
- [ ] Feed page (home for logged-in users)
- [ ] Settings page (profile edit, account)
- [ ] Thread page (`/threads/:id`)
- [ ] Import flow (upload + review + confirm)
- [ ] Export flow
- [ ] `[post-MVP]` Set operation builder UI
- [ ] `[post-MVP]` Book edit submission UI

---

## Polish / Post-MVP

- [ ] Email notifications (new follower, reply to your thread)
- [ ] Reading goals (e.g. 50 books in a year) with progress tracking
- [ ] "Reading journal" per book (private notes, date started/finished)
- [ ] Lists shareable as public URLs with a static view
- [ ] Mobile app (React Native, reusing API)
- [ ] Algorithmic recommendations based on collection overlap with followed users
