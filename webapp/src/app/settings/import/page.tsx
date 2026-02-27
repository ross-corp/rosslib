import { redirect } from "next/navigation";
import Link from "next/link";
import ImportTabs from "@/components/import-tabs";
import SettingsNav from "@/components/settings-nav";
import { getUser } from "@/lib/auth";

export default async function ImportPage() {
  const user = await getUser();
  if (!user) redirect("/login");

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

        <div className="flex items-center justify-between mb-2">
          <h2 className="text-xl font-semibold text-text-primary">Import Library</h2>
          <Link
            href="/settings/imports/pending"
            className="text-sm text-accent hover:underline transition-colors"
          >
            Review pending imports
          </Link>
        </div>
        <p className="text-sm text-text-primary mb-8">
          Import your books, ratings, and reviews from a CSV export.
        </p>

        <ImportTabs username={user.username} />
      </main>
    </div>
  );
}
