# Features

Detailed specifications for all planned features.

---

## Basic auth

- [ ]  User Accounts
  - [ ] Registration & Login
  - [x] Register with username + email + password (bcrypt hashed, stored in `users` table).
  - [x] Login with email + password, returns a 30-day JWT set as an httpOnly cookie.
  - [ ] Email verification required before full access.
  - [ ] Password reset via email link.
  - [ ] `[post-MVP]` OAuth via Google.
- [ ] Profile
  - [ ] Public profile page at `/@username` showing:
    - [ ] Display name, bio, avatar.
    - [ ] Public collections (read, want to read, favorites, etc.).
    - [ ] Recent activity (reviews, threads, list updates).
    - [ ] Stats: books read, reviews written, followers/following count.
  - [ ] Profiles can be set to private; followers must be approved.
- [ ] Follow System
  - [ ] Asymmetric: you follow someone without them needing to follow back.
  - [ ] Private accounts require approval before a follow is accepted.
  - [ ] Following someone surfaces their activity in your feed.
  - [ ] "Friends" (mutual follows) can be surfaced in the UI as a distinct tier.

---

## Collections

- [ ] 3 default collections: want to read, reading, read

- [ ] custom collections
  - [ ] Non-exclusive by default (a book can appear in multiple custom collections).
  - [ ] Example: "Favorites", "Recommended to me", "Books set in Japan".
  - [ ] Collections can be made private or public.
  - [ ] Custom collections can also be marked exclusive and grouped if desired (e.g. a "Currently Reading" + "audiobook").
- [ ] Computed collections
  - [ ] Union: books in list A or list B.
  - [ ] Intersection: books in both list A and list B.
  - [ ] Difference: books in list A but not list B.
  - [ ] compute an operation + save as new collection
    - [ ] Example: "Books I've read that are also in my friend's Want to Read list."
- [ ] Sublists / Hierarchical Tags
  - [ ] A collection can have sub-labels that form a hierarchy:
    - [ ] Example: a "Science Fiction" collection with sub-labels "Space Opera", "Hard SF", "Cyberpunk".
    - [ ] Sub-labels are tags on `CollectionItem`, not separate collections.
    - [ ] Display as nested groupings on the collection page.

---

## Book Catalog

- [ ] Search
  - [ ] Full-text search by title, author, ISBN.
  - [ ] Faceted filters: genre, published year range, language.
  - [ ] Results ranked by relevance, with popular books surfaced higher.
- [ ] Book pages
  - [ ] Metadata: title, author(s), cover, description, publisher, year, page count.
  - [ ] Aggregate stats: average rating, read count, want-to-read count.
  - [ ] User's own status (added to which collection, their rating/review).
  - [ ] Community reviews and discussion threads.
  - [ ] Community links to related works.
- [ ] Author page
- [ ] Genre pages

- [ ] Edition handling

---

## Reviews & Ratings

- [ ] A user can rate a book 1–5 stars with half stars
- [ ] A rating alone (no review text) is valid.
- [ ] Review text is optional; can include a spoiler flag.
- [ ] One review per user per book; can be edited or deleted.
- [ ] Reviews are shown on book pages sorted by recency and follower relationships (reviews from people you follow shown first).

---

## Discussion Threads

- [ ] Any user can open a thread on a book's page.
- [ ] Thread has a title, body, and optional spoiler flag.
- [ ] Threaded comments support one level of nesting (reply to a comment, not reply to a reply).
- [ ] No upvotes at MVP; chronological sort only.
- [ ] Author can delete their own thread or comments; soft delete.

---

## Community Links (Wiki)

Users can submit directional links between books:

- [ ] Link types: `sequel`, `prequel`, `companion`, `mentioned_in`, `similar`, `adaptation`.
- [ ] Optional note explaining the connection.
- [ ] Links are upvotable; sorted by upvotes on book pages.
- [ ] Soft-deleted by moderators if spam or incorrect.
- [ ] Future: edit queue similar to book metadata edits.

---

## Import / Export

### Goodreads Import

- [ ] Accept a Goodreads CSV export file.
- [ ] Map Goodreads shelves to rosslib collections
- [ ] Attempt to match books by ISBN, falling back to title + author fuzzy match.
- [ ] Show a review screen before committing: matched / unmatched / ambiguous.
- [ ] Import star ratings and review text where present.

### CSV Export

- [ ] Export any collection (or all collections) to CSV.
- [ ] Columns: title, author, ISBN, date added, rating, review, collection name.
- [ ] Generated server-side and made available via a pre-signed S3 URL.

---

## Feed

- [ ] Chronological feed of activity from users you follow.
- [ ] Activity types surfaced: added to collection, wrote a review, started/finished a book, created a thread, submitted a link, followed a new user.
- [ ] No algorithmic ranking at MVP; pure chronological.
- [ ] Paginated (cursor-based).

---
