# Features

Backlog of all things we need #todo

Once we're further along we'll move to GH projects. this is fine for now

## Webapp

- [ ] User Accounts
  - [x] Registration & Login
  - [x] Register with username + email + password (bcrypt hashed, stored in `users` table).
  - [x] Login with email + password, returns a 30-day JWT set as an httpOnly cookie.
  - [ ] Email verification required before full access.
  - [ ] Password reset via email link.
  - [ ] OAuth via Google.
- [ ] Profile
  - [x] Public profile page at `/{username}` — display name, byline, member since.
  - [x] Edit profile page at `/settings` — set display name and byline.
  - [x] Default shelves (Want to Read, Currently Reading, Read) shown on profile with item counts; cards link to shelf pages.
  - [x] Shelf pages at `/{username}/shelves/{slug}` — cover grid with title, owner can remove books inline.
  - [ ] Avatar upload.
  - [ ] Recent activity (reviews, threads, list updates) on profile.
  - [ ] Stats: books read (done), reviews written, followers/following count (done).
  - [ ] Profiles can be set to private; followers must be approved.
- [ ] Objects
  - [ ] Work pages at /w/dune
  - [ ] Author pages at /a/frank_herbert

- [ ] Search & Discovery
  - [x] Search bar in nav — submits GET to /search.
  - [x] `/search` page — Books tab searches by title via Open Library API (up to 20 results with cover, authors, year). People tab searches users by username or display name.
  - [x] `/users` page — browse all users, alphabetical, paginated (20/page).
  - [x] Tab selector on `/search` to filter between Books and People.
  - [x] "Add to shelf" picker on each book search result — logged-in users can add/move/remove books across their 3 default shelves inline.
  - [ ] Author tab in search.
  - [ ] Full-text book/author search via Meilisearch (will replace Open Library as primary search backend).

- [ ] Social
  - [x] Follow / unfollow users (asymmetric). Follow button on profile page; `is_following` returned from profile endpoint.
  - [x] `follows` table with `(follower_id, followee_id)` PK and `status` field.
  - [ ] Private accounts require follow approval (status = 'pending').
  - [x] Followers / following counts on profile.
  - [ ] "Friends" (mutual follows) surfaced in UI.
  - [ ] Follow authors, see new publications.
  - [ ] Follow works, see sequels / new discussions / links.

## Data Model

- [ ] Collections
  - [x] 3 default collections for all users: Want to Read, Currently Reading, Read — created on registration (or lazily on first `/me/shelves` call for existing users).
  - [x] `books` table — global catalog keyed by `open_library_id`; upserted when a user adds a book to a shelf.
  - [x] `collections` + `collection_items` tables with `is_exclusive` / `exclusive_group` for mutual exclusivity enforcement.
  - [x] API: `GET /users/:username/shelves`, `GET /users/:username/shelves/:slug`, `GET /me/shelves`, `POST /shelves/:shelfId/books`, `DELETE /shelves/:shelfId/books/:olId`.
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
    - [ ] enable continuous v. one-time computed collections

- [ ] Sublists / Hierarchical Tags
  - [ ] A collection can have sub-labels that form a hierarchy:
    - [ ] Example: a "Science Fiction" collection with sub-labels "Space Opera", "Hard SF", "Cyberpunk".
    - [ ] Sub-labels are tags on `CollectionItem`, not separate collections.
    - [ ] Display as nested groupings on the collection page.

## Connection to Book DBs

- [ ] Search
  - [x] Book title search via Open Library API (`GET /books/search?q=<title>`). Returns title, authors, cover image, first publish year, ISBNs.
  - [ ] Author search.
  - [ ] ISBN-direct lookup (for Goodreads import matching).
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

## Reviews & Ratings

- [ ] A user can rate a book 1–5 stars with half stars
- [ ] A rating alone (no review text) is valid.
- [ ] Review text is optional; can include a spoiler flag.
- [ ] One review per user per book; can be edited or deleted.
- [ ] Reviews are shown on book pages sorted by recency and follower relationships (reviews from people you follow shown first).

## Discussion Threads

- [ ] Any user can open a thread on a book's page.
- [ ] Thread has a title, body, and optional spoiler flag.
- [ ] Threads get reccommended for union if they're similar enough
  - [ ] link to similar comments if similarity score > some percentage
- [ ] Threaded comments support one level of nesting (reply to a comment, not reply to a reply).
- [ ] No upvotes at MVP; chronological sort only.
- [ ] Author can delete their own thread or comments; soft delete.

## Community Links (Wiki)

- [ ] Link types: `sequel`, `prequel`, `companion`, `mentioned_in`, `similar`, `adaptation`.
- [ ] Optional note explaining the connection.
- [ ] Links are upvotable; sorted by upvotes on book pages.
- [ ] Soft-deleted by moderators if spam or incorrect.
- [ ] Future: edit queue similar to book metadata edits.

## Import / Export

- [ ] Goodreads Import
  - [ ] Accept a Goodreads CSV export file.
  - [ ] Map Goodreads shelves to rosslib collections
  - [ ] Attempt to match books by ISBN, falling back to title + author fuzzy match.
  - [ ] Show a review screen before committing: matched / unmatched / ambiguous.
  - [ ] Import star ratings and review text where present.
- [ ]  CSV Export
  - [ ] Export any collection (or all collections) to CSV.
  - [ ] Columns: title, author, ISBN, date added, rating, review, collection name.
  - [ ] Generated server-side and made available via a pre-signed S3 URL.

## Feed

- [ ] Chronological feed of activity from users you follow.
- [ ] Activity types surfaced: added to collection, wrote a review, started/finished a book, created a thread, submitted a link, followed a new user.
- [ ] No algorithmic ranking at MVP; pure chronological.
- [ ] Paginated (cursor-based).
