import { redirect } from "next/navigation";
import Link from "next/link";
import Nav from "@/components/nav";
import { getUser, getToken } from "@/lib/auth";

type ActivityUser = {
  user_id: string;
  username: string;
  display_name: string | null;
  avatar_url: string | null;
};

type ActivityBook = {
  open_library_id: string;
  title: string;
  cover_url: string | null;
};

type ActivityItem = {
  id: string;
  type: string;
  created_at: string;
  user: ActivityUser;
  book?: ActivityBook;
  target_user?: ActivityUser;
  shelf_name?: string;
  rating?: number;
  review_snippet?: string;
  thread_title?: string;
};

type FeedResponse = {
  activities: ActivityItem[];
  next_cursor?: string;
};

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

function formatTime(iso: string): string {
  const d = new Date(iso);
  const now = new Date();
  const diff = now.getTime() - d.getTime();
  const mins = Math.floor(diff / 60000);
  if (mins < 1) return "just now";
  if (mins < 60) return `${mins}m ago`;
  const hours = Math.floor(mins / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  if (days < 7) return `${days}d ago`;
  return d.toLocaleDateString("en-US", { month: "short", day: "numeric" });
}

function Stars({ rating }: { rating: number }) {
  return (
    <span className="text-amber-500 text-sm tracking-tight">
      {"★".repeat(rating)}
      {"☆".repeat(5 - rating)}
    </span>
  );
}

function ActivityCard({ item }: { item: ActivityItem }) {
  const displayName = item.user.display_name || item.user.username;

  return (
    <div className="flex gap-3 py-4 border-b border-stone-100 last:border-0">
      {/* Book cover or avatar */}
      <div className="shrink-0 w-10">
        {item.book?.cover_url ? (
          <Link href={`/books/${item.book.open_library_id}`}>
            <img
              src={item.book.cover_url}
              alt=""
              className="w-10 h-14 object-cover rounded shadow-sm"
            />
          </Link>
        ) : (
          <div className="w-10 h-10 rounded-full bg-stone-200 flex items-center justify-center text-stone-500 text-xs font-medium">
            {displayName.charAt(0).toUpperCase()}
          </div>
        )}
      </div>

      <div className="flex-1 min-w-0">
        {/* Action description */}
        <p className="text-sm text-stone-700 leading-snug">
          <Link
            href={`/${item.user.username}`}
            className="font-semibold text-stone-900 hover:underline"
          >
            {displayName}
          </Link>{" "}
          <ActivityDescription item={item} />
        </p>

        {/* Rating */}
        {item.rating && item.type === "rated" && (
          <div className="mt-1">
            <Stars rating={item.rating} />
          </div>
        )}

        {/* Review snippet */}
        {item.review_snippet && (
          <p className="mt-1 text-sm text-stone-500 line-clamp-2">
            {item.review_snippet}
          </p>
        )}

        {/* Timestamp */}
        <p className="text-xs text-stone-400 mt-1">
          {formatTime(item.created_at)}
        </p>
      </div>
    </div>
  );
}

function ActivityDescription({ item }: { item: ActivityItem }) {
  switch (item.type) {
    case "shelved":
      return (
        <>
          added{" "}
          {item.book && (
            <Link
              href={`/books/${item.book.open_library_id}`}
              className="font-medium text-stone-900 hover:underline"
            >
              {item.book.title}
            </Link>
          )}{" "}
          to {item.shelf_name || "a shelf"}
        </>
      );
    case "rated":
      return (
        <>
          rated{" "}
          {item.book && (
            <Link
              href={`/books/${item.book.open_library_id}`}
              className="font-medium text-stone-900 hover:underline"
            >
              {item.book.title}
            </Link>
          )}
          {item.rating && (
            <>
              {" "}
              <Stars rating={item.rating} />
            </>
          )}
        </>
      );
    case "reviewed":
      return (
        <>
          reviewed{" "}
          {item.book && (
            <Link
              href={`/books/${item.book.open_library_id}`}
              className="font-medium text-stone-900 hover:underline"
            >
              {item.book.title}
            </Link>
          )}
        </>
      );
    case "created_thread":
      return (
        <>
          started a discussion
          {item.thread_title && (
            <>
              {" "}
              &ldquo;{item.thread_title}&rdquo;
            </>
          )}
          {item.book && (
            <>
              {" "}
              on{" "}
              <Link
                href={`/books/${item.book.open_library_id}`}
                className="font-medium text-stone-900 hover:underline"
              >
                {item.book.title}
              </Link>
            </>
          )}
        </>
      );
    case "followed_user":
      return (
        <>
          followed{" "}
          {item.target_user && (
            <Link
              href={`/${item.target_user.username}`}
              className="font-medium text-stone-900 hover:underline"
            >
              {item.target_user.display_name || item.target_user.username}
            </Link>
          )}
        </>
      );
    default:
      return <>{item.type}</>;
  }
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
      <Nav />
      <main className="max-w-2xl mx-auto px-4 sm:px-6 py-12">
        <h1 className="text-2xl font-bold text-stone-900 mb-8">Feed</h1>

        {feed.activities.length === 0 ? (
          <div className="text-center py-16">
            <p className="text-stone-500 text-sm">
              No activity yet. Follow some people to see their updates here.
            </p>
            <Link
              href="/users"
              className="inline-block mt-4 text-sm text-stone-600 hover:text-stone-900 border border-stone-300 px-4 py-2 rounded transition-colors"
            >
              Browse people
            </Link>
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
                  href={`/feed?cursor=${encodeURIComponent(feed.next_cursor)}`}
                  className="text-sm text-stone-500 hover:text-stone-900 border border-stone-300 px-4 py-2 rounded transition-colors"
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
