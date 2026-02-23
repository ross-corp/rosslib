import Link from "next/link";
import Nav from "@/components/nav";

// ── Types ─────────────────────────────────────────────────────────────────────

type BookResult = {
  key: string;
  title: string;
  authors: string[] | null;
  publish_year: number | null;
  isbn: string[] | null;
  cover_url: string | null;
  edition_count: number;
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

// ── Data fetchers ─────────────────────────────────────────────────────────────

async function searchBooks(q: string): Promise<BookSearchResponse> {
  if (!q.trim()) return { total: 0, results: [] };
  const res = await fetch(
    `${process.env.API_URL}/books/search?q=${encodeURIComponent(q)}`,
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

// ── Page ──────────────────────────────────────────────────────────────────────

export default async function SearchPage({
  searchParams,
}: {
  searchParams: Promise<{ q?: string; type?: string }>;
}) {
  const { q = "", type } = await searchParams;
  const activeTab = type === "people" ? "people" : "books";

  const [bookData, users] = await Promise.all([
    activeTab === "books" ? searchBooks(q) : Promise.resolve({ total: 0, results: [] }),
    activeTab === "people" ? searchUsers(q) : Promise.resolve([]),
  ]);

  const tabLink = (tab: "books" | "people") =>
    q ? `/search?q=${encodeURIComponent(q)}&type=${tab}` : `/search?type=${tab}`;

  return (
    <div className="min-h-screen">
      <Nav />
      <main className="max-w-5xl mx-auto px-4 sm:px-6 py-10">
        <form action="/search" method="get" className="mb-6 max-w-md flex gap-2">
          <input
            name="q"
            type="search"
            defaultValue={q}
            placeholder={activeTab === "books" ? "Search by title..." : "Search by name..."}
            autoFocus
            className="flex-1 px-3 py-2 text-sm border border-stone-300 rounded text-stone-900 placeholder-stone-400 focus:outline-none focus:ring-2 focus:ring-stone-900 focus:border-transparent"
          />
          <input type="hidden" name="type" value={activeTab} />
        </form>

        {/* Tab selector */}
        <div className="flex gap-1 mb-6 border-b border-stone-200">
          <Link
            href={tabLink("books")}
            className={`px-4 py-2 text-sm font-medium border-b-2 -mb-px transition-colors ${
              activeTab === "books"
                ? "border-stone-900 text-stone-900"
                : "border-transparent text-stone-400 hover:text-stone-700"
            }`}
          >
            Books
          </Link>
          <Link
            href={tabLink("people")}
            className={`px-4 py-2 text-sm font-medium border-b-2 -mb-px transition-colors ${
              activeTab === "people"
                ? "border-stone-900 text-stone-900"
                : "border-transparent text-stone-400 hover:text-stone-700"
            }`}
          >
            People
          </Link>
        </div>

        {/* Result count */}
        {q && activeTab === "books" && (
          <p className="text-sm text-stone-400 mb-6">
            {bookData.results.length === 0
              ? `No books found for "${q}"`
              : `${bookData.total.toLocaleString()} result${bookData.total === 1 ? "" : "s"} for "${q}"`}
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
        {activeTab === "books" && bookData.results.length > 0 && (
          <ul className="divide-y divide-stone-100">
            {bookData.results.map((book) => (
              <li key={book.key} className="flex gap-4 py-4">
                {book.cover_url ? (
                  <img
                    src={book.cover_url}
                    alt={book.title}
                    width={48}
                    height={64}
                    className="w-12 h-16 object-cover rounded shrink-0 bg-stone-100"
                  />
                ) : (
                  <div className="w-12 h-16 bg-stone-100 rounded shrink-0" />
                )}
                <div className="flex flex-col justify-center gap-0.5 min-w-0">
                  <span className="text-sm font-medium text-stone-900 truncate">
                    {book.title}
                  </span>
                  {book.authors && book.authors.length > 0 && (
                    <span className="text-xs text-stone-500">
                      {book.authors.slice(0, 3).join(", ")}
                    </span>
                  )}
                  {book.publish_year && (
                    <span className="text-xs text-stone-400">{book.publish_year}</span>
                  )}
                </div>
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
