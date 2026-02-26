# ROSSLIB

> we will win because we are insane
> we will win because we are regarded

a better version of goodreads

recently I got an email that the "didn't finish shelf" is a feature "coming soon" to goodreads. this is a company with a hideous product and 428 employees. suck my nuts

## guidance

Technical notes are in `docs/documentation/`

TODOs and planning are in `docs/planning/`

## how work gets done

Human brains work in the docs. Agents write the code.

### unchestrator (`/unchestrator`)

The task list curator. Run it manually before handing off to nephewbot overnight.

- Reads `docs/planning/todo.md`, `features.md`, and `completed.md`
- Rewrites rough notes into detailed, agent-ready task specs (endpoints, DB fields, components)
- Graduates items from `features.md` to `todo.md` when they're concrete enough
- Removes stale/completed items, reconciles merged PRs

```
claude -p "/unchestrator"
```

### nephewbot (`/neph`)

Our favorite IC. Runs overnight on a cron, picks tasks off `todo.md`, and ships them as PRs.

- Picks the top unchecked `- [ ]` item from `docs/planning/todo.md`
- Creates a `neph/<slug>` branch, implements the feature, opens a PR to main
- Moves the task to `## Pending PRs` in `todo.md` with the PR link
- Auto-pauses after 3 consecutive failures

The cron (`nephewbot/nephewbot.sh`) runs hourly overnight on Tristan's server with 3 tasks per batch. Logs go to `nephewbot/nephewbot.log.jsonl`.

```
nephewbot/nephewbot.sh status   # check if running/paused
nephewbot/nephewbot.sh pause    # pause without removing cron
nephewbot/nephewbot.sh resume   # unpause + reset failure count
```

### daily workflow

1. During the day: edit `todo.md` and `features.md` with rough ideas, notes, bug reports
2. Before bed: run `/unchestrator` to clean up and spec out the tasks
3. Overnight: nephewbot cron picks up tasks, opens PRs
4. Next morning: review and merge PRs, move items from Pending PRs to `completed.md`
