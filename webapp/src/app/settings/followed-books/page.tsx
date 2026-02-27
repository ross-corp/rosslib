import { redirect } from "next/navigation";
import Link from "next/link";
import FollowedBooksList from "./followed-books-list";
import { getUser, getToken } from "@/lib/auth";

type FollowedBook = {
  open_library_id: string;
  title: string;
  authors: string[] | null;
  cover_url: string | null;
};

async function fetchFollowedBooks(token: string): Promise<FollowedBook[]> {
  const res = await fetch(`${process.env.API_URL}/me/followed-books`, {
    headers: { Authorization: `Bearer ${token}` },
    cache: "no-store",
  });
  if (!res.ok) return [];
  return res.json();
}

export default async function FollowedBooksPage() {
  const [user, token] = await Promise.all([getUser(), getToken()]);
  if (!user || !token) redirect("/login");

  const books = await fetchFollowedBooks(token);

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
          <span className="text-text-primary">Followed books</span>
        </div>

        <h1 className="text-2xl font-bold text-text-primary mb-2">Followed books</h1>
        <p className="text-sm text-text-primary mb-8">
          Books you follow will notify you of new threads and activity.
        </p>

        <FollowedBooksList initialBooks={books} />
      </main>
    </div>
  );
}
