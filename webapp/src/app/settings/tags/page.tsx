import { redirect } from "next/navigation";
import Link from "next/link";
import Nav from "@/components/nav";
import TagSettingsForm from "@/components/tag-settings-form";
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
      <Nav />
      <main className="max-w-2xl mx-auto px-4 sm:px-6 py-10">
        <div className="mb-8 flex items-center gap-2 text-sm text-stone-400">
          <Link href="/settings" className="hover:text-stone-700 transition-colors">
            Settings
          </Link>
          <span>/</span>
          <span className="text-stone-600">Labels</span>
        </div>

        <h1 className="text-2xl font-bold text-stone-900 mb-2">Label categories</h1>
        <p className="text-sm text-stone-500 mb-8">
          Define categories and their allowed values. Assign one value (or multiple, depending on mode) to any book on your shelves.
        </p>

        <TagSettingsForm initialTagKeys={tagKeys} />
      </main>
    </div>
  );
}
