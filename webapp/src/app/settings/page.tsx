import { redirect } from "next/navigation";
import Link from "next/link";
import DeleteDataForm from "@/components/delete-data-form";
import EmailVerificationBanner from "@/components/email-verification-banner";
import NotificationPreferences from "@/components/notification-preferences";
import PasswordForm from "@/components/password-form";
import ReadingGoalForm from "@/components/reading-goal-form";
import SettingsForm from "@/components/settings-form";
import SettingsNav from "@/components/settings-nav";
import { getUser, getToken } from "@/lib/auth";

type UserProfile = {
  display_name: string | null;
  bio: string | null;
  avatar_url: string | null;
  banner_url: string | null;
  is_private: boolean;
};

type GoalData = {
  id: string;
  year: number;
  target: number;
};

async function fetchProfile(username: string): Promise<UserProfile | null> {
  const res = await fetch(`${process.env.API_URL}/users/${username}`, {
    cache: "no-store",
  });
  if (!res.ok) return null;
  return res.json();
}

async function fetchCurrentGoal(token: string): Promise<GoalData | null> {
  const year = new Date().getFullYear();
  const res = await fetch(`${process.env.API_URL}/me/goals/${year}`, {
    headers: { Authorization: `Bearer ${token}` },
    cache: "no-store",
  });
  if (!res.ok) return null;
  return res.json();
}

export default async function SettingsPage() {
  const [user, token] = await Promise.all([getUser(), getToken()]);
  if (!user || !token) redirect("/login");

  const [profile, currentGoal] = await Promise.all([
    fetchProfile(user.username),
    fetchCurrentGoal(token),
  ]);

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

        <EmailVerificationBanner />

        <SettingsForm
          username={user.username}
          initialDisplayName={profile?.display_name ?? ""}
          initialBio={profile?.bio ?? ""}
          initialAvatarUrl={profile?.avatar_url ?? null}
          initialBannerUrl={profile?.banner_url ?? null}
          initialIsPrivate={profile?.is_private ?? false}
        />

        <NotificationPreferences />

        <ReadingGoalForm initialGoal={currentGoal} />

        <PasswordForm />

        <DeleteDataForm />
      </main>
    </div>
  );
}
