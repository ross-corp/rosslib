import { notFound } from "next/navigation";
import Link from "next/link";
import Nav from "@/components/nav";
import AuthorFollowButton from "@/components/author-follow-button";
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

// ── Data fetchers ─────────────────────────────────────────────────────────────

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

  return (
    <div className="min-h-screen">
      <Nav />
      <main className="max-w-3xl mx-auto px-4 sm:px-6 py-12">
        {/* ── Author header ── */}
        <div className="flex gap-8 items-start mb-10">
          {/* Photo */}
          {author.photo_url ? (
            <img
              src={author.photo_url}
              alt={author.name}
              className="w-32 h-40 shrink-0 rounded shadow-sm object-cover bg-stone-100"
              onError={undefined}
            />
          ) : (
            <div className="w-32 h-40 shrink-0 bg-stone-100 rounded flex items-center justify-center text-3xl font-semibold text-stone-400">
              {author.name.charAt(0)}
            </div>
          )}

          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-3 mb-1">
              <h1 className="text-2xl font-bold text-stone-900">
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
              <p className="text-sm text-stone-500 mb-3">
                {author.birth_date ?? "?"}
                {" \u2013 "}
                {author.death_date ?? "present"}
              </p>
            )}

            <p className="text-xs text-stone-400 mb-4">
              {author.work_count} work{author.work_count === 1 ? "" : "s"}
            </p>

            {author.bio && (
              <p className="text-stone-700 text-sm leading-relaxed whitespace-pre-wrap">
                {author.bio}
              </p>
            )}

            {author.links && author.links.length > 0 && (
              <div className="flex flex-wrap gap-3 mt-4">
                {author.links.map((link) => (
                  <a
                    key={link.url}
                    href={link.url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-xs text-stone-500 underline hover:text-stone-700 transition-colors"
                  >
                    {link.title}
                  </a>
                ))}
              </div>
            )}
          </div>
        </div>

        {/* ── Works ── */}
        <section className="border-t border-stone-100 pt-8">
          <h2 className="text-sm font-semibold text-stone-500 uppercase tracking-wider mb-6">
            {works.length > 0
              ? `Works (${author.work_count})`
              : "Works"}
          </h2>

          {works.length === 0 ? (
            <p className="text-stone-400 text-sm">No works found.</p>
          ) : (
            <div className="grid grid-cols-4 sm:grid-cols-5 md:grid-cols-6 gap-4">
              {works.map((work) => (
                <Link
                  key={work.key}
                  href={`/books/${work.key}`}
                  className="group"
                >
                  {work.cover_url ? (
                    <img
                      src={work.cover_url}
                      alt={work.title}
                      className="w-full aspect-[2/3] object-cover rounded shadow-sm bg-stone-100 group-hover:shadow-md transition-shadow"
                    />
                  ) : (
                    <div className="w-full aspect-[2/3] bg-stone-100 rounded flex items-center justify-center p-2">
                      <span className="text-[10px] text-stone-400 text-center leading-tight line-clamp-3">
                        {work.title}
                      </span>
                    </div>
                  )}
                  <p className="mt-1.5 text-xs text-stone-700 leading-tight line-clamp-2 group-hover:text-stone-900 transition-colors">
                    {work.title}
                  </p>
                </Link>
              ))}
            </div>
          )}
        </section>
      </main>
    </div>
  );
}
