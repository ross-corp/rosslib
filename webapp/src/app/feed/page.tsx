import { redirect } from "next/navigation";
import Link from "next/link";
import { ActivityCard } from "@/components/activity";
import type { FeedResponse } from "@/components/activity";
import { getUser, getToken } from "@/lib/auth";

const FILTER_OPTIONS = [
  { label: "All", value: "" },
  { label: "Reviews", value: "reviewed" },
  { label: "Ratings", value: "rated" },
  { label: "Status Updates", value: "shelved,started_book,finished_book" },
  { label: "Threads", value: "created_thread" },
  { label: "Social", value: "followed_user,followed_author" },
] as const;

async function fetchFeed(
  token: string,
  cursor?: string,
  type?: string
): Promise<FeedResponse> {
  const url = new URL(`${process.env.API_URL}/me/feed`);
  if (cursor) url.searchParams.set("cursor", cursor);
  if (type) url.searchParams.set("type", type);

  const res = await fetch(url.toString(), {
    cache: "no-store",
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!res.ok) return { activities: [] };
  return res.json();
}

export default async function FeedPage({
  searchParams,
}: {
  searchParams: Promise<{ cursor?: string; type?: string }>;
}) {
  const [user, token] = await Promise.all([getUser(), getToken()]);

  if (!user || !token) redirect("/login");

  const { cursor, type } = await searchParams;
  const feed = await fetchFeed(token, cursor, type);

  return (
    <div className="min-h-screen">
      <main className="max-w-2xl mx-auto px-4 sm:px-6 py-12">
        <h1 className="text-2xl font-bold text-text-primary mb-8">Feed</h1>

        <div className="flex flex-wrap gap-2 mb-6">
          {FILTER_OPTIONS.map((opt) => {
            const isActive = (type || "") === opt.value;
            const href = opt.value
              ? `/feed?type=${encodeURIComponent(opt.value)}`
              : "/feed";
            return (
              <Link
                key={opt.value}
                href={href}
                className={`text-sm px-3 py-1.5 rounded-full border transition-colors ${
                  isActive
                    ? "bg-text-primary text-bg-primary border-text-primary"
                    : "bg-transparent text-text-secondary border-border hover:border-text-primary hover:text-text-primary"
                }`}
              >
                {opt.label}
              </Link>
            );
          })}
        </div>

        {feed.activities.length === 0 ? (
          <div className="text-center py-16">
            <p className="text-text-primary text-sm">
              {type
                ? "No activity matches this filter."
                : "No activity yet. Follow some people to see their updates here."}
            </p>
            {!type && (
              <Link
                href="/users"
                className="inline-block mt-4 text-sm text-text-primary hover:text-text-primary border border-border px-4 py-2 rounded transition-colors"
              >
                Browse people
              </Link>
            )}
          </div>
        ) : (
          <>
            <div>
              {feed.activities.map((item) => (
                <ActivityCard key={item.id} item={item} />
              ))}
            </div>

            {feed.next_cursor && (
              <div className="mt-8 text-center">
                <Link
                  href={`/feed?${type ? `type=${encodeURIComponent(type)}&` : ""}cursor=${encodeURIComponent(feed.next_cursor)}`}
                  className="text-sm text-text-primary hover:text-text-primary border border-border px-4 py-2 rounded transition-colors"
                >
                  Load more
                </Link>
              </div>
            )}
          </>
        )}
      </main>
    </div>
  );
}
