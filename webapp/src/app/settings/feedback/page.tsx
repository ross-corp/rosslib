import { redirect } from "next/navigation";
import Link from "next/link";
import FeedbackList from "./feedback-list";
import SettingsNav from "@/components/settings-nav";
import { getUser, getToken } from "@/lib/auth";

type FeedbackItem = {
  id: string;
  type: string;
  title: string;
  description: string;
  steps_to_reproduce: string | null;
  severity: string | null;
  status: string;
  created_at: string;
};

async function fetchFeedback(token: string): Promise<FeedbackItem[]> {
  const res = await fetch(`${process.env.API_URL}/me/feedback`, {
    cache: "no-store",
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!res.ok) return [];
  return res.json();
}

export default async function FeedbackPage() {
  const user = await getUser();
  if (!user) redirect("/login");

  const token = await getToken();
  const items = await fetchFeedback(token!);

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

        <h2 className="text-xl font-semibold text-text-primary mb-2">My feedback</h2>
        <p className="text-sm text-text-primary mb-8">
          Bug reports and feature requests you&apos;ve submitted.
        </p>

        <FeedbackList initialItems={items} />
      </main>
    </div>
  );
}
