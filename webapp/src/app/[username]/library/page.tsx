import { redirect } from "next/navigation";
import LibraryManager, { ShelfSummary } from "@/components/library-manager";
import ShelfBookGrid from "@/components/shelf-book-grid";
import { getUser } from "@/lib/auth";
import { TagKey } from "@/components/book-tag-picker";
import Link from "next/link";
import EmptyState from "@/components/empty-state";

type StatusBook = {
  book_id: string;
  open_library_id: string;
  title: string;
  cover_url: string | null;
  added_at: string;
  rating: number | null;
};

type StatusGroup = {
  name: string;
  slug: string;
  count: number;
  books: StatusBook[];
};

type UserBooksResponse = {
  statuses: StatusGroup[];
  unstatused_count: number;
  unstatused_books?: StatusBook[];
};

async function fetchUserBooks(username: string): Promise<UserBooksResponse> {
  const res = await fetch(
    `${process.env.API_URL}/users/${username}/books?limit=50`,
    { cache: "no-store" }
  );
  if (!res.ok) return { statuses: [], unstatused_count: 0 };
  return res.json();
}

async function fetchUserShelves(username: string): Promise<ShelfSummary[]> {
  const res = await fetch(`${process.env.API_URL}/users/${username}/shelves`, {
    cache: "no-store",
  });
  if (!res.ok) return [];
  return res.json();
}

async function fetchTagKeys(username: string): Promise<TagKey[]> {
  const res = await fetch(
    `${process.env.API_URL}/users/${username}/tag-keys`,
    { cache: "no-store" }
  );
  if (!res.ok) return [];
  return res.json();
}

export default async function LibraryIndexPage({
  params,
}: {
  params: Promise<{ username: string }>;
}) {
  const { username } = await params;
  const currentUser = await getUser();
  const isOwner = currentUser?.username === username;

  const [userBooks, allShelves, tagKeys] = await Promise.all([
    fetchUserBooks(username),
    fetchUserShelves(username),
    isOwner ? fetchTagKeys(username) : Promise.resolve([] as TagKey[]),
  ]);

  // Default to showing all books (statused + unstatused)
  const allBooks = [
    ...userBooks.statuses.flatMap((s) => s.books ?? []),
    ...(userBooks.unstatused_books ?? []),
  ];

  if (isOwner) {
    const statusCounts: Record<string, number> = {};
    let total = userBooks.unstatused_count;
    for (const s of userBooks.statuses) {
      statusCounts[s.slug] = s.count;
      total += s.count;
    }
    statusCounts["_all"] = total;
    statusCounts["_unstatused"] = userBooks.unstatused_count;

    const statusList = userBooks.statuses.map((s) => ({
      slug: s.slug,
      name: s.name,
      count: s.count,
    }));

    return (
      <div className="h-screen flex flex-col overflow-hidden">
        <LibraryManager
          username={username}
          initialBooks={allBooks}
          initialShelf={{ id: "all", name: "All Books", slug: "_all" }}
          allShelves={allShelves}
          tagKeys={tagKeys}
          statusCounts={statusCounts}
          statusList={statusList}
        />
      </div>
    );
  }

  // Visitor view: show all statuses with their books
  return (
    <div className="min-h-screen">
      <main className="max-w-5xl mx-auto px-4 sm:px-6 py-12">
        <nav className="flex items-center gap-2 text-xs text-text-primary mb-8">
          <Link
            href={`/${username}`}
            className="hover:text-text-primary transition-colors"
          >
            {username}
          </Link>
          <span>/</span>
          <span className="text-text-primary">Library</span>
        </nav>

        <h1 className="text-2xl font-bold text-text-primary mb-8">Library</h1>

        {userBooks.statuses.length > 0 && (
          <div className="flex flex-wrap gap-2 mb-8">
            {userBooks.statuses.map((s) => (
              <Link
                key={s.slug}
                href={`/${username}/library/${s.slug}`}
                className="text-sm px-3 py-1 rounded border border-border text-text-primary hover:border-border-strong hover:text-text-primary transition-colors"
              >
                {s.name}
                <span className="ml-1.5 text-xs opacity-60">{s.count}</span>
              </Link>
            ))}
          </div>
        )}

        {allBooks.length > 0 ? (
          <ShelfBookGrid
            shelfId="all"
            initialBooks={allBooks}
            isOwner={false}
            tagKeys={[]}
          />
        ) : (
          <EmptyState
            message="No books yet. Search for a book to get started."
            actionLabel="Search books"
            actionHref="/search"
          />
        )}
      </main>
    </div>
  );
}
