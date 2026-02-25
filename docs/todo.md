# Features

Backlog of all things we need #todo

Once we're further along we'll move to GH projects. this is fine for now

## UNSORTED

- [ ] add a DNF option with a date, so a book can be completed OR stopped on a date.

## Book Scanning

- [ ] Book cover scan — take a picture of a cover, lookup ISBN for import
- [ ] Bookshelf scan — same idea but for many books at once (10-20 on a shelf)

## User Accounts

- [ ] OAuth via Google
- [ ] Email verification before full access (?)
- [ ] Password reset via email link

## Social

- [ ] Follow authors, see new publications
- [ ] Follow works, see sequels / new discussions / links
- [ ] Represent users who are also authors (badges)

## Search & Book Pages

- [ ] Edition handling
- [ ] Community links to related works on book pages

## Computed Lists

- [ ] Set operations on lists (union, intersection, difference)
  - [ ] Example: "Books I've read that are also in my friend's Want to Read list"
- [ ] Save result as a new list
- [ ] Continuous vs. one-time computed lists

## Sublists / Hierarchical Tags

- [ ] Sub-labels within a list that form a hierarchy (e.g. "Science Fiction" > "Space Opera", "Hard SF", "Cyberpunk")
- [ ] Sub-labels are tags on list items, not separate lists
- [ ] Display as nested groupings on the list page

## Discussion Threads

- [ ] Recommend merging similar threads
  - [ ] Link to similar threads if similarity score > threshold

## Community Links (Wiki)

- [ ] Link types: `sequel`, `prequel`, `companion`, `mentioned_in`, `similar`, `adaptation`
- [ ] Optional note explaining the connection
- [ ] Links are upvotable; sorted by upvotes on book pages
- [ ] Soft-deleted by moderators if spam/incorrect
- [ ] Future: edit queue similar to book metadata edits

## Genre Ratings

- [ ] Rate a book on genre dimensions (e.g. "is this book a comedy? 0-10")
- [ ] Data model changes needed; useful for recommendations

## Reviews

- [ ] Wikilinks to other books in review text

## API

- [ ] Swagger docs / route documentation so users can write CLIs or tools
  - [ ] Rate-limit upstream-proxied routes to avoid getting banned from book DBs

## Import / Export

- [ ] Kindle integration (TBD)

## Feed

- [ ] Activity type: submitted a link (requires community links feature)
