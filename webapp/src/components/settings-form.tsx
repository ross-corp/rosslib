"use client";

import { useRef, useState } from "react";
import { useRouter } from "next/navigation";

export default function SettingsForm({
  username,
  initialDisplayName,
  initialBio,
  initialAvatarUrl,
  initialBannerUrl,
  initialIsPrivate,
}: {
  username: string;
  initialDisplayName: string;
  initialBio: string;
  initialAvatarUrl: string | null;
  initialBannerUrl: string | null;
  initialIsPrivate: boolean;
}) {
  const router = useRouter();
  const [displayName, setDisplayName] = useState(initialDisplayName);
  const [bio, setBio] = useState(initialBio);
  const [isPrivate, setIsPrivate] = useState(initialIsPrivate);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [saved, setSaved] = useState(false);

  const [avatarUrl, setAvatarUrl] = useState(initialAvatarUrl);
  const [avatarUploading, setAvatarUploading] = useState(false);
  const [avatarError, setAvatarError] = useState("");
  const fileInputRef = useRef<HTMLInputElement>(null);

  const [bannerUrl, setBannerUrl] = useState(initialBannerUrl);
  const [bannerUploading, setBannerUploading] = useState(false);
  const [bannerError, setBannerError] = useState("");
  const bannerInputRef = useRef<HTMLInputElement>(null);

  async function handleAvatarChange(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file) return;

    setAvatarUploading(true);
    setAvatarError("");

    const formData = new FormData();
    formData.append("avatar", file);

    const res = await fetch("/api/me/avatar", {
      method: "POST",
      body: formData,
    });

    setAvatarUploading(false);

    if (!res.ok) {
      const data = await res.json();
      setAvatarError(data.error || "Upload failed.");
      return;
    }

    const data = await res.json();
    setAvatarUrl(data.avatar_url);
    router.refresh();

    // Reset the input so the same file can be re-selected if needed.
    if (fileInputRef.current) fileInputRef.current.value = "";
  }

  async function handleBannerChange(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file) return;

    setBannerUploading(true);
    setBannerError("");

    const formData = new FormData();
    formData.append("banner", file);

    const res = await fetch("/api/me/banner", {
      method: "POST",
      body: formData,
    });

    setBannerUploading(false);

    if (!res.ok) {
      const data = await res.json();
      setBannerError(data.error || "Upload failed.");
      return;
    }

    const data = await res.json();
    setBannerUrl(data.banner_url);
    router.refresh();

    if (bannerInputRef.current) bannerInputRef.current.value = "";
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setSaved(false);
    setLoading(true);

    const res = await fetch("/api/users/me", {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ display_name: displayName, bio, is_private: isPrivate }),
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
    <div className="max-w-md space-y-8">
      {/* Avatar */}
      <div>
        <p className="text-sm font-medium text-text-primary mb-3">Photo</p>
        <div className="flex items-center gap-4">
          <div className="relative shrink-0">
            {avatarUrl ? (
              <img
                src={avatarUrl}
                alt=""
                className="w-20 h-20 rounded-full object-cover bg-surface-2"
              />
            ) : (
              <div className="w-20 h-20 rounded-full bg-surface-2 flex items-center justify-center">
                <span className="text-text-primary text-2xl font-medium select-none">
                  {username[0].toUpperCase()}
                </span>
              </div>
            )}
            {avatarUploading && (
              <div className="absolute inset-0 rounded-full bg-black/40 flex items-center justify-center">
                <span className="text-white text-xs">...</span>
              </div>
            )}
          </div>
          <div>
            <label className="cursor-pointer text-sm text-text-primary hover:text-text-primary underline underline-offset-2 transition-colors">
              Change photo
              <input
                ref={fileInputRef}
                type="file"
                accept="image/jpeg,image/png,image/gif,image/webp"
                className="hidden"
                onChange={handleAvatarChange}
                disabled={avatarUploading}
              />
            </label>
            {avatarError && (
              <p className="text-xs text-red-600 mt-1">{avatarError}</p>
            )}
          </div>
        </div>
      </div>

      {/* Banner */}
      <div>
        <p className="text-sm font-medium text-text-primary mb-3">Banner</p>
        <div className="relative">
          {bannerUrl ? (
            <img
              src={bannerUrl}
              alt=""
              className="w-full h-32 object-cover rounded-lg bg-surface-2"
            />
          ) : (
            <div className="w-full h-32 rounded-lg bg-surface-2 flex items-center justify-center">
              <span className="text-text-tertiary text-sm">No banner</span>
            </div>
          )}
          {bannerUploading && (
            <div className="absolute inset-0 rounded-lg bg-black/40 flex items-center justify-center">
              <span className="text-white text-xs">...</span>
            </div>
          )}
        </div>
        <div className="mt-2">
          <label className="cursor-pointer text-sm text-text-primary hover:text-text-primary underline underline-offset-2 transition-colors">
            Change banner
            <input
              ref={bannerInputRef}
              type="file"
              accept="image/jpeg,image/png,image/gif,image/webp"
              className="hidden"
              onChange={handleBannerChange}
              disabled={bannerUploading}
            />
          </label>
          <span className="text-xs text-text-tertiary ml-2">Recommended: 1200x300</span>
          {bannerError && (
            <p className="text-xs text-red-600 mt-1">{bannerError}</p>
          )}
        </div>
      </div>

      {/* Profile fields */}
      <form onSubmit={handleSubmit} className="space-y-5">
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
            className="block text-sm font-medium text-text-primary mb-1"
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
            className="w-full px-3 py-2 border border-border rounded text-text-primary placeholder-text-tertiary focus:outline-none focus:ring-2 focus:ring-accent focus:border-transparent text-sm"
          />
        </div>

        <div>
          <label
            htmlFor="bio"
            className="block text-sm font-medium text-text-primary mb-1"
          >
            Byline
          </label>
          <textarea
            id="bio"
            value={bio}
            onChange={(e) => setBio(e.target.value)}
            placeholder="A short line about yourself"
            rows={3}
            maxLength={2000}
            className="w-full px-3 py-2 border border-border rounded text-text-primary placeholder-text-tertiary focus:outline-none focus:ring-2 focus:ring-accent focus:border-transparent text-sm resize-none"
          />
        </div>

        <div>
          <label className="flex items-center gap-2 cursor-pointer">
            <input
              type="checkbox"
              checked={isPrivate}
              onChange={(e) => setIsPrivate(e.target.checked)}
              className="rounded border-border text-text-primary focus:ring-accent"
            />
            <span className="text-sm text-text-primary">
              Private account
            </span>
          </label>
          <p className="text-xs text-text-primary mt-1 ml-6">
            Only approved followers can see your books and activity
          </p>
        </div>

        <div className="flex items-center gap-3 pt-1">
          <button
            type="submit"
            disabled={loading}
            className="bg-accent text-text-inverted px-4 py-2 rounded text-sm font-medium hover:bg-surface-3 transition-colors disabled:opacity-50"
          >
            {loading ? "Saving..." : "Save"}
          </button>
          <a
            href={`/${username}`}
            className="text-sm text-text-primary hover:text-text-primary transition-colors"
          >
            Cancel
          </a>
        </div>
      </form>
    </div>
  );
}
