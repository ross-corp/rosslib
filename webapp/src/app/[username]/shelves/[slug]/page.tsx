import { notFound } from "next/navigation";
import Link from "next/link";
import ShelfBookGrid from "@/components/shelf-book-grid";
import LibraryManager, { ShelfSummary } from "@/components/library-manager";
import { getUser, getToken } from "@/lib/auth";
import { TagKey } from "@/components/book-tag-picker";

// ── Types ─────────────────────────────────────────────────────────────────────

type ComputedInfo = {
  operation: string;
  is_continuous: boolean;
  last_computed_at: string;
};

type ShelfDetail = {
  id: string;
  name: string;
  slug: string;
  exclusive_group: string;
  computed?: ComputedInfo;
  books: {
    book_id: string;
    open_library_id: string;
    title: string;
    cover_url: string | null;
    added_at: string;
    rating: number | null;
  }[];
};

type StatusBooksResponse = {
  books: {
    book_id: string;
    open_library_id: string;
    title: string;
    cover_url: string | null;
    authors: string | null;
    rating: number | null;
    added_at: string;
  }[];
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

async function fetchStatusBooks(
  username: string,
  slug: string
): Promise<StatusBooksResponse | null> {
  const res = await fetch(
    `${process.env.API_URL}/users/${username}/books?status=${slug}`,
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

async function fetchTagKeys(token: string): Promise<TagKey[]> {
  const res = await fetch(`${process.env.API_URL}/me/tag-keys`, {
    headers: { Authorization: `Bearer ${token}` },
    cache: "no-store",
  });
  if (!res.ok) return [];
  return res.json();
}

// Status slugs that should be fetched from the books endpoint instead of shelves
const STATUS_SLUGS = new Set([
  "want-to-read",
  "owned-to-read",
  "currently-reading",
  "finished",
  "dnf",
]);

// ── Page ──────────────────────────────────────────────────────────────────────

export default async function ShelfPage({
  params,
}: {
  params: Promise<{ username: string; slug: string }>;
}) {
  const { username, slug } = await params;
  const [currentUser, token] = await Promise.all([getUser(), getToken()]);
  const isOwner = currentUser?.username === username;

  const isStatusSlug = STATUS_SLUGS.has(slug);

  const [shelf, statusBooks, allShelves, tagKeys] = await Promise.all([
    isStatusSlug ? Promise.resolve(null) : fetchShelf(username, slug),
    isStatusSlug ? fetchStatusBooks(username, slug) : Promise.resolve(null),
    fetchUserShelves(username),
    isOwner && token ? fetchTagKeys(token) : Promise.resolve([] as TagKey[]),
  ]);

  // Build books array from either source
  const books = statusBooks?.books?.map((b) => ({
    book_id: b.book_id,
    open_library_id: b.open_library_id,
    title: b.title,
    cover_url: b.cover_url,
    added_at: b.added_at,
    rating: b.rating,
  })) ?? shelf?.books ?? [];

  const displayName = isStatusSlug
    ? slug.split("-").map((w) => w.charAt(0).toUpperCase() + w.slice(1)).join(" ")
    : shelf?.name ?? slug;

  if (!isStatusSlug && !shelf) notFound();

  // ── Owner view: full library manager layout ────────────────────────────────

  if (isOwner) {
    // For status slugs, synthesize a shelf-like object for LibraryManager
    const initialShelf = shelf
      ? { id: shelf.id, name: shelf.name, slug: shelf.slug }
      : { id: `status-${slug}`, name: displayName, slug };

    return (
      <div className="h-screen flex flex-col overflow-hidden">
        {shelf?.computed?.is_continuous && (
          <div className="bg-blue-50 border-b border-blue-200 px-4 py-2 text-sm text-blue-700 flex items-center gap-2">
            <span className="font-medium">Live list</span>
            <span className="text-blue-500">&middot;</span>
            <span>
              This list auto-updates from a{" "}
              {shelf.computed.operation} operation.
            </span>
          </div>
        )}
        <LibraryManager
          username={username}
          initialBooks={books}
          initialShelf={initialShelf}
          allShelves={allShelves}
          tagKeys={tagKeys}
        />
      </div>
    );
  }

  // ── Visitor view: classic layout ───────────────────────────────────────────

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
          <span className="text-text-primary">{displayName}</span>
        </nav>

        <div className="flex items-baseline gap-3 mb-8">
          <h1 className="text-2xl font-bold text-text-primary">{displayName}</h1>
          <span className="text-sm text-text-primary">
            {books.length} {books.length === 1 ? "book" : "books"}
          </span>
          {shelf?.computed?.is_continuous && (
            <span className="text-xs px-2 py-0.5 rounded bg-blue-50 text-blue-700 border border-blue-200">
              Live &middot; {shelf.computed.operation}
            </span>
          )}
        </div>

        {allShelves.length > 1 && (
          <div className="flex flex-wrap gap-2 mb-8">
            {allShelves
              .filter((s) => s.collection_type !== "tag")
              .map((s) => (
                <Link
                  key={s.id}
                  href={`/${username}/shelves/${s.slug}`}
                  className={`text-sm px-3 py-1 rounded border transition-colors ${
                    s.slug === slug
                      ? "border-accent bg-accent text-white"
                      : "border-border text-text-primary hover:border-border hover:text-text-primary"
                  }`}
                >
                  {s.name}
                  <span className="ml-1.5 text-xs opacity-60">
                    {s.item_count}
                  </span>
                </Link>
              ))}
          </div>
        )}

        <ShelfBookGrid
          shelfId={shelf?.id ?? `status-${slug}`}
          initialBooks={books}
          isOwner={false}
          tagKeys={[]}
        />
      </main>
    </div>
  );
}
