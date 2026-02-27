import Link from "next/link";

type UserRow = {
  user_id: string;
  username: string;
  display_name: string | null;
  avatar_url: string | null;
};

const PER_PAGE = 20;

const SORT_OPTIONS = [
  { value: "newest", label: "Newest" },
  { value: "books", label: "Most books" },
  { value: "followers", label: "Most followers" },
] as const;

type SortValue = (typeof SORT_OPTIONS)[number]["value"];

async function fetchUsers(page: number, sort: SortValue): Promise<UserRow[]> {
  const res = await fetch(
    `${process.env.API_URL}/users?page=${page}&sort=${sort}`,
    { cache: "no-store" }
  );
  if (!res.ok) return [];
  const data = await res.json();
  return Array.isArray(data) ? data : [];
}

export default async function UsersPage({
  searchParams,
}: {
  searchParams: Promise<{ page?: string; sort?: string }>;
}) {
  const { page: pageParam = "1", sort: sortParam = "newest" } =
    await searchParams;
  const page = Math.max(1, parseInt(pageParam, 10) || 1);
  const sort = SORT_OPTIONS.some((o) => o.value === sortParam)
    ? (sortParam as SortValue)
    : "newest";
  const users = await fetchUsers(page, sort);
  const has_next = users.length >= PER_PAGE;

  function buildHref(params: { page?: number; sort?: string }) {
    const p = params.page ?? page;
    const s = params.sort ?? sort;
    const qs = new URLSearchParams();
    if (p > 1) qs.set("page", String(p));
    if (s !== "newest") qs.set("sort", s);
    const str = qs.toString();
    return `/users${str ? `?${str}` : ""}`;
  }

  return (
    <div className="min-h-screen">
      <main className="max-w-5xl mx-auto px-4 sm:px-6 py-10">
        <div className="flex items-center justify-between mb-8">
          <h1 className="text-2xl font-bold text-text-primary">People</h1>
          <Link
            href="/search"
            className="text-sm text-text-primary hover:text-text-primary transition-colors"
          >
            Search users
          </Link>
        </div>

        <div className="flex items-center gap-2 mb-6">
          <span className="text-sm text-text-secondary">Sort by:</span>
          <div className="flex gap-1">
            {SORT_OPTIONS.map((option) => (
              <Link
                key={option.value}
                href={buildHref({ sort: option.value, page: 1 })}
                className={`text-sm px-3 py-1 rounded-md transition-colors ${
                  sort === option.value
                    ? "bg-surface-2 text-text-primary font-medium"
                    : "text-text-secondary hover:text-text-primary hover:bg-surface-2"
                }`}
              >
                {option.label}
              </Link>
            ))}
          </div>
        </div>

        {users.length === 0 ? (
          <p className="text-sm text-text-primary">No users yet.</p>
        ) : (
          <ul className="divide-y divide-border max-w-md">
            {users.map((user) => (
              <li key={user.user_id}>
                <Link
                  href={`/${user.username}`}
                  className="flex items-center gap-3 py-3 hover:bg-surface-2 -mx-3 px-3 rounded transition-colors"
                >
                  {user.avatar_url ? (
                    <img
                      src={user.avatar_url}
                      alt=""
                      className="w-9 h-9 rounded-full object-cover bg-surface-2 shrink-0"
                    />
                  ) : (
                    <div className="w-9 h-9 rounded-full bg-surface-2 flex items-center justify-center shrink-0">
                      <span className="text-text-tertiary text-sm font-medium select-none">
                        {(user.display_name || user.username)[0].toUpperCase()}
                      </span>
                    </div>
                  )}
                  <div className="flex flex-col">
                    <span className="text-sm font-medium text-text-primary">
                      {user.display_name || user.username}
                    </span>
                    {user.display_name && (
                      <span className="text-xs text-text-tertiary mt-0.5">
                        @{user.username}
                      </span>
                    )}
                  </div>
                </Link>
              </li>
            ))}
          </ul>
        )}

        {(page > 1 || has_next) && (
          <div className="flex items-center gap-4 mt-8">
            {page > 1 ? (
              <Link
                href={buildHref({ page: page - 1 })}
                className="text-sm text-text-primary hover:text-text-primary transition-colors"
              >
                &larr; Previous
              </Link>
            ) : (
              <span />
            )}
            {has_next && (
              <Link
                href={buildHref({ page: page + 1 })}
                className="text-sm text-text-primary hover:text-text-primary transition-colors ml-auto"
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
