import { redirect } from "next/navigation";
import Link from "next/link";
import { getUser, getToken } from "@/lib/auth";
import NotificationList from "./notification-list";
import NotificationCard from "./notification-card";

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

