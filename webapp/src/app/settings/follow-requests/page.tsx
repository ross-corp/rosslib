import { redirect } from "next/navigation";
import Link from "next/link";
import FollowRequestsList from "./follow-requests-list";
import SettingsNav from "@/components/settings-nav";
import { getUser, getToken } from "@/lib/auth";

type FollowRequest = {
  user_id: string;
  username: string;
  display_name: string | null;
  avatar_url: string | null;
  created_at: string;
};

async function fetchFollowRequests(token: string): Promise<FollowRequest[]> {
  const res = await fetch(`${process.env.API_URL}/me/follow-requests`, {
    cache: "no-store",
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!res.ok) return [];
  return res.json();
}

export default async function FollowRequestsPage() {
  const user = await getUser();
  if (!user) redirect("/login");

  const token = await getToken();
  const requests = await fetchFollowRequests(token!);

  return (
    <div className="min-h-screen">
      <main className="max-w-5xl mx-auto px-4 sm:px-6 py-10">
        <div className="mb-8 flex items-center gap-2 text-sm text-text-primary">
          <Link href={`/${user.username}`} className="hover:text-text-primary transition-colors">
            {user.username}
          </Link>
          <span>/</span>
          <span className="text-text-primary">Settings</span>
        </div>

        <h1 className="text-2xl font-bold text-text-primary mb-4">Settings</h1>

        <SettingsNav />

        <h2 className="text-xl font-semibold text-text-primary mb-8">Follow requests</h2>

        <FollowRequestsList initialRequests={requests} />
      </main>
    </div>
  );
}
