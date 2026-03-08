"use client";

import { useEffect, useState } from "react";

type APIToken = {
  id: string;
  name: string;
  created: string;
  last_used_at: string;
};

export default function APITokensForm() {
  const [tokens, setTokens] = useState<APIToken[]>([]);
  const [loading, setLoading] = useState(true);
  const [name, setName] = useState("");
  const [creating, setCreating] = useState(false);
  const [error, setError] = useState("");
  const [newToken, setNewToken] = useState("");
  const [deletingId, setDeletingId] = useState<string | null>(null);

  async function fetchTokens() {
    const res = await fetch("/api/me/api-tokens");
    if (res.ok) {
      const data = await res.json();
      setTokens(data.tokens || []);
    }
    setLoading(false);
  }

  useEffect(() => {
    fetchTokens();
  }, []);

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setNewToken("");

    if (!name.trim()) {
      setError("Token name is required.");
      return;
    }

    setCreating(true);
    const res = await fetch("/api/me/api-tokens", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name: name.trim() }),
    });

    const data = await res.json();
    setCreating(false);

    if (!res.ok) {
      setError(data.error || "Failed to create token.");
      return;
    }

    setNewToken(data.token);
    setName("");
    fetchTokens();
  }

  async function handleDelete(tokenId: string) {
    setDeletingId(tokenId);
    const res = await fetch(`/api/me/api-tokens/${tokenId}`, {
      method: "DELETE",
    });
    setDeletingId(null);

    if (res.ok) {
      setTokens((prev) => prev.filter((t) => t.id !== tokenId));
    }
  }

  function formatDate(dateStr: string) {
    if (!dateStr) return "Never";
    return new Date(dateStr).toLocaleDateString("en-US", {
      year: "numeric",
      month: "short",
      day: "numeric",
    });
  }

  if (loading) return null;

  return (
    <div className="space-y-6">
      {newToken && (
        <div className="bg-green-50 border border-green-200 rounded p-4">
          <p className="text-sm font-medium text-green-800 mb-2">
            Token created. Copy it now — you won&apos;t see it again.
          </p>
          <code className="block bg-white border border-green-300 rounded px-3 py-2 text-sm font-mono text-green-900 break-all select-all">
            {newToken}
          </code>
        </div>
      )}

      <form onSubmit={handleCreate} className="space-y-3 max-w-md">
        {error && (
          <p className="text-sm text-red-600 bg-red-50 border border-red-200 rounded px-3 py-2">
            {error}
          </p>
        )}
        <div>
          <label
            htmlFor="token_name"
            className="block text-sm font-medium text-text-primary mb-1"
          >
            Token name
          </label>
          <input
            id="token_name"
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder='e.g. "CLI", "Calibre"'
            maxLength={100}
            className="w-full px-3 py-2 border border-border rounded text-text-primary placeholder-text-tertiary focus:outline-none focus:ring-2 focus:ring-accent focus:border-transparent text-sm"
          />
        </div>
        <button
          type="submit"
          disabled={creating || tokens.length >= 5}
          className="bg-accent text-text-inverted px-4 py-2 rounded text-sm font-medium hover:bg-surface-3 transition-colors disabled:opacity-50"
        >
          {creating ? "Creating..." : "Create token"}
        </button>
        {tokens.length >= 5 && (
          <p className="text-sm text-text-secondary">
            Maximum of 5 tokens reached. Delete one to create a new token.
          </p>
        )}
      </form>

      {tokens.length > 0 && (
        <div className="border border-border rounded overflow-hidden">
          <table className="w-full text-sm">
            <thead>
              <tr className="bg-surface-2 text-text-primary">
                <th className="text-left px-4 py-2 font-medium">Name</th>
                <th className="text-left px-4 py-2 font-medium">Created</th>
                <th className="text-left px-4 py-2 font-medium">Last used</th>
                <th className="px-4 py-2"></th>
              </tr>
            </thead>
            <tbody>
              {tokens.map((t) => (
                <tr key={t.id} className="border-t border-border">
                  <td className="px-4 py-2 text-text-primary font-mono">
                    {t.name}
                  </td>
                  <td className="px-4 py-2 text-text-secondary">
                    {formatDate(t.created)}
                  </td>
                  <td className="px-4 py-2 text-text-secondary">
                    {formatDate(t.last_used_at)}
                  </td>
                  <td className="px-4 py-2 text-right">
                    <button
                      onClick={() => handleDelete(t.id)}
                      disabled={deletingId === t.id}
                      className="text-red-600 hover:text-red-800 text-sm font-medium disabled:opacity-50"
                    >
                      {deletingId === t.id ? "Deleting..." : "Revoke"}
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {tokens.length === 0 && (
        <p className="text-sm text-text-secondary">
          No API tokens yet. Create one to use the API from external tools.
        </p>
      )}
    </div>
  );
}
