import Link from "next/link";
import Nav from "@/components/nav";

type UserRow = {
  user_id: string;
  username: string;
  display_name: string | null;
};

const PER_PAGE = 20;

async function fetchUsers(page: number): Promise<UserRow[]> {
  const res = await fetch(
    `${process.env.API_URL}/users?page=${page}`,
    { cache: "no-store" }
  );
  if (!res.ok) return [];
  const data = await res.json();
  return Array.isArray(data) ? data : [];
}

export default async function UsersPage({
  searchParams,
}: {
  searchParams: Promise<{ page?: string }>;
}) {
  const { page: pageParam = "1" } = await searchParams;
  const page = Math.max(1, parseInt(pageParam, 10) || 1);
  const users = await fetchUsers(page);
  const has_next = users.length >= PER_PAGE;

  return (
    <div className="min-h-screen">
      <Nav />
      <main className="max-w-5xl mx-auto px-4 sm:px-6 py-10">
        <div className="flex items-center justify-between mb-8">
          <h1 className="text-2xl font-bold text-stone-900">People</h1>
          <Link
            href="/search"
            className="text-sm text-stone-500 hover:text-stone-900 transition-colors"
          >
            Search users
          </Link>
        </div>

        {users.length === 0 ? (
          <p className="text-sm text-stone-400">No users yet.</p>
        ) : (
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

        {(page > 1 || has_next) && (
          <div className="flex items-center gap-4 mt-8">
            {page > 1 ? (
              <Link
                href={`/users?page=${page - 1}`}
                className="text-sm text-stone-600 hover:text-stone-900 transition-colors"
              >
                &larr; Previous
              </Link>
            ) : (
              <span />
            )}
            {has_next && (
              <Link
                href={`/users?page=${page + 1}`}
                className="text-sm text-stone-600 hover:text-stone-900 transition-colors ml-auto"
              >
                Next &rarr;
              </Link>
            )}
          </div>
        )}
      </main>
    </div>
  );
}
