"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

export default function SettingsForm({
  username,
  initialDisplayName,
  initialBio,
}: {
  username: string;
  initialDisplayName: string;
  initialBio: string;
}) {
  const router = useRouter();
  const [displayName, setDisplayName] = useState(initialDisplayName);
  const [bio, setBio] = useState(initialBio);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [saved, setSaved] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setSaved(false);
    setLoading(true);

    const res = await fetch("/api/users/me", {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ display_name: displayName, bio }),
    });

    setLoading(false);

    if (!res.ok) {
      const data = await res.json();
      setError(data.error || "Something went wrong.");
      return;
    }

    setSaved(true);
    router.refresh();
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-5 max-w-md">
      {error && (
        <p className="text-sm text-red-600 bg-red-50 border border-red-200 rounded px-3 py-2">
          {error}
        </p>
      )}
      {saved && (
        <p className="text-sm text-green-700 bg-green-50 border border-green-200 rounded px-3 py-2">
          Profile updated.
        </p>
      )}

      <div>
        <label
          htmlFor="display_name"
          className="block text-sm font-medium text-stone-700 mb-1"
        >
          Display name
        </label>
        <input
          id="display_name"
          type="text"
          value={displayName}
          onChange={(e) => setDisplayName(e.target.value)}
          placeholder={username}
          maxLength={100}
          className="w-full px-3 py-2 border border-stone-300 rounded text-stone-900 placeholder-stone-400 focus:outline-none focus:ring-2 focus:ring-stone-900 focus:border-transparent text-sm"
        />
      </div>

      <div>
        <label
          htmlFor="bio"
          className="block text-sm font-medium text-stone-700 mb-1"
        >
          Byline
        </label>
        <textarea
          id="bio"
          value={bio}
          onChange={(e) => setBio(e.target.value)}
          placeholder="A short line about yourself"
          rows={3}
          className="w-full px-3 py-2 border border-stone-300 rounded text-stone-900 placeholder-stone-400 focus:outline-none focus:ring-2 focus:ring-stone-900 focus:border-transparent text-sm resize-none"
        />
      </div>

      <div className="flex items-center gap-3 pt-1">
        <button
          type="submit"
          disabled={loading}
          className="bg-stone-900 text-white px-4 py-2 rounded text-sm font-medium hover:bg-stone-700 transition-colors disabled:opacity-50"
        >
          {loading ? "Saving..." : "Save"}
        </button>
        <a
          href={`/${username}`}
          className="text-sm text-stone-500 hover:text-stone-900 transition-colors"
        >
          Cancel
        </a>
      </div>
    </form>
  );
}
