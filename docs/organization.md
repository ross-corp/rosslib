# Organization

Rosslib has three distinct primitives for organizing books: **tags**, **nested tags**, and **labels**. They share the same underlying collection infrastructure but have different semantics and display behaviour.

---

## Tags

A tag is a flat label applied to a book.

```
#favorites
#action
#want-to-lend
```

Tags are non-exclusive collections with `collection_type = 'tag'`. A book can have any number of tags. Tags are displayed as chips on profile pages and can be browsed at `/{username}/tags/{slug}`.

---

## Nested Tags

Tags can form a hierarchy using `/` as a path separator in the slug.

```
#scifi
#scifi/dystopian
#scifi/dystopian/post-nuclear-war
```

**Inheritance is inferred, not stored.** A book explicitly tagged `#scifi/dystopian/post-nuclear-war` is also implicitly reachable via:

- `#scifi` — all books under the scifi subtree
- `#scifi/dystopian` — all books under the dystopian subtree

But **not** via `#post-nuclear-war`, because that slug doesn't exist as an ancestor in the path.

When browsing `/{username}/tags/scifi`, the result set is all books tagged with any slug that starts with `scifi/` (or equals `scifi` exactly). This is a slug-prefix query, not a separate relationship in the database.

**Creating nested tags:** the tag name may include `/` separators. Each segment is independently slugified and rejoined.

```
"Science Fiction/Dystopian" → slug: "science-fiction/dystopian"
```

---

## Labels

A label is a **key/value pair** attached to a book. The user defines the key and its allowed values upfront; each book is then assigned zero or one (select-one) or zero or more (select-multiple) values for each key.

Labels live in two tables:

- `tag_keys` — the category (e.g. "Gifted from", "Read in"). Owned per-user.
- `tag_values` — the predefined options for a key (e.g. "mom", "dad"). Owned per-key.
- `book_tag_values` — which value(s) a user has assigned to a given book for a given key.

### Select-one

Exactly one value may be assigned per key per book. Assigning a new value replaces the previous one.

```
gifted_from: "mom"
gifted_from: "dad"    ← replaces "mom"
```

Defined as:

```
key: "Gifted from"   mode: select_one
values: ["mom", "dad", "kaitlyn", "liam"]
```

A book gets one entry: `gifted_from: "kaitlyn"`.

### Select-multiple

Any number of values may be assigned per key per book. Values are toggled independently.

```
read_in: ["high school", "college", "2023"]
```

Defined as:

```
key: "Read in"   mode: select_multiple
values: ["middle school", "high school", "college", "2019", "2020", "2021", "2022", "2023", "2024"]
```

A book can simultaneously have `read_in: "high school"` and `read_in: "2023"`.

### Free-form values

When assigning a label on a book, you can also type a new value directly in the picker. This creates the value in the predefined list for that key and assigns it to the book in one step. The new value then appears as an option for all future books.

---

## Comparison

| | Tags | Labels |
|---|---|---|
| Structure | Flat or hierarchical | Key/value |
| Cardinality | Any number per book | One key → one value (select-one) or many values (select-multiple) |
| Query by ancestor | Yes — `#scifi` matches `#scifi/dystopian` | No — label lookup is exact key+value |
| Visible on profile | Yes, as chips | Not currently (on book cards only) |
| Predefined options | No — any slug is valid | Yes — values are defined on the key; free-form extends the list |
| Use case | Genre browsing, mood, theme | Provenance, reading period, personal metadata |

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
POST   /me/tag-keys/:keyId/values               { name }
DELETE /me/tag-keys/:keyId/values/:valueId

GET    /me/books/:olId/tags
PUT    /me/books/:olId/tags/:keyId              { value_id } or { value_name }
DELETE /me/books/:olId/tags/:keyId              (clears all values for this key)
DELETE /me/books/:olId/tags/:keyId/values/:valueId   (removes one value; select_multiple)
```
