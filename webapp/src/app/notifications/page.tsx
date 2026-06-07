import type { Metadata } from "next";
import { redirect } from "next/navigation";
import Link from "next/link";
import { getUser, getToken } from "@/lib/auth";
import NotificationList from "./notification-list";
import EmptyState from "@/components/empty-state";

export const metadata: Metadata = {
  title: "Notifications",
};

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
  total: number;
  page: number;
  perPage: number;
};

async function fetchNotifications(
  token: string,
  page?: number
): Promise<NotificationsResponse> {
  const url = new URL(`${process.env.API_URL}/me/notifications`);
  if (page) url.searchParams.set("page", String(page));

  const res = await fetch(url.toString(), {
    cache: "no-store",
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!res.ok) return { notifications: [], total: 0, page: 1, perPage: 20 };
  const json = await res.json();
  return {
    notifications: json.notifications ?? [],
    total: json.total ?? 0,
    page: json.page ?? 1,
    perPage: json.perPage ?? 20,
  };
}

export default async function NotificationsPage({
  searchParams,
}: {
  searchParams: Promise<{ page?: string }>;
}) {
  const [user, token] = await Promise.all([getUser(), getToken()]);

  if (!user || !token) redirect("/login");

  const { page: pageParam } = await searchParams;
  const page = pageParam ? parseInt(pageParam, 10) : 1;
  const data = await fetchNotifications(token, page);
  const totalPages = Math.ceil(data.total / data.perPage);

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

        {data.notifications.length === 0 ? (
          <EmptyState
            message="No notifications. You're all caught up!"
            actionLabel="Go to feed"
            actionHref="/feed"
          />
        ) : (
          <NotificationList
            notifications={data.notifications}
            page={data.page}
            totalPages={totalPages}
          />
        )}
      </main>
    </div>
  );
}
