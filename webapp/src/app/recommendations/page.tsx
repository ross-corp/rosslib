import { redirect } from "next/navigation";
import Link from "next/link";
import { getUser, getToken } from "@/lib/auth";
import { formatTime } from "@/components/activity";
import RecommendationList from "./recommendation-list";

type Recommendation = {
  id: string;
  note: string | null;
  status: string;
  created_at: string;
  sender: {
    user_id: string;
    username: string;
    display_name: string | null;
    avatar_url: string | null;
  };
  book: {
    open_library_id: string;
    title: string;
    cover_url: string | null;
    authors: string | null;
  };
};

async function fetchRecommendations(
  token: string,
  status?: string
): Promise<Recommendation[]> {
  const url = new URL(`${process.env.API_URL}/me/recommendations`);
  if (status) url.searchParams.set("status", status);

  const res = await fetch(url.toString(), {
    cache: "no-store",
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!res.ok) return [];
  return res.json();
}

export default async function RecommendationsPage({
  searchParams,
}: {
  searchParams: Promise<{ status?: string }>;
}) {
  const [user, token] = await Promise.all([getUser(), getToken()]);
  if (!user || !token) redirect("/login");

  const { status } = await searchParams;
  const filter = status || "pending";
  const recommendations = await fetchRecommendations(token, filter);

  return (
    <div className="min-h-screen">
      <main className="max-w-2xl mx-auto px-4 sm:px-6 py-12">
        <h1 className="text-2xl font-bold text-text-primary mb-6">
          Recommendations
        </h1>

        {/* Status filter tabs */}
        <div className="flex gap-1 mb-6 border-b border-border">
          {["pending", "seen", "dismissed", "all"].map((s) => (
            <Link
              key={s}
              href={`/recommendations?status=${s}`}
              className={`px-3 py-2 text-xs font-medium transition-colors ${
                filter === s
                  ? "text-text-primary border-b-2 border-text-primary"
                  : "text-text-secondary hover:text-text-primary"
              }`}
            >
              {s.charAt(0).toUpperCase() + s.slice(1)}
            </Link>
          ))}
        </div>

        {recommendations.length === 0 ? (
          <div className="text-center py-16">
            <p className="text-text-secondary text-sm">
              {filter === "pending"
                ? "No pending recommendations."
                : "No recommendations found."}
            </p>
          </div>
        ) : (
          <RecommendationList recommendations={recommendations} />
        )}
      </main>
    </div>
  );
}
