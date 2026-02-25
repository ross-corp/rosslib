import { redirect } from "next/navigation";
import Link from "next/link";
import BookScanner from "@/components/book-scanner";
import { type StatusValue } from "@/components/shelf-picker";
import { getUser, getToken } from "@/lib/auth";

type TagKey = {
  id: string;
  name: string;
  slug: string;
  mode: string;
  values: StatusValue[];
};

async function fetchTagKeys(token: string): Promise<TagKey[]> {
  const res = await fetch(`${process.env.API_URL}/me/tag-keys`, {
    headers: { Authorization: `Bearer ${token}` },
    cache: "no-store",
  });
  if (!res.ok) return [];
  return res.json();
}

async function fetchStatusMap(token: string): Promise<Record<string, string>> {
  const res = await fetch(`${process.env.API_URL}/me/books/status-map`, {
    headers: { Authorization: `Bearer ${token}` },
    cache: "no-store",
  });
  if (!res.ok) return {};
  return res.json();
}

export default async function ScanPage() {
  const user = await getUser();
  if (!user) redirect("/login");

  const token = await getToken();
  const [tagKeys, statusMap] = await Promise.all([
    token ? fetchTagKeys(token) : Promise.resolve([]),
    token ? fetchStatusMap(token) : Promise.resolve({}),
  ]);

  const statusKey = tagKeys.find((k) => k.slug === "status") ?? null;
  const statusValues: StatusValue[] = statusKey ? statusKey.values : [];
  const statusKeyId: string | null = statusKey?.id ?? null;

  return (
    <div className="min-h-screen">
      <main className="max-w-5xl mx-auto px-4 sm:px-6 py-10">
        <div className="mb-8 flex items-center gap-2 text-sm text-text-primary">
          <Link href={`/${user.username}`} className="hover:text-text-primary transition-colors">
            {user.username}
          </Link>
          <span>/</span>
          <span className="text-text-primary">Scan</span>
        </div>

        <h1 className="text-2xl font-bold text-text-primary mb-2">Scan a Book</h1>
        <p className="text-sm text-text-primary mb-8">
          Scan an ISBN barcode on the back of a book to quickly look it up and add it to your library.
        </p>

        <BookScanner
          statusValues={statusValues}
          statusKeyId={statusKeyId}
          bookStatusMap={statusMap}
        />
      </main>
    </div>
  );
}
