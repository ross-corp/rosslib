# Features

Detailed specifications for all planned features.

---

## User Accounts

### Registration & Login

- [ ] Register with email + password, or via OAuth (Google).
- [ ] Email verification required before full access.
- [ ] Login returns a JWT access token + httpOnly refresh token cookie.
- [ ] Password reset via email link.

### Profile

- [ ] Public profile page at `/@username` showing:
  - [ ] Display name, bio, avatar.
  - [ ] Public collections (read, want to read, favorites, etc.).
  - [ ] Recent activity (reviews, threads, list updates).
  - [ ] Stats: books read, reviews written, followers/following count.
- [ ] Profiles can be set to private; followers must be approved.

### Follow System

- [ ] Asymmetric: you follow someone without them needing to follow back.
- [ ] Private accounts require approval before a follow is accepted.
- [ ] Following someone surfaces their activity in your feed.
- [ ] "Friends" (mutual follows) can be surfaced in the UI as a distinct tier.

---

## Collections

Collections are the core organizational unit. Every user starts with three default exclusive collections:

- [ ] **Read** — books the user has finished.
- [ ] **Currently Reading** — in progress.
- [ ] **Want to Read** — the to-read pile.

These three share an `exclusive_group`, so a book can only be in one of them at a time. Moving a book from "Want to Read" to "Read" removes it from the former automatically.

### Custom Collections

Users can create additional collections with any name:

- [ ] Non-exclusive by default (a book can appear in multiple custom collections).
- [ ] Example: "Favorites", "Recommended to me", "Books set in Japan".
- [ ] Collections can be made private or public.
- [ ] Custom collections can also be marked exclusive and grouped if desired (e.g. a "Currently Listening" audiobook group).

### Set Operations

Users can derive new views from their collections using set operations:

- [ ] **Union**: books in list A or list B.
- [ ] **Intersection**: books in both list A and list B.
- [ ] **Difference**: books in list A but not list B.

Example: "Books I've read that are also in my friend's Want to Read list."

Set operation results are not persisted — they are computed on demand. Users can optionally save a result as a new collection.

### Sublists / Hierarchical Tags

A collection can have sub-labels that form a hierarchy:

- [ ] Example: a "Science Fiction" collection with sub-labels "Space Opera", "Hard SF", "Cyberpunk".
- [ ] Sub-labels are tags on `CollectionItem`, not separate collections.
- [ ] Display as nested groupings on the collection page.

---

## Book Catalog

### Search

- [ ] Full-text search by title, author, ISBN.
- [ ] Faceted filters: genre, published year range, language.
- [ ] Results ranked by relevance, with popular books surfaced higher.

### Book Pages

Each book has a public page showing:

- [ ] Metadata: title, author(s), cover, description, publisher, year, page count.
- [ ] Aggregate stats: average rating, read count, want-to-read count.
- [ ] User's own status (added to which collection, their rating/review).
- [ ] Community reviews and discussion threads.
- [ ] Community links to related works.

### Editions

Multiple editions of the same work are modeled as separate `Book` records linked by an `editions` relationship (future work). MVP treats each ISBN as a distinct book.

---

## Reviews & Ratings

- [ ] A user can rate a book 1–5 stars (half-star not supported at MVP).
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
- [ ] Map Goodreads shelves to rosslib collections:
  - [ ] "read" → Read
  - [ ] "to-read" → Want to Read
  - [ ] "currently-reading" → Currently Reading
  - [ ] Custom shelves → custom collections (created if they don't exist).
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

## Differentiation from Goodreads

The features that make rosslib meaningfully better:

| Area | Goodreads | Rosslib |
|---|---|---|
| Collections | Fixed shelves, limited custom | Flexible collections with set operations and sublists |
| Social graph | Mutual friends, clunky | Asymmetric follows, clean feed |
| Discussion | Cluttered, dated UI | Clean threads, spoiler flags |
| Book connections | None | Community-submitted typed links |
| Data portability | CSV export (limited) | Full CSV export + Goodreads import |
| Metadata edits | No | Wiki-style edit queue |
