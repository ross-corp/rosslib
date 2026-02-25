import { redirect } from "next/navigation";
import Link from "next/link";
import Nav from "@/components/nav";
import EmailVerificationBanner from "@/components/email-verification-banner";
import PasswordForm from "@/components/password-form";
import SettingsForm from "@/components/settings-form";
import { getUser } from "@/lib/auth";

type UserProfile = {
  display_name: string | null;
  bio: string | null;
  avatar_url: string | null;
  is_private: boolean;
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

        <div className="flex items-center justify-between mb-8">
          <h1 className="text-2xl font-bold text-stone-900">Profile</h1>
          <div className="flex items-center gap-4">
            <Link
              href="/settings/follow-requests"
              className="text-sm text-stone-500 hover:text-stone-900 transition-colors"
            >
              Follow requests
            </Link>
            <Link
              href="/settings/tags"
              className="text-sm text-stone-500 hover:text-stone-900 transition-colors"
            >
              Tag categories
            </Link>
            <Link
              href="/scan"
              className="text-sm text-stone-500 hover:text-stone-900 transition-colors"
            >
              Scan ISBN
            </Link>
            <Link
              href="/settings/import"
              className="text-sm text-stone-500 hover:text-stone-900 transition-colors"
            >
              Import from Goodreads
            </Link>
            <Link
              href="/settings/export"
              className="text-sm text-stone-500 hover:text-stone-900 transition-colors"
            >
              Export to CSV
            </Link>
            <Link
              href="/library/compare"
              className="text-sm text-stone-500 hover:text-stone-900 transition-colors"
            >
              Compare lists
            </Link>
            <Link
              href="/settings/ghost-activity"
              className="text-sm text-stone-500 hover:text-stone-900 transition-colors"
            >
              Ghost activity
            </Link>
          </div>
        </div>

        <EmailVerificationBanner />

        <SettingsForm
          username={user.username}
          initialDisplayName={profile?.display_name ?? ""}
          initialBio={profile?.bio ?? ""}
          initialAvatarUrl={profile?.avatar_url ?? null}
          initialIsPrivate={profile?.is_private ?? false}
        />

        <PasswordForm />
      </main>
    </div>
  );
}
