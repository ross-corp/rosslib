# ROSSLIB

a better version of goodreads

recently I got an email that the "didn't finish shelf" is a feature "coming soon" to goodreads. this is a company with a hideous product and 428 employees. suck my nuts

## ROADMAP

Thigns ROSSLIB supports that Goodreads doesn't:

- more customizable UI
- better book pages with discussion threads, auto-grouped with embeddings / RAGlike
- better wiki-style links of relevant works, see above
- integrations
  - CLI
  - calibre
  - kindle
- community
  - wikis
  - book clubs
  - submitted works (non-ISBN)
- organizing
  - accessiblke API
  - computed collections
- granularity
  - support for treating short stories as independent works (inside of anthologies)

## nephewbot

`nephewbot/` contains a script that runs Claude Code's `/worker` skill on a cron (every 2 hours, 5 iterations). It works through the TODO list in `docs/TODO.md` autonomously. Logs go to `nephewbot/nephewbot.log`.
