import { redirect } from "next/navigation";
import Link from "next/link";
import { getUser, getToken } from "@/lib/auth";
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
  const json = await res.json();
  // API returns a plain array; normalize to expected shape
  if (Array.isArray(json)) {
    return { notifications: json };
  }
  return { notifications: json.notifications ?? [], next_cursor: json.next_cursor };
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
          <div className="flex items-center gap-4">
            <h1 className="text-2xl font-bold text-text-primary">Notifications</h1>
            <Link
              href="/recommendations"
              className="text-xs text-text-secondary hover:text-text-primary border border-border rounded px-2.5 py-1.5 transition-colors"
            >
              Recommendations
            </Link>
          </div>
        </div>

        <NotificationList
          notifications={data.notifications}
          nextCursor={data.next_cursor}
        />
      </main>
    </div>
  );
}
