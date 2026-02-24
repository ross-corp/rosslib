# Ghost Users

Bot accounts that simulate real reading activity to populate social features (feeds, follows, reviews) during development and early usage. Not intended for production deception — they exist so the app doesn't feel empty and so we can test inter-account connectivity end-to-end.

---

## Goals

- 4 ghost accounts with distinct reading tastes and pacing
- On-demand activity generation — click a button in the app or hit an API endpoint and all ghosts do something
- Identifiable as ghosts in the database so an admin can filter them out
- No deployment overhead: an internal API endpoint + a page in the webapp

---

## Ghost Personas

Each ghost has a reading personality that determines what books it picks and how fast it moves through them.

| Username | Display Name | Personality | Pace |
|---|---|---|---|
| `ghost-jeff` | jeff | libertarian politics, scifi classics | Slow — finishes a book every 5–7 days |
| `ghost-goob` | goob | Sci-fi, fantasy, long series | Medium — finishes every 3–5 days |
| `ghost-casey` | casey | mathematics, political extremism | Fast — finishes every 1–3 days |
| `ghost-bobert` | Bobert | Romance, mystery, thrillers | Fast — finishes every 1–3 days, reviews everything |

All ghost accounts are public (`is_private = false`). Their emails follow the pattern `ghost-<name>@rosslib.local`.

---

## Schema Change

Add a column to `users`:

```sql
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_ghost BOOLEAN DEFAULT false;
```

Added via `db.Migrate` as usual. Ghost users are created with `is_ghost = true`. This column is the single source of truth for filtering — no need to parse usernames.

---

## Book Catalog

Each ghost draws from a curated seed list of real Open Library work IDs. The lists are hardcoded in the ghost binary, roughly 40–60 books per persona. Books are upserted into the `books` table on first use (same path as the normal `AddBookToShelf` flow — look up via OL API, upsert locally).

We use real OL IDs so book pages, covers, and metadata all work naturally. Example seeds:

```
ghost-jeff:   OL82592W (1984), OL1168083W (Atlas Shrugged), OL46234W (Starship Troopers), ...
ghost-goob:   OL27479W (Dune), OL45883W (Neuromancer), OL17930368W (The Fifth Season), ...
ghost-casey:  OL7956004W (Gödel, Escher, Bach), OL151411W (A Short History of Nearly Everything), ...
ghost-bobert: OL47755W (Gone Girl), OL27776W (Pride and Prejudice), OL20930735W (The Love Hypothesis), ...
```

The full lists live in the source code. Some overlap between personas is fine — it makes cross-user book pages more interesting.

---

## Activity Simulation

### How It Works

A new package at `api/internal/ghosts/` exposes a `Simulate(ctx, pool)` function. Each call triggers one round of activity for all ghost users. There's no pacing or cooldown — every call produces visible results, so you can spam it during testing.

Each round, for each ghost:

1. **Pick 1–3 random actions** from the action pool (weighted by persona):
   - **Shelve a new book** — pick an unshelved book from their seed list, add to "Want to Read"
   - **Start reading** — move a random "Want to Read" book to "Currently Reading"
   - **Finish a book** — move a "Currently Reading" book to "Read", assign a rating, maybe write a review
   - **Follow someone** — follow a random user (ghost or real) they aren't already following
2. Each action goes through the normal handler logic so activity records, exclusive shelf enforcement, etc. all work correctly

No time-based pacing. The ghosts don't check how long they've been "reading" — every call to simulate just advances their state. Hit it 10 times and they'll burn through 10 rounds of activity.

### API Endpoint

```
POST /admin/ghosts/simulate
```

Protected by auth middleware. Returns a summary of what each ghost did:

```json
{
  "results": [
    { "username": "ghost-jeff", "actions": ["shelved OL82592W", "started OL1168083W"] },
    { "username": "ghost-goob", "actions": ["finished OL27479W (rated 4)", "followed ghost-casey"] },
    ...
  ]
}
```

For now, any logged-in user can trigger it. If we add admin roles later, gate it behind that.

### Webapp Page

A simple page at `/settings/ghost-activity` (or similar, linked from settings):

- **"Simulate Round" button** — calls `POST /admin/ghosts/simulate`, shows the results
- **Ghost status cards** — for each ghost, show what they're currently reading, how many books they've read, who they follow
- **"Seed Ghosts" button** — creates the 4 accounts if they don't exist (calls `POST /admin/ghosts/seed`)

This is a dev/admin tool, not a user-facing feature. Keep it minimal.

---

## Reviews

Ghosts generate short, genre-appropriate reviews. These can be:

- A small hardcoded pool of template sentences per persona (e.g., Bobert: "Couldn't put it down!", "The twist at the end!", etc.)
- Randomly assembled from fragments: `[opener] + [middle] + [closer]`

Keep it simple. 1–3 sentences max. Not every finished book gets a review — roughly 50% for most personas, 90% for Bobert.

Ratings follow a distribution: mostly 3–5 stars (ghosts are reading books they'd presumably enjoy). Bobert skews higher (4–5), Jeff is more critical (2–4).

---

## Social Graph

On first setup (or when a new real user registers), the ghost bot establishes follows:

- All 4 ghosts follow each other (12 follow edges)
- Ghosts follow any real users they discover (the "follow someone" action picks a random non-followed user)
- Real users can follow ghosts normally through the UI

This means new users immediately see ghost activity in their feed once they follow a ghost or a ghost follows them.

---

## Admin Filtering

### API

Add an optional query parameter to feed/activity endpoints:

- `GET /me/feed?exclude_ghosts=true` — filters out activities where `users.is_ghost = true`

The profile endpoint (`GET /users/:username`) already returns all user fields, so `is_ghost` will be available to the frontend once added to the schema.

### Webapp

- Ghost profiles show a small "Bot" badge next to the display name
- The feed page gets a toggle: "Hide bot activity" (persisted in localStorage)
- The `/users` browse page can optionally filter out ghosts
- Admin-only: no admin system yet, so for now filtering is user-facing and opt-in

---

## Setup & Teardown

### Initial Setup

1. Run the API to apply the `is_ghost` migration
2. Hit `POST /admin/ghosts/seed` (or click "Seed Ghosts" in the webapp) — creates the 4 accounts + default shelves
3. Hit `POST /admin/ghosts/simulate` a few times to generate initial activity

Seeding is idempotent — skips any ghost username that already exists.

### Teardown

To remove all ghost data:

```sql
-- Delete in dependency order
DELETE FROM activities WHERE user_id IN (SELECT id FROM users WHERE is_ghost = true);
DELETE FROM collection_items WHERE collection_id IN (
  SELECT id FROM collections WHERE user_id IN (SELECT id FROM users WHERE is_ghost = true)
);
DELETE FROM collections WHERE user_id IN (SELECT id FROM users WHERE is_ghost = true);
DELETE FROM follows WHERE follower_id IN (SELECT id FROM users WHERE is_ghost = true)
   OR followee_id IN (SELECT id FROM users WHERE is_ghost = true);
DELETE FROM users WHERE is_ghost = true;
```

Or just soft-delete: `UPDATE users SET deleted_at = NOW() WHERE is_ghost = true`.

Could also add a `POST /admin/ghosts/teardown` endpoint if we want this from the UI.

---

## Implementation Order

1. **Schema**: add `is_ghost` column to `users` table in `db.Migrate`
2. **Seed list**: curate OL work IDs for each persona
3. **`api/internal/ghosts/`**: the package — seed + simulate functions
4. **API routes**: `POST /admin/ghosts/seed`, `POST /admin/ghosts/simulate`
5. **Webapp page**: ghost control panel at `/settings/ghost-activity`
6. **Filtering**: add `is_ghost` to user response DTOs, bot badge, feed toggle

Steps 1–4 are the core. Steps 5–6 are polish and can come later.

---

## Open Questions

- Should ghosts create discussion threads on books? Would make thread pages less empty, but generating coherent thread titles/bodies is harder than short reviews.
- Should ghosts have avatars? Could use placeholder illustrations to make profiles feel complete.
- Do we want ghost activity to backfill (create historical activity with past timestamps) or only generate forward from first run?
