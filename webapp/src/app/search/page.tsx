import Link from "next/link";
import Nav from "@/components/nav";
import BookList from "@/components/book-list";
import { type StatusValue } from "@/components/shelf-picker";
import { getToken, getUser } from "@/lib/auth";

// ── Types ─────────────────────────────────────────────────────────────────────

type BookResult = {
  key: string;
  title: string;
  authors: string[] | null;
  publish_year: number | null;
  isbn: string[] | null;
  cover_url: string | null;
  edition_count: number;
  average_rating: number | null;
  rating_count: number;
  already_read_count: number;
  subjects: string[] | null;
};

type BookSearchResponse = {
  total: number;
  results: BookResult[];
};

type UserResult = {
  user_id: string;
  username: string;
  display_name: string | null;
};

type AuthorResult = {
  key: string;
  name: string;
  birth_date: string | null;
  death_date: string | null;
  top_work: string | null;
  work_count: number;
  top_subjects: string[] | null;
  photo_url: string | null;
};

type AuthorSearchResponse = {
  total: number;
  results: AuthorResult[];
};

type TagKey = {
  id: string;
  name: string;
  slug: string;
  mode: string;
  values: StatusValue[];
};

// ── Constants ─────────────────────────────────────────────────────────────────

const SORT_OPTIONS = [
  { value: "", label: "Relevance" },
  { value: "reads", label: "Most read" },
  { value: "rating", label: "Highest rated" },
] as const;

const GENRE_OPTIONS = [
  "Fiction",
  "Non-fiction",
  "Fantasy",
  "Science fiction",
  "Mystery",
  "Romance",
  "Horror",
  "Thriller",
  "Biography",
  "History",
  "Poetry",
  "Children",
] as const;

const LANGUAGE_OPTIONS = [
  { code: "eng", label: "English" },
  { code: "spa", label: "Spanish" },
  { code: "fre", label: "French" },
  { code: "ger", label: "German" },
  { code: "ita", label: "Italian" },
  { code: "por", label: "Portuguese" },
  { code: "rus", label: "Russian" },
  { code: "chi", label: "Chinese" },
  { code: "jpn", label: "Japanese" },
  { code: "kor", label: "Korean" },
] as const;

// ── Data fetchers ─────────────────────────────────────────────────────────────

async function searchBooks(
  q: string,
  sort: string,
  yearMin: string,
  yearMax: string,
  subject: string,
  language: string,
): Promise<BookSearchResponse> {
  if (!q.trim()) return { total: 0, results: [] };
  const params = new URLSearchParams({ q });
  if (sort) params.set("sort", sort);
  if (yearMin) params.set("year_min", yearMin);
  if (yearMax) params.set("year_max", yearMax);
  if (subject) params.set("subject", subject);
  if (language) params.set("language", language);
  const res = await fetch(
    `${process.env.API_URL}/books/search?${params}`,
    { cache: "no-store" }
  );
  if (!res.ok) return { total: 0, results: [] };
  return res.json();
}

async function searchUsers(q: string): Promise<UserResult[]> {
  if (!q.trim()) return [];
  const res = await fetch(
    `${process.env.API_URL}/users?q=${encodeURIComponent(q)}`,
    { cache: "no-store" }
  );
  if (!res.ok) return [];
  return res.json();
}

async function searchAuthors(q: string): Promise<AuthorSearchResponse> {
  if (!q.trim()) return { total: 0, results: [] };
  const res = await fetch(
    `${process.env.API_URL}/authors/search?q=${encodeURIComponent(q)}`,
    { cache: "no-store" }
  );
  if (!res.ok) return { total: 0, results: [] };
  return res.json();
}

async function fetchTagKeys(token: string): Promise<TagKey[]> {
  const res = await fetch(`${process.env.API_URL}/me/tag-keys`, {
    headers: { Authorization: `Bearer ${token}` },
    cache: "no-store",
  });
  if (!res.ok) return [];
  return res.json();
}

async function fetchStatusMap(token: string): Promise<Record<string, string>> {
  const res = await fetch(`${process.env.API_URL}/me/books/status-map`, {
    headers: { Authorization: `Bearer ${token}` },
    cache: "no-store",
  });
  if (!res.ok) return {};
  return res.json();
}

// ── Helpers ───────────────────────────────────────────────────────────────────

function buildSearchParams(base: {
  q: string;
  type: string;
  sort: string;
  year_min: string;
  year_max: string;
  subject: string;
  language: string;
}, overrides: Partial<typeof base> = {}) {
  const merged = { ...base, ...overrides };
  const p = new URLSearchParams({ type: merged.type });
  if (merged.q) p.set("q", merged.q);
  if (merged.sort) p.set("sort", merged.sort);
  if (merged.year_min) p.set("year_min", merged.year_min);
  if (merged.year_max) p.set("year_max", merged.year_max);
  if (merged.subject) p.set("subject", merged.subject);
  if (merged.language) p.set("language", merged.language);
  return `/search?${p}`;
}

// ── Page ──────────────────────────────────────────────────────────────────────

export default async function SearchPage({
  searchParams,
}: {
  searchParams: Promise<{
    q?: string;
    type?: string;
    sort?: string;
    year_min?: string;
    year_max?: string;
    subject?: string;
    language?: string;
  }>;
}) {
  const {
    q = "",
    type,
    sort = "",
    year_min = "",
    year_max = "",
    subject = "",
    language = "",
  } = await searchParams;
  const activeTab = type === "authors" ? "authors" : type === "people" ? "people" : "books";

  const base = { q, type: activeTab, sort, year_min, year_max, subject, language };

  const [currentUser, token] = await Promise.all([getUser(), getToken()]);

  const [bookData, users, authorData, tagKeys, statusMap] = await Promise.all([
    activeTab === "books" ? searchBooks(q, sort, year_min, year_max, subject, language) : Promise.resolve({ total: 0, results: [] }),
    activeTab === "people" ? searchUsers(q) : Promise.resolve([]),
    activeTab === "authors" ? searchAuthors(q) : Promise.resolve({ total: 0, results: [] }),
    currentUser && token ? fetchTagKeys(token) : Promise.resolve(null),
    currentUser && token ? fetchStatusMap(token) : Promise.resolve(null),
  ]);

  const statusKey = tagKeys?.find((k) => k.slug === "status") ?? null;
  const statusValues: StatusValue[] | null = statusKey ? statusKey.values : null;
  const statusKeyId: string | null = statusKey?.id ?? null;
  const bookStatusMap: Record<string, string> | null = statusMap;

  const hasYearFilter = year_min || year_max;
  const hasAnyFilter = hasYearFilter || subject || language;

  return (
    <div className="min-h-screen">
      <Nav />
      <main className="max-w-5xl mx-auto px-4 sm:px-6 py-10">
        <form action="/search" method="get" className="mb-6 max-w-md flex gap-2">
          <input
            name="q"
            type="search"
            defaultValue={q}
            placeholder={activeTab === "books" ? "Search by title..." : activeTab === "authors" ? "Search by author name..." : "Search by name..."}
            autoFocus
            className="flex-1 px-3 py-2 text-sm border border-stone-300 rounded text-stone-900 placeholder-stone-400 focus:outline-none focus:ring-2 focus:ring-stone-900 focus:border-transparent"
          />
          <input type="hidden" name="type" value={activeTab} />
          {sort && <input type="hidden" name="sort" value={sort} />}
          {subject && <input type="hidden" name="subject" value={subject} />}
          {language && <input type="hidden" name="language" value={language} />}
        </form>

        {/* Tab selector */}
        <div className="flex gap-1 mb-6 border-b border-stone-200">
          <Link
            href={buildSearchParams(base, { type: "books", sort: "", year_min: "", year_max: "", subject: "", language: "" })}
            className={`px-4 py-2 text-sm font-medium border-b-2 -mb-px transition-colors ${
              activeTab === "books"
                ? "border-stone-900 text-stone-900"
                : "border-transparent text-stone-400 hover:text-stone-700"
            }`}
          >
            Books
          </Link>
          <Link
            href={buildSearchParams(base, { type: "authors", sort: "", year_min: "", year_max: "", subject: "", language: "" })}
            className={`px-4 py-2 text-sm font-medium border-b-2 -mb-px transition-colors ${
              activeTab === "authors"
                ? "border-stone-900 text-stone-900"
                : "border-transparent text-stone-400 hover:text-stone-700"
            }`}
          >
            Authors
          </Link>
          <Link
            href={buildSearchParams(base, { type: "people", sort: "", year_min: "", year_max: "", subject: "", language: "" })}
            className={`px-4 py-2 text-sm font-medium border-b-2 -mb-px transition-colors ${
              activeTab === "people"
                ? "border-stone-900 text-stone-900"
                : "border-transparent text-stone-400 hover:text-stone-700"
            }`}
          >
            People
          </Link>
        </div>

        {/* Sort controls (books only) */}
        {activeTab === "books" && q && (
          <div className="flex items-center gap-2 mb-5">
            <span className="text-xs text-stone-400">Sort by</span>
            {SORT_OPTIONS.map((opt) => (
              <Link
                key={opt.value}
                href={buildSearchParams(base, { sort: opt.value })}
                className={`text-xs px-2.5 py-1 rounded-full border transition-colors ${
                  sort === opt.value
                    ? "border-stone-900 text-stone-900 bg-stone-900 text-white"
                    : "border-stone-300 text-stone-500 hover:border-stone-500 hover:text-stone-700"
                }`}
              >
                {opt.label}
              </Link>
            ))}
          </div>
        )}

        {/* Genre filter (books only) */}
        {activeTab === "books" && q && (
          <div className="flex items-center gap-2 mb-5 flex-wrap">
            <span className="text-xs text-stone-400">Genre</span>
            {GENRE_OPTIONS.map((genre) => (
              <Link
                key={genre}
                href={buildSearchParams(base, { subject: subject === genre.toLowerCase() ? "" : genre.toLowerCase() })}
                className={`text-xs px-2.5 py-1 rounded-full border transition-colors ${
                  subject === genre.toLowerCase()
                    ? "border-stone-900 bg-stone-900 text-white"
                    : "border-stone-300 text-stone-500 hover:border-stone-500 hover:text-stone-700"
                }`}
              >
                {genre}
              </Link>
            ))}
          </div>
        )}

        {/* Language filter (books only) */}
        {activeTab === "books" && q && (
          <div className="flex items-center gap-2 mb-5 flex-wrap">
            <span className="text-xs text-stone-400">Language</span>
            {LANGUAGE_OPTIONS.map((lang) => (
              <Link
                key={lang.code}
                href={buildSearchParams(base, { language: language === lang.code ? "" : lang.code })}
                className={`text-xs px-2.5 py-1 rounded-full border transition-colors ${
                  language === lang.code
                    ? "border-stone-900 bg-stone-900 text-white"
                    : "border-stone-300 text-stone-500 hover:border-stone-500 hover:text-stone-700"
                }`}
              >
                {lang.label}
              </Link>
            ))}
          </div>
        )}

        {/* Year range filter (books only) */}
        {activeTab === "books" && q && (
          <div className="flex items-center gap-2 mb-5">
            <span className="text-xs text-stone-400">Year</span>
            <form action="/search" method="get" className="flex items-center gap-2">
              <input type="hidden" name="type" value="books" />
              <input type="hidden" name="q" value={q} />
              {sort && <input type="hidden" name="sort" value={sort} />}
              {subject && <input type="hidden" name="subject" value={subject} />}
              {language && <input type="hidden" name="language" value={language} />}
              <input
                name="year_min"
                type="number"
                defaultValue={year_min}
                placeholder="From"
                min="0"
                max="2099"
                className="w-20 px-2 py-1 text-xs border border-stone-300 rounded text-stone-900 placeholder-stone-400 focus:outline-none focus:ring-1 focus:ring-stone-900"
              />
              <span className="text-xs text-stone-300">&ndash;</span>
              <input
                name="year_max"
                type="number"
                defaultValue={year_max}
                placeholder="To"
                min="0"
                max="2099"
                className="w-20 px-2 py-1 text-xs border border-stone-300 rounded text-stone-900 placeholder-stone-400 focus:outline-none focus:ring-1 focus:ring-stone-900"
              />
              <button
                type="submit"
                className="text-xs px-2.5 py-1 rounded-full border border-stone-900 bg-stone-900 text-white hover:bg-stone-700 transition-colors"
              >
                Apply
              </button>
              {hasYearFilter && (
                <Link
                  href={buildSearchParams(base, { year_min: "", year_max: "" })}
                  className="text-xs px-2.5 py-1 rounded-full border border-stone-300 text-stone-500 hover:border-stone-500 hover:text-stone-700 transition-colors"
                >
                  Clear
                </Link>
              )}
            </form>
          </div>
        )}

        {/* Clear all filters */}
        {activeTab === "books" && q && hasAnyFilter && (
          <div className="mb-5">
            <Link
              href={buildSearchParams({ q, type: "books", sort, year_min: "", year_max: "", subject: "", language: "" })}
              className="text-xs text-stone-400 hover:text-stone-700 underline transition-colors"
            >
              Clear all filters
            </Link>
          </div>
        )}

        {/* Result count */}
        {q && activeTab === "books" && (
          <p className="text-sm text-stone-400 mb-6">
            {bookData.results.length === 0
              ? `No books found for "${q}"`
              : `${bookData.total.toLocaleString()} result${bookData.total === 1 ? "" : "s"} for "${q}"`}
          </p>
        )}
        {q && activeTab === "authors" && (
          <p className="text-sm text-stone-400 mb-6">
            {authorData.results.length === 0
              ? `No authors found for "${q}"`
              : `${authorData.total.toLocaleString()} result${authorData.total === 1 ? "" : "s"} for "${q}"`}
          </p>
        )}
        {q && activeTab === "people" && (
          <p className="text-sm text-stone-400 mb-6">
            {users.length === 0
              ? `No people found for "${q}"`
              : `${users.length} result${users.length === 1 ? "" : "s"} for "${q}"`}
          </p>
        )}

        {/* Book results */}
        {activeTab === "books" && (
          <BookList
            books={bookData.results}
            statusValues={statusValues}
            statusKeyId={statusKeyId}
            bookStatusMap={bookStatusMap}
          />
        )}

        {/* Author results */}
        {activeTab === "authors" && authorData.results.length > 0 && (
          <ul className="divide-y divide-stone-100 max-w-2xl">
            {authorData.results.map((author) => (
              <li key={author.key}>
                <Link
                  href={`/authors/${author.key}`}
                  className="flex items-start gap-3 py-4 hover:bg-stone-50 -mx-3 px-3 rounded transition-colors"
                >
                  <div className="w-10 h-10 rounded-full bg-stone-100 flex-shrink-0 flex items-center justify-center text-sm font-medium text-stone-500">
                    {author.name.charAt(0)}
                  </div>
                  <div className="flex-1 min-w-0">
                    <span className="text-sm font-medium text-stone-900">
                      {author.name}
                    </span>
                    {(author.birth_date || author.death_date) && (
                      <span className="text-xs text-stone-400 ml-2">
                        {author.birth_date ?? "?"}
                        {" \u2013 "}
                        {author.death_date ?? ""}
                      </span>
                    )}
                    {author.top_work && (
                      <p className="text-xs text-stone-500 mt-0.5 truncate">
                        Best known for <span className="italic">{author.top_work}</span>
                      </p>
                    )}
                    <p className="text-xs text-stone-400 mt-0.5">
                      {author.work_count} work{author.work_count === 1 ? "" : "s"}
                      {author.top_subjects && author.top_subjects.length > 0 && (
                        <span> &middot; {author.top_subjects.slice(0, 3).join(", ")}</span>
                      )}
                    </p>
                  </div>
                </Link>
              </li>
            ))}
          </ul>
        )}

        {/* People results */}
        {activeTab === "people" && users.length > 0 && (
          <ul className="divide-y divide-stone-100 max-w-md">
            {users.map((user) => (
              <li key={user.user_id}>
                <Link
                  href={`/${user.username}`}
                  className="flex flex-col py-3 hover:bg-stone-50 -mx-3 px-3 rounded transition-colors"
                >
                  <span className="text-sm font-medium text-stone-900">
                    {user.display_name || user.username}
                  </span>
                  {user.display_name && (
                    <span className="text-xs text-stone-400 mt-0.5">
                      @{user.username}
                    </span>
                  )}
                </Link>
              </li>
            ))}
          </ul>
        )}
      </main>
    </div>
  );
}
