import { redirect } from "next/navigation";
import Link from "next/link";
import Nav from "@/components/nav";
import FollowRequestsList from "./follow-requests-list";
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
      <Nav />
      <main className="max-w-5xl mx-auto px-4 sm:px-6 py-10">
        <div className="mb-8 flex items-center gap-2 text-sm text-stone-400">
          <Link href={`/${user.username}`} className="hover:text-stone-700 transition-colors">
            {user.username}
          </Link>
          <span>/</span>
          <Link href="/settings" className="hover:text-stone-700 transition-colors">
            Settings
          </Link>
          <span>/</span>
          <span className="text-stone-600">Follow requests</span>
        </div>

        <h1 className="text-2xl font-bold text-stone-900 mb-8">Follow requests</h1>

        <FollowRequestsList initialRequests={requests} />
      </main>
    </div>
  );
}
