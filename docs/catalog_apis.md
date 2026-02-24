# Catalog APIs

This document covers the external book data APIs used by rosslib and the strategy for integrating them.

---

## Current: Open Library (openlibrary.org)

Open Library is the primary catalog API. It is free, requires no API key, and covers tens of millions of works via the Internet Archive.

### Search

**Endpoint:** `GET https://openlibrary.org/search.json`

| Param    | Description                                              |
|----------|----------------------------------------------------------|
| `title`  | Title query string (URL-encoded)                         |
| `fields` | Comma-separated list of fields to return                 |
| `limit`  | Max results (we use 20)                                  |

**Fields we request:**

| Field               | Description                                      |
|---------------------|--------------------------------------------------|
| `key`               | Work key, e.g. `/works/OL82592W`                 |
| `title`             | Book title                                       |
| `author_name`       | Array of author display names                    |
| `first_publish_year`| Year of earliest known publication               |
| `isbn`              | Array of ISBNs across all editions               |
| `cover_i`           | Cover image ID (integer)                         |
| `edition_count`     | Number of editions indexed                       |

**Cover images:**

Cover images are served by Open Library at:

```
https://covers.openlibrary.org/b/id/{cover_i}-{size}.jpg
```

Sizes: `S` (small), `M` (medium), `L` (large). We use `M`.

**Work page URL:**

```
https://openlibrary.org{key}
```

### Author Search

**Endpoint:** `GET https://openlibrary.org/search/authors.json`

| Param  | Description                       |
|--------|-----------------------------------|
| `q`    | Author name query (URL-encoded)   |
| `limit`| Max results (we use 20)           |

**Fields returned per author:**

| Field          | Description                                  |
|----------------|----------------------------------------------|
| `key`          | Author key, e.g. `OL26320A`                 |
| `name`         | Author display name                          |
| `birth_date`   | Birth date string (nullable)                 |
| `death_date`   | Death date string (nullable)                 |
| `top_work`     | Title of most popular work                   |
| `work_count`   | Number of works attributed                   |
| `top_subjects` | Array of most common subjects                |

**Author photo images:**

```
https://covers.openlibrary.org/a/olid/{key}-{size}.jpg
```

Sizes: `S`, `M`, `L`. Availability is inconsistent — many authors have no photo.

### rosslib API routes

Our backend proxies the title search:

```
GET /books/search?q=<title>
```

Response shape:

```json
{
  "total": 1234,
  "results": [
    {
      "key": "/works/OL82592W",
      "title": "The Great Gatsby",
      "authors": ["F. Scott Fitzgerald"],
      "publish_year": 1925,
      "isbn": ["9780743273565"],
      "cover_url": "https://covers.openlibrary.org/b/id/8410459-M.jpg",
      "edition_count": 120
    }
  ]
}
```

`authors`, `isbn`, and `cover_url` can be `null` if Open Library does not have that data for a given work.

Author search:

```
GET /authors/search?q=<name>
```

Response shape:

```json
{
  "total": 41,
  "results": [
    {
      "key": "OL26320A",
      "name": "J.R.R. Tolkien",
      "birth_date": "3 January 1892",
      "death_date": "2 September 1973",
      "top_work": "The Hobbit",
      "work_count": 392,
      "top_subjects": ["Fiction", "Fantasy", "Juvenile fiction"],
      "photo_url": "https://covers.openlibrary.org/a/olid/OL26320A-M.jpg"
    }
  ]
}
```

`birth_date`, `death_date`, `top_work`, `top_subjects`, and `photo_url` can be `null`.

### Author Detail

**Endpoint:** `GET https://openlibrary.org/authors/{key}.json`

Returns full author metadata including biography, dates, photos, and external links.

**Author Works Endpoint:** `GET https://openlibrary.org/authors/{key}/works.json?limit=20`

Returns a paginated list of the author's works with `entries` array and `size` (total count).

Our backend wraps these as:

```
GET /authors/:authorKey
```

Response shape:

```json
{
  "key": "OL26320A",
  "name": "J.R.R. Tolkien",
  "bio": "John Ronald Reuel Tolkien...",
  "birth_date": "3 January 1892",
  "death_date": "2 September 1973",
  "photo_url": "https://covers.openlibrary.org/a/olid/OL26320A-L.jpg",
  "links": [{ "title": "Wikipedia", "url": "https://en.wikipedia.org/wiki/J._R._R._Tolkien" }],
  "work_count": 392,
  "works": [
    {
      "key": "OL82592W",
      "title": "The Hobbit",
      "cover_url": "https://covers.openlibrary.org/b/id/8406786-M.jpg"
    }
  ]
}
```

`bio`, `birth_date`, `death_date`, `photo_url`, `links`, and `works` can be `null`.

---

## Planned additions

### Google Books API

Google Books provides richer metadata (description, page count, categories, publisher) and higher-resolution covers. It requires an API key but has a generous free tier (1,000 req/day without a key, more with one).

Use case: supplement Open Library results with description and page count when displaying a full book page. Do not replace Open Library as the primary search source — Google's catalog has gaps for older/academic works.

**Endpoint:** `GET https://www.googleapis.com/books/v1/volumes?q=intitle:<title>&key=<key>`

### ISBN-based lookup

For exact ISBN lookups (e.g. during Goodreads CSV import), we use the Open Library search API with an `isbn=` param:

```
GET https://openlibrary.org/search.json?isbn=<isbn>&fields=key,title,author_name,first_publish_year,cover_i&limit=1
```

Our backend wraps this as:

```
GET /books/lookup?isbn=<isbn>
```

Returns the matched book (upserted into the local `books` table) or 404 if not found. The implementation strips non-digit characters from the ISBN before querying. OL IDs are stored as bare work IDs (`OL82592W`) without the `/works/` path prefix.

The Goodreads import pipeline calls `LookupBookByISBN` with a nil pool during the preview phase (no DB writes) and with the real pool during commit.

---

## Data model notes

- Store the Open Library `key` (`/works/OL...`) as the canonical identifier for a work in our `works` table.
- Store ISBNs in a separate `editions` table keyed to a `work_id`, since a work can have many ISBNs across editions.
- Cache Open Library responses in Redis to avoid hammering the API during popular searches. A short TTL (1–6 hours) is appropriate; catalog data rarely changes quickly.
