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

### rosslib API route

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

---

## Planned additions

### Google Books API

Google Books provides richer metadata (description, page count, categories, publisher) and higher-resolution covers. It requires an API key but has a generous free tier (1,000 req/day without a key, more with one).

Use case: supplement Open Library results with description and page count when displaying a full book page. Do not replace Open Library as the primary search source — Google's catalog has gaps for older/academic works.

**Endpoint:** `GET https://www.googleapis.com/books/v1/volumes?q=intitle:<title>&key=<key>`

### ISBN-based lookup

For exact ISBN lookups (e.g. during Goodreads CSV import), prefer the Open Library Books API over search:

```
GET https://openlibrary.org/api/books?bibkeys=ISBN:<isbn>&format=json&jscmd=data
```

This returns a single authoritative record for a given ISBN with full metadata including publisher, number of pages, subjects, and identifiers.

---

## Data model notes

- Store the Open Library `key` (`/works/OL...`) as the canonical identifier for a work in our `works` table.
- Store ISBNs in a separate `editions` table keyed to a `work_id`, since a work can have many ISBNs across editions.
- Cache Open Library responses in Redis to avoid hammering the API during popular searches. A short TTL (1–6 hours) is appropriate; catalog data rarely changes quickly.
