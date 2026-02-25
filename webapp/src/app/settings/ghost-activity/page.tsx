import { redirect } from "next/navigation";
import Link from "next/link";
import GhostControls from "./ghost-controls";
import { getUser, getToken } from "@/lib/auth";

type GhostStatus = {
  username: string;
  display_name: string;
  user_id: string;
  books_read: number;
  currently_reading: number;
  want_to_read: number;
  following_count: number;
  followers_count: number;
};

async function fetchGhostStatus(token: string): Promise<GhostStatus[]> {
  const res = await fetch(`${process.env.API_URL}/admin/ghosts/status`, {
    cache: "no-store",
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!res.ok) return [];
  const data = await res.json();
  return Array.isArray(data) ? data : [];
}

export default async function GhostActivityPage() {
  const user = await getUser();
  if (!user) redirect("/login");

  const token = await getToken();
  const ghosts = await fetchGhostStatus(token!);

  return (
    <div className="min-h-screen">
      <main className="max-w-5xl mx-auto px-4 sm:px-6 py-10">
        <div className="mb-8 flex items-center gap-2 text-sm text-text-primary">
          <Link href={`/${user.username}`} className="hover:text-text-primary transition-colors">
            {user.username}
          </Link>
          <span>/</span>
          <Link href="/settings" className="hover:text-text-primary transition-colors">
            Settings
          </Link>
          <span>/</span>
          <span className="text-text-primary">Ghost activity</span>
        </div>

        <h1 className="text-2xl font-bold text-text-primary mb-8">Ghost activity</h1>

        <GhostControls initialGhosts={ghosts} />
      </main>
    </div>
  );
}
