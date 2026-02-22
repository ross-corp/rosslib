import Link from "next/link";
import Nav from "@/components/nav";

type SearchResult = {
  user_id: string;
  username: string;
  display_name: string | null;
};

async function searchUsers(q: string): Promise<SearchResult[]> {
  if (!q.trim()) return [];
  const res = await fetch(
    `${process.env.API_URL}/users?q=${encodeURIComponent(q)}`,
    { cache: "no-store" }
  );
  if (!res.ok) return [];
  return res.json();
}

export default async function SearchPage({
  searchParams,
}: {
  searchParams: Promise<{ q?: string }>;
}) {
  const { q = "" } = await searchParams;
  const results = await searchUsers(q);

  return (
    <div className="min-h-screen">
      <Nav />
      <main className="max-w-5xl mx-auto px-4 sm:px-6 py-10">
        <form action="/search" method="get" className="mb-8 max-w-md">
          <input
            name="q"
            type="search"
            defaultValue={q}
            placeholder="Search users..."
            autoFocus
            className="w-full px-3 py-2 text-sm border border-stone-300 rounded text-stone-900 placeholder-stone-400 focus:outline-none focus:ring-2 focus:ring-stone-900 focus:border-transparent"
          />
        </form>

        {q && (
          <p className="text-sm text-stone-400 mb-4">
            {results.length === 0
              ? `No users found for "${q}"`
              : `${results.length} result${results.length === 1 ? "" : "s"} for "${q}"`}
          </p>
        )}

        {results.length > 0 && (
          <ul className="divide-y divide-stone-100 max-w-md">
            {results.map((user) => (
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
