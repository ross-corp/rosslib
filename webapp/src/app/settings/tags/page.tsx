import { redirect } from "next/navigation";
import Link from "next/link";
import TagSettingsForm from "@/components/tag-settings-form";
import SettingsNav from "@/components/settings-nav";
import { getUser, getToken } from "@/lib/auth";

type TagValue = {
  id: string;
  name: string;
  slug: string;
};

type TagKey = {
  id: string;
  name: string;
  slug: string;
  mode: "select_one" | "select_multiple";
  values: TagValue[];
};

async function fetchTagKeys(token: string): Promise<TagKey[]> {
  const res = await fetch(`${process.env.API_URL}/me/tag-keys`, {
    headers: { Authorization: `Bearer ${token}` },
    cache: "no-store",
  });
  if (!res.ok) return [];
  return res.json();
}

export default async function TagSettingsPage() {
  const [user, token] = await Promise.all([getUser(), getToken()]);
  if (!user || !token) redirect("/login");

  const tagKeys = await fetchTagKeys(token);

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

        <h2 className="text-xl font-semibold text-text-primary mb-2">Label categories</h2>
        <p className="text-sm text-text-primary mb-8">
          Define categories and their allowed values. Assign one value (or multiple, depending on mode) to any book on your shelves.
        </p>

        <TagSettingsForm initialTagKeys={tagKeys} />
      </main>
    </div>
  );
}
