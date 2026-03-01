import { redirect } from "next/navigation";
import Link from "next/link";
import FollowedAuthorsList from "./followed-authors-list";
import { getUser, getToken } from "@/lib/auth";

type FollowedAuthor = {
  author_key: string;
  author_name: string;
};

async function fetchFollowedAuthors(token: string): Promise<FollowedAuthor[]> {
  const res = await fetch(`${process.env.API_URL}/me/followed-authors`, {
    headers: { Authorization: `Bearer ${token}` },
    cache: "no-store",
  });
  if (!res.ok) return [];
  return res.json();
}

export default async function FollowedAuthorsPage() {
  const [user, token] = await Promise.all([getUser(), getToken()]);
  if (!user || !token) redirect("/login");

  const authors = await fetchFollowedAuthors(token);

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
          <span className="text-text-primary">Followed authors</span>
        </div>

        <h1 className="text-2xl font-bold text-text-primary mb-2">Followed authors</h1>
        <p className="text-sm text-text-primary mb-8">
          Authors you follow will notify you when they have new publications.
        </p>

        <FollowedAuthorsList initialAuthors={authors} />
      </main>
    </div>
  );
}
