import { redirect } from "next/navigation";
import Link from "next/link";
import { getUser, getToken } from "@/lib/auth";
import { formatTime } from "@/components/activity";
import NotificationList from "./notification-list";

type Notification = {
  id: string;
  notif_type: string;
  title: string;
  body: string | null;
  metadata: Record<string, string> | null;
  read: boolean;
  created_at: string;
};

type NotificationsResponse = {
  notifications: Notification[];
  next_cursor?: string;
};

async function fetchNotifications(
  token: string,
  cursor?: string
): Promise<NotificationsResponse> {
  const url = new URL(`${process.env.API_URL}/me/notifications`);
  if (cursor) url.searchParams.set("cursor", cursor);

  const res = await fetch(url.toString(), {
    cache: "no-store",
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!res.ok) return { notifications: [] };
  return res.json();
}

export default async function NotificationsPage({
  searchParams,
}: {
  searchParams: Promise<{ cursor?: string }>;
}) {
  const [user, token] = await Promise.all([getUser(), getToken()]);

  if (!user || !token) redirect("/login");

  const { cursor } = await searchParams;
  const data = await fetchNotifications(token, cursor);

  return (
    <div className="min-h-screen">
      <main className="max-w-2xl mx-auto px-4 sm:px-6 py-12">
        <div className="flex items-center justify-between mb-8">
          <h1 className="text-2xl font-bold text-text-primary">Notifications</h1>
          {data.notifications.length > 0 && (
            <NotificationList
              notifications={data.notifications}
              nextCursor={data.next_cursor}
            />
          )}
        </div>

        {data.notifications.length === 0 ? (
          <div className="text-center py-16">
            <p className="text-text-primary text-sm">
              No notifications yet. Follow authors to get notified about
              new publications, or follow books to hear about new discussions,
              links, and reviews.
            </p>
            <Link
              href="/feed"
              className="inline-block mt-4 text-sm text-text-primary hover:text-text-primary border border-border px-4 py-2 rounded transition-colors"
            >
              Go to feed
            </Link>
          </div>
        ) : (
          <>
            <div className="divide-y divide-border">
              {data.notifications.map((notif) => (
                <NotificationCard key={notif.id} notif={notif} />
              ))}
            </div>

            {data.next_cursor && (
              <div className="mt-8 text-center">
                <Link
                  href={`/notifications?cursor=${encodeURIComponent(data.next_cursor)}`}
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

function NotificationCard({ notif }: { notif: Notification }) {
  const authorKey = notif.metadata?.author_key;
  const bookOlId = notif.metadata?.book_ol_id;

  return (
    <div
      className={`py-4 flex gap-3 ${notif.read ? "opacity-60" : ""}`}
    >
      <div className="shrink-0 w-10 h-10 rounded-full bg-surface-2 flex items-center justify-center text-text-primary">
        <svg
          xmlns="http://www.w3.org/2000/svg"
          viewBox="0 0 20 20"
          fill="currentColor"
          className="w-5 h-5"
        >
          <path d="M10 1a.75.75 0 01.75.75v1.5a.75.75 0 01-1.5 0v-1.5A.75.75 0 0110 1zM5.05 3.05a.75.75 0 011.06 0l1.062 1.06A.75.75 0 116.11 5.173L5.05 4.11a.75.75 0 010-1.06zm9.9 0a.75.75 0 010 1.06l-1.06 1.062a.75.75 0 01-1.062-1.061l1.061-1.06a.75.75 0 011.06 0zM10 7a3 3 0 100 6 3 3 0 000-6zm-6.25 3a.75.75 0 01.75-.75h1.5a.75.75 0 010 1.5H4.5a.75.75 0 01-.75-.75zm12 0a.75.75 0 01.75-.75h1.5a.75.75 0 010 1.5h-1.5a.75.75 0 01-.75-.75zM5.05 16.95a.75.75 0 011.06-1.06l1.062 1.06a.75.75 0 11-1.061 1.062L5.05 16.95zm9.9 0a.75.75 0 01-1.06 0l-1.062-1.06a.75.75 0 111.061-1.062l1.06 1.061a.75.75 0 010 1.06zM10 14.5a.75.75 0 01.75.75v1.5a.75.75 0 01-1.5 0v-1.5a.75.75 0 01.75-.75z" />
        </svg>
      </div>
      <div className="flex-1 min-w-0">
        <p className="text-sm font-medium text-text-primary">{notif.title}</p>
        {notif.body && (
          <p className="text-sm text-text-primary mt-0.5">{notif.body}</p>
        )}
        <div className="flex items-center gap-3 mt-1">
          <span className="text-xs text-text-primary">
            {formatTime(notif.created_at)}
          </span>
          {authorKey && (
            <Link
              href={`/authors/${authorKey}`}
              className="text-xs text-text-primary hover:text-text-primary hover:underline"
            >
              View author
            </Link>
          )}
          {bookOlId && (
            <Link
              href={`/books/${bookOlId}`}
              className="text-xs text-text-primary hover:text-text-primary hover:underline"
            >
              View book
            </Link>
          )}
        </div>
      </div>
      {!notif.read && (
        <div className="shrink-0 mt-1.5">
          <div className="w-2 h-2 rounded-full bg-blue-500" />
        </div>
      )}
    </div>
  );
}
