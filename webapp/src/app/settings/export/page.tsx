import { redirect } from "next/navigation";
import Link from "next/link";
import ExportForm from "@/components/export-form";
import { getUser, getToken } from "@/lib/auth";

type Shelf = {
  id: string;
  name: string;
  slug: string;
  item_count: number;
};

async function fetchShelves(username: string, token: string): Promise<Shelf[]> {
  const res = await fetch(`${process.env.API_URL}/users/${username}/shelves`, {
    headers: { Authorization: `Bearer ${token}` },
    cache: "no-store",
  });
  if (!res.ok) return [];
  return res.json();
}

export default async function ExportPage() {
  const user = await getUser();
  if (!user) redirect("/login");

  const token = await getToken();
  const shelves = await fetchShelves(user.username, token || "");

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
          <span className="text-text-primary">Export</span>
        </div>

        <h1 className="text-2xl font-bold text-text-primary mb-2">Export to CSV</h1>
        <p className="text-sm text-text-primary mb-8">
          Download your library as a CSV file. Export all shelves or pick a specific one.
        </p>

        <ExportForm shelves={shelves} />
      </main>
    </div>
  );
}
