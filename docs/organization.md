# Organization

Rosslib has two distinct primitives for organizing books: **tags** and **labels**. Tags are standalone. labels are keys, and books can be given labels for that key. Both tags and labels support nesting.

- Tags
- Labels
- Nested Tags & Labels

---

## Tags

A tag is a flat label applied to a book. These work like any other tagging system.

```
#favorites
#action
#want-to-read
```

Tags are non-exclusive collections with `collection_type = 'tag'`. A book can have any number of tags. Tags are displayed as chips on profile pages and can be browsed at `/{username}/tags/{slug}`.

---

## Nested Tags

Tags can be nested to form subcategories.

Tags can form a hierarchy using `/` as a separator in the slug. Items tagged with nested tags are also "tagged" with their parent tags. A book can be tagged with `#scifi/dystopian`, which means it's also tagged with `#scifi`. A book tagged with `#scifi/dystopian/post-nuclear-war` is tagged with both of its parent tags. However, that book would NOT be reachable with the tag `#post-nuclear-war`.

---

## Labels

Labels are like tags, but... with a label. This is useful if you want to name a field across many books. A user can define a key, and books can be given a label for that key.

Labels can be exclusive (one label allowed per book) or non-exclusive (multiple labels can be selected per book).

Exclusive labels:

- "Klara and the Sun"
  - gifted_from: "ella"
  - #scifi/robots
- "20,000 Leagues Under the Sea"
  - gifted_from: "mom"
  - #scifi

Non-exclusive labels are great for overlapping categories:

- "The Grapes of Wrath"
  - read_in: ["Middle school", "2012"]
- "The Great Gatsby"
  - read_in: ["High School", "2015"]
- "The Reluctant Fundamentalist"
  - read_in: ["High school", "2013", "College", "2019", "Postgrad", "2022"]

---

## Nested Labels

Label values can be nested using `/` as a path separator, exactly like nested tags.

A book labeled `genre: History/Engineering` is also reachable at `genre: History`. A book labeled `genre: History/Engineering/Ancient` is reachable at both `genre: History/Engineering` and `genre: History`.

```
genre: History                   ← matches all books below
genre: History/Engineering       ← matches books at this level and below
genre: History/Engineering/Ancient
```

The nesting is implicit — you only store the most specific value. Parent paths work automatically via a `LIKE` query on value slugs. Sub-labels are returned in the API response so UIs can render drill-down navigation.

---

## Putting it all together

These tools give you lots of options for organizing and querying your books.

## Comparison

| | Tags | Labels |
|---|---|---|
| Structure | Flat or hierarchical | Key/value |
| Cardinality | Any number per book | One key → one value (select-one) or many values (select-multiple) |
| Query by ancestor | Yes — `#scifi` matches `#scifi/dystopian` | Yes — `genre:history` matches `genre:history/engineering` |
| Visible on profile | Yes, as chips | Not currently (on book cards only) |
| Predefined options | No — any slug is valid | Yes — values are defined on the key; free-form extends the list |
| Use case | Genre browsing, mood, theme | Provenance, reading period, personal metadata |

---

## UI surfaces

### Per-book: `BookTagPicker`

`components/book-tag-picker.tsx` — a dropdown attached to each book card in the owner's shelf/grid views. Lazily loads the book's current assignments on first open. Supports toggling predefined values and typing a free-form value.

### Bulk: library manager toolbar

`components/library-manager.tsx` — when the owner selects multiple books in the library manager, the top toolbar shows a **Labels** dropdown. Clicking a value calls `PUT /api/me/books/:olId/tags/:keyId` for every selected book in parallel. Each key in the panel also has a **clear** button that calls `DELETE /api/me/books/:olId/tags/:keyId` across all selected books.

The bulk Labels action works in both shelf-filtered and tag-filtered views (unlike Move/Remove which require knowing the current shelf ID).

### Tag management: settings

`app/settings/tags/page.tsx` — create/delete label categories, set their mode, add/remove predefined values.

---

## Implementation notes

### Tags (collections)

- Stored as `collections` rows with `collection_type = 'tag'`.
- Slug may contain `/` for hierarchy. Slugify is applied per-segment.
- `GET /users/:username/tags/*path` — returns books where slug equals the path or starts with `path/`.
- Queried with: `slug = $path OR slug LIKE $path || '/%'`

### Labels (tag_keys / tag_values / book_tag_values)

- `tag_keys.mode` is `'select_one'` or `'select_multiple'`.
- `book_tag_values` PK is `(user_id, book_id, tag_value_id)` — allows multiple values per key for select_multiple.
- For select_one: existing value for `(user_id, book_id, tag_key_id)` is deleted before inserting.
- For select_multiple: values are inserted individually; each is removed individually.
- Free-form assignment uses `INSERT ... ON CONFLICT DO UPDATE` on `tag_values` to find-or-create the value, then proceeds with the normal assignment flow.
- `tag_values.slug` may contain `/` for nested values (e.g. `history/engineering`). Each segment is slugified independently via `slugifyValue()`. The column is `VARCHAR(255)`.
- `GET /users/:username/labels/:keySlug/*valuePath` queries with `slug = valuePath OR slug LIKE valuePath || '/%'` so parent paths include all descendants.
- The response includes `sub_labels: string[]` — the direct child value paths, derived the same way as `sub_tags` for nested tags.

### API surface

**Tags (collections):**

```
GET  /users/:username/tags/*path
POST /me/shelves  { type: "tag", name: "scifi/dystopian" }
```

**Labels:**

```
GET    /me/tag-keys
POST   /me/tag-keys                              { name, mode }
DELETE /me/tag-keys/:keyId
POST   /me/tag-keys/:keyId/values               { name }          (name may contain "/" for nesting)
DELETE /me/tag-keys/:keyId/values/:valueId

GET    /me/books/:olId/tags
PUT    /me/books/:olId/tags/:keyId              { value_id } or { value_name }
DELETE /me/books/:olId/tags/:keyId              (clears all values for this key)
DELETE /me/books/:olId/tags/:keyId/values/:valueId   (removes one value; select_multiple)

GET    /users/:username/labels/:keySlug/*valuePath   (public; includes sub-values)
```
