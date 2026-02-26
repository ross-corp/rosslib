# Features

Backlog of all things we need #todo which are small enough to knock out at random with little additional thinking. note that this doc can get thrashed by AI so watch out

- [ ] Bookshelf scan — scan many books at once (10-20 on a shelf); batch mode for book scanner
- [ ] Improve barcode detection — try multiple barcode formats, rotation/crop preprocessing
- [ ] Kindle integration (TBD)

## genres

- [ ] generate a list of major genres (~20)
- [ ] connect to local embeddings endpoint
- [ ] for each book in a user's account, send the book (title, author, summary) to embeddings model and predict the genre
- [ ] present this prediction to the user somehow
- [ ] give them sliders on 0-5 of how much a given book fits into a given genre

- [ ] store all this in a DB
  - [ ] we get a "predicted genres", per-user genre ratings, and global (agerage of all users) genre