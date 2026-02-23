import { notFound } from "next/navigation";
import Link from "next/link";
import Nav from "@/components/nav";
import ShelfBookGrid from "@/components/shelf-book-grid";
import { getUser } from "@/lib/auth";

// ── Types ─────────────────────────────────────────────────────────────────────

type ShelfDetail = {
  id: string;
  name: string;
  slug: string;
  exclusive_group: string;
  books: {
    book_id: string;
    open_library_id: string;
    title: string;
    cover_url: string | null;
    added_at: string;
  }[];
};

type ShelfSummary = {
  id: string;
  name: string;
  slug: string;
  exclusive_group: string;
  item_count: number;
};

// ── Data fetchers ──────────────────────────────────────────────────────────────

async function fetchShelf(
  username: string,
  slug: string
): Promise<ShelfDetail | null> {
  const res = await fetch(
    `${process.env.API_URL}/users/${username}/shelves/${slug}`,
    { cache: "no-store" }
  );
  if (!res.ok) return null;
  return res.json();
}

async function fetchUserShelves(username: string): Promise<ShelfSummary[]> {
  const res = await fetch(`${process.env.API_URL}/users/${username}/shelves`, {
    cache: "no-store",
  });
  if (!res.ok) return [];
  return res.json();
}

// ── Page ──────────────────────────────────────────────────────────────────────

export default async function ShelfPage({
  params,
}: {
  params: Promise<{ username: string; slug: string }>;
}) {
  const { username, slug } = await params;
  const [currentUser, shelf, allShelves] = await Promise.all([
    getUser(),
    fetchShelf(username, slug),
    fetchUserShelves(username),
  ]);

  if (!shelf) notFound();

  const isOwner = currentUser?.username === username;

  return (
    <div className="min-h-screen">
      <Nav />
      <main className="max-w-5xl mx-auto px-4 sm:px-6 py-12">
        <nav className="flex items-center gap-2 text-xs text-stone-400 mb-8">
          <Link href={`/${username}`} className="hover:text-stone-700 transition-colors">
            {username}
          </Link>
          <span>/</span>
          <span className="text-stone-600">{shelf.name}</span>
        </nav>

        <div className="flex items-baseline gap-3 mb-8">
          <h1 className="text-2xl font-bold text-stone-900">{shelf.name}</h1>
          <span className="text-sm text-stone-400">
            {shelf.books.length} {shelf.books.length === 1 ? "book" : "books"}
          </span>
        </div>

        {allShelves.length > 1 && (
          <div className="flex flex-wrap gap-2 mb-8">
            {allShelves.map((s) => (
              <Link
                key={s.id}
                href={`/${username}/shelves/${s.slug}`}
                className={`text-sm px-3 py-1 rounded border transition-colors ${
                  s.slug === slug
                    ? "border-stone-800 bg-stone-800 text-white"
                    : "border-stone-200 text-stone-600 hover:border-stone-400 hover:text-stone-900"
                }`}
              >
                {s.name}
                <span className="ml-1.5 text-xs opacity-60">{s.item_count}</span>
              </Link>
            ))}
          </div>
        )}

        <ShelfBookGrid
          shelfId={shelf.id}
          initialBooks={shelf.books}
          isOwner={isOwner}
        />
      </main>
    </div>
  );
}
