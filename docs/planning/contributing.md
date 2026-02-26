# Contributing

## Principles

- Human brains plan, agents implement. Write good task specs and let the bots do the rest.
- Outstanding work is tracked in `docs/planning/todo.md`. Completed work lives in `completed.md`.
- Larger ideas that need design work go in `features.md` until they're ready to be broken into tasks.

## Agents

### unchestrator (`.claude/skills/unchestrator/`)

Task list curator. Run manually before overnight runs.

- Refines rough notes in `todo.md` into detailed, self-contained task specs
- Graduates concrete items from `features.md` into `todo.md`
- Removes stale items, reconciles merged/closed PRs
- Only touches files in `docs/planning/` — never modifies code

```
claude -p "/unchestrator"
```

### nephewbot (`.claude/skills/neph/`)

Autonomous worker. Runs overnight via cron (`nephewbot/nephewbot.sh`).

- Picks the top unchecked item from `todo.md`
- Creates a `neph/<slug>` branch, implements the feature, runs verification (go build, tsc, lint)
- Opens a PR to main, moves the task to `## Pending PRs` in todo.md
- Switches back to main and repeats for the next task
- Auto-pauses after 3 consecutive failures

Cron schedule: hourly overnight (10 PM – 8 AM), 3 tasks per batch.

```
nephewbot/nephewbot.sh status   # check state
nephewbot/nephewbot.sh pause    # pause without removing cron
nephewbot/nephewbot.sh resume   # unpause + reset failure count
```

## Daily workflow

1. **During the day**: edit `todo.md` and `features.md` freely — rough notes, ideas, bug reports. develop as you see fit
2. **Before bed**: run `/unchestrator` to clean up task specs
3. **Overnight**: nephewbot cron works through the list, opens PRs
4. **Next morning**: review PRs, merge, move Pending PRs entries to `completed.md`

## Writing good tasks for todo.md

Each item should be implementable by an agent with zero human guidance:

- **Name the API endpoints**: `POST /me/foo`, `GET /users/:id/bar`
- **Name the DB fields**: what collection, what column, what type
- **Name the components**: which file to modify, what the UI should look like
- **Scope to one PR**: if it's too big, break it into sequential pieces
- **Call out what NOT to do**: if there are gotchas or things to avoid

Bad: `- [ ] improve the search`
Good: `- [ ] Add published year range filter to book search. API: accept year_min/year_max query params on GET /books/search, pass to Open Library query. Frontend: add two number inputs above the results list on /search.`

## Guidelines

- Follow existing patterns: handlers in `api/handlers/`, proxy routes in `webapp/src/app/api/`
- Consult `docs/documentation/datamodel.md` and `docs/documentation/sysdesign.md` before designing new features
- Run `cd api && go build .`, `cd webapp && npm run lint`, and `npx tsc --noEmit` before opening a PR
