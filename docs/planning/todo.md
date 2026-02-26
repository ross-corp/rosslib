# Features

Backlog of all things we need #todo which are small enough to knock out at random with little additional thinking. note that this doc can get thrashed by AI so watch out


## genres

- [ ] I need a tool to delete all content from my account. just for testing for now

- [ ] generate a list of major genres (~20)
- [ ] connect to local embeddings endpoint
- [ ] for each book in a user's account, send the book (title, author, summary) to embeddings model and predict the genre
- [ ] present this prediction to the user somehow
- [ ] give them sliders on 0-5 of how much a given book fits into a given genre

- [ ] store all this in a DB
  - [ ] we get a "predicted genres", per-user genre ratings, and global (agerage of all users) genre

## edition handling

- [ ] for a book in a user's account, user should be able to "change edition" 
  - [ ] this shows the other editions available, lets a user pick one
  - [ ] point of this is mostly to let the users pick what covers appear on their page

## misc

- [ ] we need a bug report form
  - [ ] similarly, a feature request form
- [ ] organize the top bar into dropdown menus
  - [ ] BROWSE -> search books, search by genre, etc
  - [ ] PEOPLE -> my account page, my friends, browse users

- [ ] improve UI for computed collections, they're currently burried in the user page

- [ ] I shouldn't have an option to rate/review books when viewing them on a different user's page
  - [ ] there should be a button witha "want to read" option, which adds this title to that label in my account
  - [ ] dropdown on that button which also lets me add to any other tag/label
  - [ ] also a button in there to review and rate

- [ ] put a caching layer / proxy in front of openbooks API so we don't hit it when we don't need to
  - [ ] consider tradeoffs, not certain we want this 