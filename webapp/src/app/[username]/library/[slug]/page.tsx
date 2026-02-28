import { notFound } from "next/navigation";
import Link from "next/link";
import ShelfBookGrid from "@/components/shelf-book-grid";
import LibraryManager, { ShelfSummary } from "@/components/library-manager";
import { getUser, getToken } from "@/lib/auth";
import { TagKey } from "@/components/book-tag-picker";
import type { StatusValue } from "@/components/shelf-picker";

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
  description?: string;
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
  slug: string,
  sort?: string
): Promise<ShelfDetail | null> {
  const qs = sort ? `?sort=${sort}` : "";
  const res = await fetch(
    `${process.env.API_URL}/users/${username}/shelves/${slug}${qs}`,
    { cache: "no-store" }
  );
  if (!res.ok) return null;
  return res.json();
}

async function fetchStatusBooks(
  username: string,
  slug: string,
  sort?: string
): Promise<StatusBooksResponse | null> {
  const sortParam = sort ? `&sort=${sort}` : "";
  const res = await fetch(
    `${process.env.API_URL}/users/${username}/books?status=${slug}&limit=1000${sortParam}`,
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

async function fetchTagKeys(username: string): Promise<TagKey[]> {
  const res = await fetch(
    `${process.env.API_URL}/users/${username}/tag-keys`,
    { cache: "no-store" }
  );
  if (!res.ok) return [];
  return res.json();
}

type UserBooksResponse = {
  statuses: { slug: string; name: string; count: number }[];
  unstatused_count: number;
};

async function fetchUserBooksSummary(username: string): Promise<UserBooksResponse> {
  const res = await fetch(
    `${process.env.API_URL}/users/${username}/books?limit=0`,
    { cache: "no-store" }
  );
  if (!res.ok) return { statuses: [], unstatused_count: 0 };
  return res.json();
}

type ViewerTagKey = {
  id: string;
  name: string;
  slug: string;
  mode: string;
  values: StatusValue[];
};

async function fetchViewerTagKeys(token: string): Promise<ViewerTagKey[]> {
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

const SORT_OPTIONS = [
  { value: "date_added", label: "Date added" },
  { value: "title", label: "Title" },
  { value: "author", label: "Author" },
  { value: "rating", label: "Rating" },
] as const;

export default async function ShelfPage({
  params,
  searchParams,
}: {
  params: Promise<{ username: string; slug: string }>;
  searchParams: Promise<{ sort?: string }>;
}) {
  const { username, slug } = await params;
  const { sort: sortParam } = await searchParams;
  const sort = SORT_OPTIONS.some((o) => o.value === sortParam) ? sortParam : "date_added";
  const [currentUser, token] = await Promise.all([getUser(), getToken()]);
  const isOwner = currentUser?.username === username;

  const isStatusSlug = STATUS_SLUGS.has(slug);

  const [shelf, statusBooks, allShelves, tagKeys, userBooksSummary, viewerTagKeys] = await Promise.all([
    isStatusSlug ? Promise.resolve(null) : fetchShelf(username, slug, sort),
    isStatusSlug ? fetchStatusBooks(username, slug, sort) : Promise.resolve(null),
    fetchUserShelves(username),
    isOwner ? fetchTagKeys(username) : Promise.resolve([] as TagKey[]),
    isOwner ? fetchUserBooksSummary(username) : Promise.resolve({ statuses: [], unstatused_count: 0 } as UserBooksResponse),
    !isOwner && token ? fetchViewerTagKeys(token) : Promise.resolve([] as ViewerTagKey[]),
  ]);

  // Extract status values for QuickAddButton (visitor view)
  const viewerStatusKey = viewerTagKeys.find((k) => k.slug === "status") ?? null;
  const statusValues: StatusValue[] = viewerStatusKey?.values ?? [];
  const statusKeyId = viewerStatusKey?.id ?? null;

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

    const statusCounts: Record<string, number> = {};
    const statusList = userBooksSummary.statuses.map((s) => {
      statusCounts[s.slug] = s.count;
      return { slug: s.slug, name: s.name ?? s.slug, count: s.count };
    });
    let total = userBooksSummary.unstatused_count;
    for (const s of userBooksSummary.statuses) {
      total += s.count;
    }
    statusCounts["_all"] = total;
    statusCounts["_unstatused"] = userBooksSummary.unstatused_count;

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
          statusCounts={statusCounts}
          statusList={statusList}
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

        <div className="flex items-baseline gap-3 mb-2">
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

        {shelf?.description && (
          <p className="text-sm text-text-secondary mb-4">{shelf.description}</p>
        )}

        {!shelf?.description && <div className="mb-2" />}

        {books.length > 1 && (
          <div className="flex items-center gap-2 mb-8">
            <span className="text-sm text-text-secondary">Sort by:</span>
            <div className="flex gap-1">
              {SORT_OPTIONS.map((option) => (
                <Link
                  key={option.value}
                  href={`/${username}/library/${slug}${option.value === "date_added" ? "" : `?sort=${option.value}`}`}
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
        )}

        {allShelves.length > 1 && (
          <div className="flex flex-wrap gap-2 mb-8">
            {allShelves
              .filter((s) => s.collection_type !== "tag")
              .map((s) => (
                <Link
                  key={s.id}
                  href={`/${username}/library/${s.slug}`}
                  className={`text-sm px-3 py-1 rounded border transition-colors ${
                    s.slug === slug
                      ? "border-accent bg-accent text-text-inverted"
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
          statusValues={statusKeyId ? statusValues : undefined}
          statusKeyId={statusKeyId ?? undefined}
        />
      </main>
    </div>
  );
}
