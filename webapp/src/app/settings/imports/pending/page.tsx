import { redirect } from "next/navigation";
import Link from "next/link";
import PendingImportsManager from "@/components/pending-imports-manager";
import { getUser, getToken } from "@/lib/auth";

export default async function PendingImportsPage() {
  const user = await getUser();
  if (!user) redirect("/login");

  const token = await getToken();
  let items: any[] = [];
  try {
    const res = await fetch(`${process.env.API_URL}/me/imports/pending`, {
      headers: { Authorization: `Bearer ${token}` },
      cache: "no-store",
    });
    if (res.ok) {
      const data = await res.json();
      items = Array.isArray(data) ? data : [];
    }
  } catch {
    // fall through with empty list
  }

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
          <Link href="/settings/import" className="hover:text-text-primary transition-colors">
            Import
          </Link>
          <span>/</span>
          <span className="text-text-primary">Pending</span>
        </div>

        <h1 className="text-2xl font-bold text-text-primary mb-2">Pending Imports</h1>
        <p className="text-sm text-text-primary mb-8">
          Review unmatched books from previous imports. Search for the correct book to add it to your library, or dismiss entries you don&apos;t need.
        </p>

        <PendingImportsManager initialItems={items} />
      </main>
    </div>
  );
}
