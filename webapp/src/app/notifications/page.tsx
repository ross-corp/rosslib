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

function NotificationIcon({ type }: { type: string }) {
  switch (type) {
    case "new_publication":
      // Book/pages icon
      return (
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" className="w-5 h-5">
          <path d="M10.75 16.82A7.462 7.462 0 0115 15.5c.71 0 1.396.098 2.046.282A.75.75 0 0018 15.06v-11a.75.75 0 00-.546-.721A9.006 9.006 0 0015 3a8.963 8.963 0 00-4.25 1.065V16.82zM9.25 4.065A8.963 8.963 0 005 3c-.85 0-1.673.118-2.454.339A.75.75 0 002 4.06v11a.75.75 0 00.954.721A7.462 7.462 0 015 15.5c1.579 0 3.042.487 4.25 1.32V4.065z" />
        </svg>
      );
    case "book_new_thread":
      // Chat bubble icon
      return (
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" className="w-5 h-5">
          <path fillRule="evenodd" d="M10 2c-2.236 0-4.43.18-6.57.524C1.993 2.755 1 3.97 1 5.396v5.21c0 1.425.993 2.64 2.43 2.872 1.202.194 2.426.325 3.668.39.455.025.875.28 1.098.671L10 17.5l1.804-2.96c.223-.392.643-.647 1.098-.672a38.447 38.447 0 003.668-.39C18.007 13.246 19 12.031 19 10.605V5.397c0-1.426-.993-2.64-2.43-2.873A39.82 39.82 0 0010 2z" clipRule="evenodd" />
        </svg>
      );
    case "book_new_link":
      // Link icon
      return (
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" className="w-5 h-5">
          <path d="M12.232 4.232a2.5 2.5 0 013.536 3.536l-1.225 1.224a.75.75 0 001.061 1.06l1.224-1.224a4 4 0 00-5.656-5.656l-3 3a4 4 0 00.225 5.865.75.75 0 00.977-1.138 2.5 2.5 0 01-.142-3.667l3-3z" />
          <path d="M11.603 7.963a.75.75 0 00-.977 1.138 2.5 2.5 0 01.142 3.667l-3 3a2.5 2.5 0 01-3.536-3.536l1.225-1.224a.75.75 0 00-1.061-1.06l-1.224 1.224a4 4 0 005.656 5.656l3-3a4 4 0 00-.225-5.865z" />
        </svg>
      );
    case "book_new_review":
      // Star icon
      return (
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" className="w-5 h-5">
          <path fillRule="evenodd" d="M10.868 2.884c-.321-.772-1.415-.772-1.736 0l-1.83 4.401-4.753.381c-.833.067-1.171 1.107-.536 1.651l3.62 3.102-1.106 4.637c-.194.813.691 1.456 1.405 1.02L10 15.591l4.069 2.485c.713.436 1.598-.207 1.404-1.02l-1.106-4.637 3.62-3.102c.635-.544.297-1.584-.536-1.65l-4.752-.382-1.831-4.401z" clipRule="evenodd" />
        </svg>
      );
    default:
      // Bell icon (generic fallback)
      return (
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" className="w-5 h-5">
          <path fillRule="evenodd" d="M10 2a6 6 0 00-6 6c0 1.887-.454 3.665-1.257 5.234a.75.75 0 00.515 1.076 32.91 32.91 0 003.256.508.75.75 0 00.742.59h5.488a.75.75 0 00.742-.59 32.91 32.91 0 003.256-.508.75.75 0 00.515-1.076A11.448 11.448 0 0116 8a6 6 0 00-6-6zM8.05 14.943a33.54 33.54 0 003.9 0 2 2 0 01-3.9 0z" clipRule="evenodd" />
        </svg>
      );
  }
}

function NotificationCard({ notif }: { notif: Notification }) {
  const authorKey = notif.metadata?.author_key;
  const bookOlId = notif.metadata?.book_ol_id;

  return (
    <div
      className={`py-4 flex gap-3 ${notif.read ? "opacity-60" : ""}`}
    >
      <div className="shrink-0 w-10 h-10 rounded-full bg-surface-2 flex items-center justify-center text-text-primary">
        <NotificationIcon type={notif.notif_type} />
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
