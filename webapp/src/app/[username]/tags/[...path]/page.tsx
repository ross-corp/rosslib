import { notFound } from "next/navigation";
import Link from "next/link";
import ShelfBookGrid from "@/components/shelf-book-grid";
import { getUser } from "@/lib/auth";

// ── Types ─────────────────────────────────────────────────────────────────────

type TagBooks = {
  path: string;
  sub_tags: string[];
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

async function fetchTagBooks(
  username: string,
  tagPath: string
): Promise<TagBooks | null> {
  const res = await fetch(
    `${process.env.API_URL}/users/${username}/tags/${tagPath}`,
    { cache: "no-store" }
  );
  if (!res.ok) return null;
  return res.json();
}

// ── Helpers ────────────────────────────────────────────────────────────────────

// Given a full tag path like "sci-fi/moon", return breadcrumb segments:
// [{ label: "sci-fi", path: "sci-fi" }, { label: "moon", path: "sci-fi/moon" }]
function tagBreadcrumbs(
  tagPath: string
): { label: string; path: string }[] {
  const parts = tagPath.split("/");
  return parts.map((part, i) => ({
    label: part,
    path: parts.slice(0, i + 1).join("/"),
  }));
}

// Return the direct parent path, or null if at the root tag level.
function parentPath(tagPath: string): string | null {
  const idx = tagPath.lastIndexOf("/");
  return idx === -1 ? null : tagPath.slice(0, idx);
}

// ── Page ──────────────────────────────────────────────────────────────────────

export default async function TagPage({
  params,
}: {
  params: Promise<{ username: string; path: string[] }>;
}) {
  const { username, path } = await params;
  const tagPath = path.join("/");

  const [currentUser, tagData] = await Promise.all([
    getUser(),
    fetchTagBooks(username, tagPath),
  ]);

  if (!tagData) notFound();

  const breadcrumbs = tagBreadcrumbs(tagPath);
  const parent = parentPath(tagPath);

  // Sort sub_tags so they display consistently
  const subTags = [...(tagData.sub_tags ?? [])].sort();

  return (
    <div className="min-h-screen">
      <main className="max-w-5xl mx-auto px-4 sm:px-6 py-12">
        {/* Breadcrumb */}
        <nav className="flex items-center gap-2 text-xs text-text-primary mb-8 flex-wrap">
          <Link href={`/${username}`} className="hover:text-text-primary transition-colors">
            {username}
          </Link>
          <span>/</span>
          <Link href={`/${username}`} className="hover:text-text-primary transition-colors">
            tags
          </Link>
          {breadcrumbs.map((crumb, i) => (
            <span key={crumb.path} className="flex items-center gap-2">
              <span>/</span>
              {i < breadcrumbs.length - 1 ? (
                <Link
                  href={`/${username}/tags/${crumb.path}`}
                  className="hover:text-text-primary transition-colors"
                >
                  {crumb.label}
                </Link>
              ) : (
                <span className="text-text-primary">{crumb.label}</span>
              )}
            </span>
          ))}
        </nav>

        <div className="flex items-baseline gap-3 mb-6">
          <h1 className="text-2xl font-bold text-text-primary">{tagPath}</h1>
          <span className="text-sm text-text-primary">
            {tagData.books.length} {tagData.books.length === 1 ? "book" : "books"}
          </span>
        </div>

        {/* Sub-tag filters */}
        {subTags.length > 0 && (
          <div className="flex flex-wrap gap-2 mb-8">
            {parent && (
              <Link
                href={`/${username}/tags/${parent}`}
                className="text-sm px-3 py-1 rounded-full border border-border text-text-primary hover:border-border hover:text-text-primary transition-colors"
              >
                ← {parent.split("/").pop()}
              </Link>
            )}
            {subTags.map((sub) => {
              const label = sub.split("/").pop() ?? sub;
              return (
                <Link
                  key={sub}
                  href={`/${username}/tags/${sub}`}
                  className="text-sm px-3 py-1 rounded-full border border-border text-text-primary hover:border-border hover:text-text-primary transition-colors"
                >
                  {label}
                </Link>
              );
            })}
          </div>
        )}

        <ShelfBookGrid
          shelfId=""
          initialBooks={tagData.books}
          isOwner={false}
        />
      </main>
    </div>
  );
}
