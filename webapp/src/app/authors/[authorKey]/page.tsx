import { notFound } from "next/navigation";
import Link from "next/link";
import AuthorBio from "@/components/author-bio";
import AuthorFollowButton from "@/components/author-follow-button";
import AuthorPhoto from "@/components/author-photo";
import AuthorWorksGrid from "@/components/author-works-grid";
import { getToken } from "@/lib/auth";

// ── Types ──────────────────────────────────────────────────────────────────────

type AuthorWork = {
  key: string;
  title: string;
  cover_url: string | null;
};

type AuthorLink = {
  title: string;
  url: string;
};

type AuthorDetail = {
  key: string;
  name: string;
  bio: string | null;
  birth_date: string | null;
  death_date: string | null;
  photo_url: string | null;
  links: AuthorLink[] | null;
  work_count: number;
  works: AuthorWork[] | null;
};

type AuthorSeriesItem = {
  id: string;
  name: string;
  description: string | null;
  book_count: number;
};

// ── Data fetchers ─────────────────────────────────────────────────────────────

async function fetchAuthorSeries(authorKey: string, authorName: string): Promise<AuthorSeriesItem[]> {
  const res = await fetch(
    `${process.env.API_URL}/authors/${authorKey}/series?name=${encodeURIComponent(authorName)}`,
    { cache: "no-store" }
  );
  if (!res.ok) return [];
  return res.json();
}

async function fetchAuthor(authorKey: string): Promise<AuthorDetail | null> {
  const res = await fetch(`${process.env.API_URL}/authors/${authorKey}`, {
    cache: "no-store",
  });
  if (!res.ok) return null;
  return res.json();
}

async function fetchFollowStatus(
  authorKey: string,
  token: string
): Promise<boolean> {
  const res = await fetch(
    `${process.env.API_URL}/authors/${authorKey}/follow`,
    {
      cache: "no-store",
      headers: { Authorization: `Bearer ${token}` },
    }
  );
  if (!res.ok) return false;
  const data = await res.json();
  return data.following === true;
}

// ── Page ───────────────────────────────────────────────────────────────────────

export default async function AuthorPage({
  params,
}: {
  params: Promise<{ authorKey: string }>;
}) {
  const { authorKey } = await params;
  const token = await getToken();

  const [author, following] = await Promise.all([
    fetchAuthor(authorKey),
    token ? fetchFollowStatus(authorKey, token) : Promise.resolve(false),
  ]);

  if (!author) notFound();

  const works = author.works ?? [];
  const authorSeries = await fetchAuthorSeries(authorKey, author.name);

  return (
    <div className="min-h-screen">
      <main className="max-w-3xl mx-auto px-4 sm:px-6 py-12">
        {/* ── Author header ── */}
        <div className="flex gap-8 items-start mb-10">
          {/* Photo */}
          <AuthorPhoto
            src={author.photo_url}
            name={author.name}
            className="w-32 h-40 shrink-0"
          />

          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-3 mb-1">
              <h1 className="text-2xl font-bold text-text-primary">
                {author.name}
              </h1>
              {token && (
                <AuthorFollowButton
                  authorKey={authorKey}
                  authorName={author.name}
                  initialFollowing={following}
                />
              )}
            </div>

            {(author.birth_date || author.death_date) && (
              <p className="text-sm text-text-primary mb-3">
                {author.birth_date ?? "?"}
                {" \u2013 "}
                {author.death_date ?? "present"}
              </p>
            )}

            <p className="text-xs text-text-primary mb-4">
              {author.work_count} work{author.work_count === 1 ? "" : "s"}
            </p>

            {author.bio && <AuthorBio bio={author.bio} />}

            {author.links && author.links.length > 0 && (
              <div className="flex flex-wrap gap-3 mt-4">
                {author.links.map((link) => (
                  <a
                    key={link.url}
                    href={link.url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-xs text-text-primary underline hover:text-text-primary transition-colors"
                  >
                    {link.title}
                  </a>
                ))}
              </div>
            )}
          </div>
        </div>

        {/* ── Series ── */}
        {authorSeries.length > 0 && (
          <section className="border-t border-border pt-8 mb-10">
            <h2 className="text-sm font-semibold text-text-primary uppercase tracking-wider mb-4">
              Series ({authorSeries.length})
            </h2>
            <div className="space-y-2">
              {authorSeries.map((s) => (
                <Link
                  key={s.id}
                  href={`/series/${s.id}`}
                  className="flex items-center justify-between border border-border rounded-lg px-4 py-3 hover:border-border-strong transition-colors"
                >
                  <div>
                    <span className="text-sm font-medium text-text-primary">
                      {s.name}
                    </span>
                    {s.description && (
                      <p className="text-xs text-text-tertiary mt-0.5 line-clamp-1">
                        {s.description}
                      </p>
                    )}
                  </div>
                  <span className="text-xs text-text-tertiary shrink-0 ml-4">
                    {s.book_count} {s.book_count === 1 ? "book" : "books"}
                  </span>
                </Link>
              ))}
            </div>
          </section>
        )}

        {/* ── Works ── */}
        <section className="border-t border-border pt-8">
          <h2 className="text-sm font-semibold text-text-primary uppercase tracking-wider mb-6">
            {works.length > 0
              ? `Works (${author.work_count})`
              : "Works"}
          </h2>

          <AuthorWorksGrid
            authorKey={authorKey}
            initialWorks={works}
            workCount={author.work_count}
          />
        </section>
      </main>
    </div>
  );
}
