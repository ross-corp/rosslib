import { redirect } from "next/navigation";
import Link from "next/link";
import BlockedUsersList from "./blocked-users-list";
import SettingsNav from "@/components/settings-nav";
import { getUser, getToken } from "@/lib/auth";

type BlockedUser = {
  id: string;
  username: string;
  display_name: string | null;
  avatar_url: string | null;
  blocked_at: string;
};

type BlockedUsersResponse = {
  users: BlockedUser[];
  total: number;
  page: number;
  perPage: number;
};

async function fetchBlockedUsers(token: string): Promise<BlockedUsersResponse> {
  const res = await fetch(`${process.env.API_URL}/me/blocks?page=1&perPage=20`, {
    cache: "no-store",
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!res.ok) return { users: [], total: 0, page: 1, perPage: 20 };
  return res.json();
}

export default async function BlockedUsersPage() {
  const user = await getUser();
  if (!user) redirect("/login");

  const token = await getToken();
  const data = await fetchBlockedUsers(token!);

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

        <h2 className="text-xl font-semibold text-text-primary mb-8">Blocked users</h2>

        <BlockedUsersList initialUsers={data.users} total={data.total} />
      </main>
    </div>
  );
}
