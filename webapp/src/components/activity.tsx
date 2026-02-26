import Link from "next/link";

export type ActivityUser = {
  user_id: string;
  username: string;
  display_name: string | null;
  avatar_url: string | null;
};

export type ActivityBook = {
  open_library_id: string;
  title: string;
  cover_url: string | null;
};

export type ActivityItem = {
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
  link_type?: string;
  to_book_ol_id?: string;
  to_book_title?: string;
  author_key?: string;
  author_name?: string;
};

export type FeedResponse = {
  activities: ActivityItem[];
  next_cursor?: string;
};

export function formatTime(iso: string): string {
  if (!iso) return "";
  const d = new Date(iso);
  if (isNaN(d.getTime())) return "";
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

export function Stars({ rating }: { rating: number }) {
  return (
    <span className="text-amber-500 text-sm tracking-tight">
      {"★".repeat(rating)}
      {"☆".repeat(5 - rating)}
    </span>
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
              className="font-medium text-text-primary hover:underline"
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
              className="font-medium text-text-primary hover:underline"
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
              className="font-medium text-text-primary hover:underline"
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
                className="font-medium text-text-primary hover:underline"
              >
                {item.book.title}
              </Link>
            </>
          )}
        </>
      );
    case "started_book":
      return (
        <>
          started reading{" "}
          {item.book && (
            <Link
              href={`/books/${item.book.open_library_id}`}
              className="font-medium text-text-primary hover:underline"
            >
              {item.book.title}
            </Link>
          )}
        </>
      );
    case "finished_book":
      return (
        <>
          finished{" "}
          {item.book && (
            <Link
              href={`/books/${item.book.open_library_id}`}
              className="font-medium text-text-primary hover:underline"
            >
              {item.book.title}
            </Link>
          )}
        </>
      );
    case "follow":
    case "followed_user":
      return (
        <>
          followed{" "}
          {item.target_user && (
            <Link
              href={`/${item.target_user.username}`}
              className="font-medium text-text-primary hover:underline"
            >
              {item.target_user.display_name || item.target_user.username}
            </Link>
          )}
        </>
      );
    case "followed_author":
      return (
        <>
          followed author{" "}
          {item.author_key ? (
            <Link
              href={`/authors/${item.author_key}`}
              className="font-medium text-text-primary hover:underline"
            >
              {item.author_name || item.author_key}
            </Link>
          ) : (
            "an author"
          )}
        </>
      );
    case "followed_book":
      return (
        <>
          followed{" "}
          {item.book ? (
            <Link
              href={`/books/${item.book.open_library_id}`}
              className="font-medium text-text-primary hover:underline"
            >
              {item.book.title}
            </Link>
          ) : (
            "a book"
          )}
        </>
      );
    case "created_link":
      return (
        <>
          submitted a{" "}
          {item.link_type ? item.link_type.replace("_", " ") : "link"} link on{" "}
          {item.book && (
            <Link
              href={`/books/${item.book.open_library_id}`}
              className="font-medium text-text-primary hover:underline"
            >
              {item.book.title}
            </Link>
          )}
          {item.to_book_ol_id && item.to_book_title && (
            <>
              {" "}
              to{" "}
              <Link
                href={`/books/${item.to_book_ol_id}`}
                className="font-medium text-text-primary hover:underline"
              >
                {item.to_book_title}
              </Link>
            </>
          )}
        </>
      );
    default:
      return <>{item.type}</>;
  }
}

export function ActivityCard({
  item,
  showUser = true,
}: {
  item: ActivityItem;
  showUser?: boolean;
}) {
  const displayName = item.user.display_name || item.user.username;

  return (
    <div className="flex gap-3 py-4 border-b border-border last:border-0">
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
          <div className="w-10 h-10 rounded-full bg-surface-2 flex items-center justify-center text-text-tertiary text-xs font-medium">
            {displayName.charAt(0).toUpperCase()}
          </div>
        )}
      </div>

      <div className="flex-1 min-w-0">
        {/* Action description */}
        <p className="text-sm text-text-secondary leading-snug">
          {showUser && (
            <>
              <Link
                href={`/${item.user.username}`}
                className="font-semibold text-text-primary hover:underline"
              >
                {displayName}
              </Link>{" "}
            </>
          )}
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
          <p className="mt-1 text-sm text-text-secondary line-clamp-2">
            {item.review_snippet}
          </p>
        )}

        {/* Timestamp */}
        <p className="text-xs text-text-tertiary mt-1">
          {formatTime(item.created_at)}
        </p>
      </div>
    </div>
  );
}
