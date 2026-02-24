import { notFound } from "next/navigation";
import Link from "next/link";
import Nav from "@/components/nav";
import ShelfBookGrid from "@/components/shelf-book-grid";

// ── Types ─────────────────────────────────────────────────────────────────────

type LabelBooks = {
  key_slug: string;
  key_name: string;
  value_slug: string;
  value_name: string;
  sub_labels: string[];
  books: {
    book_id: string;
    open_library_id: string;
    title: string;
    cover_url: string | null;
    added_at: string;
    rating: number | null;
  }[];
};

// ── Data fetcher ───────────────────────────────────────────────────────────────

async function fetchLabelBooks(
  username: string,
  keySlug: string,
  valuePath: string
): Promise<LabelBooks | null> {
  const res = await fetch(
    `${process.env.API_URL}/users/${username}/labels/${keySlug}/${valuePath}`,
    { cache: "no-store" }
  );
  if (!res.ok) return null;
  return res.json();
}

// ── Helpers ────────────────────────────────────────────────────────────────────

function valueBreadcrumbs(
  valuePath: string
): { label: string; path: string }[] {
  const parts = valuePath.split("/");
  return parts.map((part, i) => ({
    label: part,
    path: parts.slice(0, i + 1).join("/"),
  }));
}

function parentPath(valuePath: string): string | null {
  const idx = valuePath.lastIndexOf("/");
  return idx === -1 ? null : valuePath.slice(0, idx);
}

// ── Page ──────────────────────────────────────────────────────────────────────

export default async function LabelPage({
  params,
}: {
  params: Promise<{ username: string; keySlug: string; valuePath: string[] }>;
}) {
  const { username, keySlug, valuePath } = await params;
  const valuePathStr = valuePath.join("/");

  const labelData = await fetchLabelBooks(username, keySlug, valuePathStr);

  if (!labelData) notFound();

  const breadcrumbs = valueBreadcrumbs(valuePathStr);
  const parent = parentPath(valuePathStr);
  const subLabels = [...(labelData.sub_labels ?? [])].sort();

  return (
    <div className="min-h-screen">
      <Nav />
      <main className="max-w-5xl mx-auto px-4 sm:px-6 py-12">
        {/* Breadcrumb */}
        <nav className="flex items-center gap-2 text-xs text-stone-400 mb-8 flex-wrap">
          <Link href={`/${username}`} className="hover:text-stone-700 transition-colors">
            {username}
          </Link>
          <span>/</span>
          <span className="text-stone-500">{labelData.key_name}</span>
          {breadcrumbs.map((crumb, i) => (
            <span key={crumb.path} className="flex items-center gap-2">
              <span>/</span>
              {i < breadcrumbs.length - 1 ? (
                <Link
                  href={`/${username}/labels/${keySlug}/${crumb.path}`}
                  className="hover:text-stone-700 transition-colors"
                >
                  {crumb.label}
                </Link>
              ) : (
                <span className="text-stone-600">{crumb.label}</span>
              )}
            </span>
          ))}
        </nav>

        <div className="flex items-baseline gap-3 mb-6">
          <h1 className="text-2xl font-bold text-stone-900">
            {labelData.key_name}: {labelData.value_name}
          </h1>
          <span className="text-sm text-stone-400">
            {labelData.books.length} {labelData.books.length === 1 ? "book" : "books"}
          </span>
        </div>

        {/* Sub-label filters */}
        {subLabels.length > 0 && (
          <div className="flex flex-wrap gap-2 mb-8">
            {parent && (
              <Link
                href={`/${username}/labels/${keySlug}/${parent}`}
                className="text-sm px-3 py-1 rounded-full border border-stone-200 text-stone-500 hover:border-stone-400 hover:text-stone-700 transition-colors"
              >
                ← {parent.split("/").pop()}
              </Link>
            )}
            {subLabels.map((sub) => {
              const label = sub.split("/").pop() ?? sub;
              return (
                <Link
                  key={sub}
                  href={`/${username}/labels/${keySlug}/${sub}`}
                  className="text-sm px-3 py-1 rounded-full border border-stone-200 text-stone-600 hover:border-stone-400 hover:text-stone-900 transition-colors"
                >
                  {label}
                </Link>
              );
            })}
          </div>
        )}

        <ShelfBookGrid
          shelfId=""
          initialBooks={labelData.books}
          isOwner={false}
        />
      </main>
    </div>
  );
}
