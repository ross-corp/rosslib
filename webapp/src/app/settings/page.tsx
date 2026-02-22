import { redirect } from "next/navigation";
import Link from "next/link";
import Nav from "@/components/nav";
import SettingsForm from "@/components/settings-form";
import { getUser } from "@/lib/auth";

type UserProfile = {
  display_name: string | null;
  bio: string | null;
};

async function fetchProfile(username: string): Promise<UserProfile | null> {
  const res = await fetch(`${process.env.API_URL}/users/${username}`, {
    cache: "no-store",
  });
  if (!res.ok) return null;
  return res.json();
}

export default async function SettingsPage() {
  const user = await getUser();
  if (!user) redirect("/login");

  const profile = await fetchProfile(user.username);

  return (
    <div className="min-h-screen">
      <Nav />
      <main className="max-w-5xl mx-auto px-4 sm:px-6 py-10">
        <div className="mb-8 flex items-center gap-2 text-sm text-stone-400">
          <Link href={`/${user.username}`} className="hover:text-stone-700 transition-colors">
            {user.username}
          </Link>
          <span>/</span>
          <span className="text-stone-600">Settings</span>
        </div>

        <h1 className="text-2xl font-bold text-stone-900 mb-8">Profile</h1>

        <SettingsForm
          username={user.username}
          initialDisplayName={profile?.display_name ?? ""}
          initialBio={profile?.bio ?? ""}
        />
      </main>
    </div>
  );
}
