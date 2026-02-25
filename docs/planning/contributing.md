# Contributing

## Principles

- Most features are implemented by AI agents.
- Outstanding work is tracked in `TODO.md`. Items move to `completed.md` when done (compact as needed).
- **nephewbot** picks up tasks from `TODO.md` on a cron (every 2 hours on Tristan's machine) and works them in isolated worktrees.
  - for now this is just a cron that runs on HUEY (tristan's big server). it invokes claude code with `--dangerously-skip-permissions` and tells it to take an item off todo.md, implement, and push. 
  - the point here is to work through the long feature list and get the most out of my claude max plan

## Workflow

1. Add tasks to `docs/TODO.md` with enough context for an agent to implement them.
2. nephewbot (or a manual agent invocation) picks up the task, creates a branch, and opens a PR.
3. Review and merge the PR into `main`. CI runs tests, lint, and typecheck on every PR.

## Guidelines

- Keep your branch up to date with `main` before pushing.
- Follow existing patterns: handler structs in `api/internal/`, proxy routes in `webapp/src/app/api/`.
- Consult `docs/datamodel.md` and `docs/sysdesign.md` before designing new features.
- Run `go test ./...`, `npm run lint`, and `npx tsc --noEmit` before opening a PR.

