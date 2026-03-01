import { redirect } from "next/navigation";
import Link from "next/link";
import { getUser, getToken } from "@/lib/auth";
import RecommendationList from "./recommendation-list";
import SentRecommendationList from "./sent-recommendation-list";

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

type SentRecommendation = {
  id: string;
  note: string | null;
  status: string;
  created_at: string;
  recipient: {
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

async function fetchSentRecommendations(
  token: string
): Promise<SentRecommendation[]> {
  const res = await fetch(
    `${process.env.API_URL}/me/recommendations/sent`,
    {
      cache: "no-store",
      headers: { Authorization: `Bearer ${token}` },
    }
  );
  if (!res.ok) return [];
  return res.json();
}

export default async function RecommendationsPage({
  searchParams,
}: {
  searchParams: Promise<{ tab?: string; status?: string }>;
}) {
  const [user, token] = await Promise.all([getUser(), getToken()]);
  if (!user || !token) redirect("/login");

  const { tab, status } = await searchParams;
  const activeTab = tab === "sent" ? "sent" : "received";
  const filter = status || "pending";

  const [recommendations, sentRecommendations] = await Promise.all([
    activeTab === "received" ? fetchRecommendations(token, filter) : Promise.resolve([]),
    activeTab === "sent" ? fetchSentRecommendations(token) : Promise.resolve([]),
  ]);

  return (
    <div className="min-h-screen">
      <main className="max-w-2xl mx-auto px-4 sm:px-6 py-12">
        <h1 className="text-2xl font-bold text-text-primary mb-6">
          Recommendations
        </h1>

        {/* Received / Sent tabs */}
        <div className="flex gap-1 mb-6 border-b border-border">
          <Link
            href="/recommendations"
            className={`px-3 py-2 text-sm font-medium transition-colors ${
              activeTab === "received"
                ? "text-text-primary border-b-2 border-text-primary"
                : "text-text-secondary hover:text-text-primary"
            }`}
          >
            Received
          </Link>
          <Link
            href="/recommendations?tab=sent"
            className={`px-3 py-2 text-sm font-medium transition-colors ${
              activeTab === "sent"
                ? "text-text-primary border-b-2 border-text-primary"
                : "text-text-secondary hover:text-text-primary"
            }`}
          >
            Sent
          </Link>
        </div>

        {activeTab === "received" && (
          <>
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
          </>
        )}

        {activeTab === "sent" && (
          <>
            {sentRecommendations.length === 0 ? (
              <div className="text-center py-16">
                <p className="text-text-secondary text-sm">
                  You haven&apos;t sent any recommendations yet.
                </p>
              </div>
            ) : (
              <SentRecommendationList recommendations={sentRecommendations} />
            )}
          </>
        )}
      </main>
    </div>
  );
}
