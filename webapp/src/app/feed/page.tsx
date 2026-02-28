import { redirect } from "next/navigation";
import Link from "next/link";
import { ActivityCard } from "@/components/activity";
import type { FeedResponse } from "@/components/activity";
import { getUser, getToken } from "@/lib/auth";
import EmptyState from "@/components/empty-state";

async function fetchFeed(
  token: string,
  cursor?: string
): Promise<FeedResponse> {
  const url = new URL(`${process.env.API_URL}/me/feed`);
  if (cursor) url.searchParams.set("cursor", cursor);

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
  searchParams: Promise<{ cursor?: string }>;
}) {
  const [user, token] = await Promise.all([getUser(), getToken()]);

  if (!user || !token) redirect("/login");

  const { cursor } = await searchParams;
  const feed = await fetchFeed(token, cursor);

  return (
    <div className="min-h-screen">
      <main className="max-w-2xl mx-auto px-4 sm:px-6 py-12">
        <h1 className="text-2xl font-bold text-text-primary mb-8">Feed</h1>

        {feed.activities.length === 0 ? (
          <EmptyState
            message="Your feed is empty. Follow some readers to see their activity."
            actionLabel="Browse people"
            actionHref="/users"
          />
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
                  href={`/feed?cursor=${encodeURIComponent(feed.next_cursor)}`}
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
