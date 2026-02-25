import { redirect } from "next/navigation";
import Link from "next/link";
import ImportForm from "@/components/import-form";
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
          <Link href="/settings" className="hover:text-text-primary transition-colors">
            Settings
          </Link>
          <span>/</span>
          <span className="text-text-primary">Import</span>
        </div>

        <h1 className="text-2xl font-bold text-text-primary mb-2">Import from Goodreads</h1>
        <p className="text-sm text-text-primary mb-8">
          Import your books, shelves, ratings, and reviews from a Goodreads CSV export.
        </p>

        <ImportForm />
      </main>
    </div>
  );
}
