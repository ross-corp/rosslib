import { redirect } from "next/navigation";
import Link from "next/link";
import Nav from "@/components/nav";
import ImportForm from "@/components/import-form";
import { getUser } from "@/lib/auth";

export default async function ImportPage() {
  const user = await getUser();
  if (!user) redirect("/login");

  return (
    <div className="min-h-screen">
      <Nav />
      <main className="max-w-5xl mx-auto px-4 sm:px-6 py-10">
        <div className="mb-8 flex items-center gap-2 text-sm text-stone-400">
          <Link href={`/${user.username}`} className="hover:text-stone-700 transition-colors">
            {user.username}
          </Link>
          <span>/</span>
          <Link href="/settings" className="hover:text-stone-700 transition-colors">
            Settings
          </Link>
          <span>/</span>
          <span className="text-stone-600">Import</span>
        </div>

        <h1 className="text-2xl font-bold text-stone-900 mb-2">Import from Goodreads</h1>
        <p className="text-sm text-stone-500 mb-8">
          Import your books, shelves, ratings, and reviews from a Goodreads CSV export.
        </p>

        <ImportForm />
      </main>
    </div>
  );
}
