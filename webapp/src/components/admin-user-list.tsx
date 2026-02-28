"use client";

import { useState, useEffect, useCallback } from "react";

type AdminUser = {
  user_id: string;
  username: string;
  display_name: string | null;
  email: string;
  is_moderator: boolean;
  author_key: string | null;
};

export default function AdminUserList() {
  const [query, setQuery] = useState("");
  const [users, setUsers] = useState<AdminUser[]>([]);
  const [page, setPage] = useState(1);
  const [hasNext, setHasNext] = useState(false);
  const [loading, setLoading] = useState(true);
  const [toggling, setToggling] = useState<string | null>(null);
  const [editingAuthor, setEditingAuthor] = useState<string | null>(null);
  const [authorKeyInput, setAuthorKeyInput] = useState("");
  const [savingAuthor, setSavingAuthor] = useState<string | null>(null);

  const fetchUsers = useCallback(async (q: string, p: number) => {
    setLoading(true);
    const params = new URLSearchParams({ page: String(p) });
    if (q) params.set("q", q);
    const res = await fetch(`/api/admin/users?${params}`);
    if (res.ok) {
      const data = await res.json();
      setUsers(data.users ?? []);
      setHasNext(data.has_next ?? false);
    }
    setLoading(false);
  }, []);

  useEffect(() => {
    const timer = setTimeout(() => {
      setPage(1);
      fetchUsers(query, 1);
    }, 300);
    return () => clearTimeout(timer);
  }, [query, fetchUsers]);

  useEffect(() => {
    fetchUsers(query, page);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [page]);

  async function toggleModerator(user: AdminUser) {
    setToggling(user.user_id);
    const res = await fetch(`/api/admin/users/${user.user_id}/moderator`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ is_moderator: !user.is_moderator }),
    });
    if (res.ok) {
      setUsers((prev) =>
        prev.map((u) =>
          u.user_id === user.user_id
            ? { ...u, is_moderator: !u.is_moderator }
            : u
        )
      );
    }
    setToggling(null);
  }

  function startEditAuthor(user: AdminUser) {
    setEditingAuthor(user.user_id);
    setAuthorKeyInput(user.author_key ?? "");
  }

  function cancelEditAuthor() {
    setEditingAuthor(null);
    setAuthorKeyInput("");
  }

  async function saveAuthorKey(user: AdminUser) {
    setSavingAuthor(user.user_id);
    const newKey = authorKeyInput.trim() || null;
    const res = await fetch(`/api/admin/users/${user.user_id}/author`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ author_key: newKey }),
    });
    if (res.ok) {
      setUsers((prev) =>
        prev.map((u) =>
          u.user_id === user.user_id ? { ...u, author_key: newKey } : u
        )
      );
      setEditingAuthor(null);
      setAuthorKeyInput("");
    }
    setSavingAuthor(null);
  }

  return (
    <div>
      <input
        type="search"
        placeholder="Search users by name or email..."
        value={query}
        onChange={(e) => setQuery(e.target.value)}
        className="w-full max-w-md px-3 py-2 text-sm border border-border rounded bg-surface-2 text-text-primary placeholder-text-tertiary focus:outline-none focus:ring-2 focus:ring-accent focus:border-transparent mb-4"
      />

      {loading ? (
        <p className="text-sm text-text-primary">Loading...</p>
      ) : users.length === 0 ? (
        <p className="text-sm text-text-primary">No users found.</p>
      ) : (
        <>
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-border text-left text-text-primary">
                <th className="pb-2 font-medium">Username</th>
                <th className="pb-2 font-medium">Display Name</th>
                <th className="pb-2 font-medium">Email</th>
                <th className="pb-2 font-medium">Author</th>
                <th className="pb-2 font-medium text-right">Moderator</th>
              </tr>
            </thead>
            <tbody>
              {users.map((u) => (
                <tr
                  key={u.user_id}
                  className="border-b border-border hover:bg-surface-2"
                >
                  <td className="py-2 text-text-primary font-medium">
                    {u.username}
                  </td>
                  <td className="py-2 text-text-primary">
                    {u.display_name ?? "—"}
                  </td>
                  <td className="py-2 text-text-primary">{u.email}</td>
                  <td className="py-2">
                    {editingAuthor === u.user_id ? (
                      <div className="flex items-center gap-1">
                        <input
                          type="text"
                          value={authorKeyInput}
                          onChange={(e) => setAuthorKeyInput(e.target.value)}
                          placeholder="OL author key"
                          className="w-28 px-2 py-0.5 text-xs border border-border rounded bg-surface-0 text-text-primary placeholder-text-tertiary focus:outline-none focus:ring-1 focus:ring-accent"
                        />
                        <button
                          onClick={() => saveAuthorKey(u)}
                          disabled={savingAuthor === u.user_id}
                          className="px-2 py-0.5 rounded text-xs font-medium bg-accent text-text-inverted hover:bg-accent-hover disabled:opacity-50"
                        >
                          Save
                        </button>
                        <button
                          onClick={cancelEditAuthor}
                          className="px-2 py-0.5 rounded text-xs text-text-primary hover:text-text-primary"
                        >
                          Cancel
                        </button>
                      </div>
                    ) : (
                      <button
                        onClick={() => startEditAuthor(u)}
                        className={`px-3 py-1 rounded text-xs font-medium transition-colors ${
                          u.author_key
                            ? "bg-amber-50 text-amber-700 border border-amber-200 hover:bg-amber-100"
                            : "bg-surface-2 text-text-primary hover:bg-surface-2"
                        }`}
                      >
                        {u.author_key ? `Author (${u.author_key})` : "Set author"}
                      </button>
                    )}
                  </td>
                  <td className="py-2 text-right">
                    <button
                      onClick={() => toggleModerator(u)}
                      disabled={toggling === u.user_id}
                      className={`px-3 py-1 rounded text-xs font-medium transition-colors ${
                        u.is_moderator
                          ? "bg-accent text-text-inverted hover:bg-accent-hover"
                          : "bg-surface-2 text-text-primary hover:bg-surface-2"
                      } disabled:opacity-50`}
                    >
                      {u.is_moderator ? "Moderator" : "Grant"}
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>

          <div className="flex items-center gap-4 mt-4">
            <button
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              disabled={page === 1}
              className="text-sm text-text-primary hover:text-text-primary disabled:opacity-30 disabled:cursor-not-allowed"
            >
              Previous
            </button>
            <span className="text-sm text-text-primary">Page {page}</span>
            <button
              onClick={() => setPage((p) => p + 1)}
              disabled={!hasNext}
              className="text-sm text-text-primary hover:text-text-primary disabled:opacity-30 disabled:cursor-not-allowed"
            >
              Next
            </button>
          </div>
        </>
      )}
    </div>
  );
}
