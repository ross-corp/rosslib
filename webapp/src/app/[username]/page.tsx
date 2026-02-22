import { notFound } from "next/navigation";
import Nav from "@/components/nav";
import { getUser } from "@/lib/auth";

type UserProfile = {
  user_id: string;
  username: string;
  display_name: string | null;
  bio: string | null;
  avatar_url: string | null;
  is_private: boolean;
  member_since: string;
};

async function fetchProfile(username: string): Promise<UserProfile | null> {
  const res = await fetch(`${process.env.API_URL}/users/${username}`, {
    cache: "no-store",
  });
  if (!res.ok) return null;
  return res.json();
}

export default async function UserPage({
  params,
}: {
  params: Promise<{ username: string }>;
}) {
  const { username } = await params;
  const [profile, currentUser] = await Promise.all([
    fetchProfile(username),
    getUser(),
  ]);

  if (!profile) notFound();

  const isOwnProfile = currentUser?.user_id === profile.user_id;
  const memberSince = new Date(profile.member_since).toLocaleDateString(
    "en-US",
    { month: "long", year: "numeric" }
  );

  return (
    <div className="min-h-screen">
      <Nav />
      <main className="max-w-5xl mx-auto px-4 sm:px-6 py-12">
        <div className="max-w-xl">
          <div className="flex items-start justify-between mb-6">
            <div>
              <h1 className="text-2xl font-bold text-stone-900">
                {profile.display_name || profile.username}
              </h1>
              {profile.display_name && (
                <p className="text-stone-400 text-sm mt-0.5">
                  @{profile.username}
                </p>
              )}
            </div>
            {isOwnProfile && (
              <button className="text-sm text-stone-500 hover:text-stone-900 border border-stone-300 px-3 py-1.5 rounded transition-colors">
                Edit profile
              </button>
            )}
          </div>

          {profile.bio && (
            <p className="text-stone-700 text-sm leading-relaxed mb-6">
              {profile.bio}
            </p>
          )}

          <p className="text-xs text-stone-400">Member since {memberSince}</p>
        </div>
      </main>
    </div>
  );
}
