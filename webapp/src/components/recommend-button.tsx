"use client";

import { useState, useEffect, useRef } from "react";

type UserResult = {
  user_id: string;
  username: string;
  display_name: string | null;
  avatar_url: string | null;
};

export default function RecommendButton({
  bookOlId,
  bookTitle,
}: {
  bookOlId: string;
  bookTitle: string;
}) {
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState("");
  const [users, setUsers] = useState<UserResult[]>([]);
  const [searching, setSearching] = useState(false);
  const [selectedUser, setSelectedUser] = useState<UserResult | null>(null);
  const [note, setNote] = useState("");
  const [sending, setSending] = useState(false);
  const [sent, setSent] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const debounceRef = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);

  useEffect(() => {
    if (!query.trim()) {
      setUsers([]);
      return;
    }
    if (debounceRef.current) clearTimeout(debounceRef.current);
    debounceRef.current = setTimeout(async () => {
      setSearching(true);
      try {
        const res = await fetch(
          `/api/users?q=${encodeURIComponent(query.trim())}`
        );
        if (res.ok) {
          const data = await res.json();
          setUsers(data);
        }
      } finally {
        setSearching(false);
      }
    }, 300);
    return () => {
      if (debounceRef.current) clearTimeout(debounceRef.current);
    };
  }, [query]);

  function reset() {
    setQuery("");
    setUsers([]);
    setSelectedUser(null);
    setNote("");
    setSending(false);
    setSent(false);
    setError(null);
  }

  function handleClose() {
    setOpen(false);
    reset();
  }

  async function handleSend() {
    if (!selectedUser) return;
    setSending(true);
    setError(null);
    try {
      const res = await fetch("/api/me/recommendations", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          username: selectedUser.username,
          book_ol_id: bookOlId,
          note,
        }),
      });
      if (res.ok) {
        setSent(true);
      } else {
        const data = await res.json();
        setError(data.error || "Failed to send recommendation");
      }
    } catch {
      setError("Failed to send recommendation");
    } finally {
      setSending(false);
    }
  }

  return (
    <>
      <button
        onClick={() => setOpen(true)}
        className="text-xs text-text-secondary hover:text-text-primary transition-colors border border-border rounded px-2.5 py-1.5"
        title="Recommend this book to a friend"
      >
        Recommend
      </button>

      {open && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
          onClick={handleClose}
        >
          <div
            className="bg-surface-0 border border-border rounded-lg shadow-lg w-full max-w-md flex flex-col mx-4"
            onClick={(e) => e.stopPropagation()}
          >
            <div className="flex items-center justify-between p-4 border-b border-border">
              <h3 className="text-sm font-semibold text-text-primary">
                Recommend &ldquo;{bookTitle}&rdquo;
              </h3>
              <button
                onClick={handleClose}
                className="text-text-secondary hover:text-text-primary text-sm"
              >
                Close
              </button>
            </div>

            <div className="p-4">
              {sent ? (
                <div className="text-center py-4">
                  <p className="text-sm text-text-primary mb-3">
                    Recommendation sent to{" "}
                    <span className="font-medium">
                      {selectedUser?.display_name ?? selectedUser?.username}
                    </span>
                    !
                  </p>
                  <button
                    onClick={handleClose}
                    className="text-xs text-text-secondary hover:text-text-primary border border-border rounded px-3 py-1.5 transition-colors"
                  >
                    Done
                  </button>
                </div>
              ) : (
                <>
                  {/* User search */}
                  {!selectedUser ? (
                    <div>
                      <label className="block text-xs font-medium text-text-primary mb-1.5">
                        Search for a user
                      </label>
                      <input
                        type="text"
                        value={query}
                        onChange={(e) => setQuery(e.target.value)}
                        placeholder="Username or display name..."
                        className="w-full px-3 py-2 text-sm border border-border rounded bg-surface-0 text-text-primary placeholder:text-text-secondary focus:outline-none focus:border-text-primary"
                        autoFocus
                      />
                      {searching && (
                        <p className="text-xs text-text-secondary mt-2">
                          Searching...
                        </p>
                      )}
                      {users.length > 0 && (
                        <div className="mt-2 border border-border rounded divide-y divide-border max-h-48 overflow-y-auto">
                          {users.map((u) => (
                            <button
                              key={u.user_id}
                              onClick={() => {
                                setSelectedUser(u);
                                setQuery("");
                                setUsers([]);
                              }}
                              className="w-full text-left flex items-center gap-3 px-3 py-2 hover:bg-surface-2/50 transition-colors"
                            >
                              {u.avatar_url ? (
                                <img
                                  src={u.avatar_url}
                                  alt={u.display_name ?? u.username}
                                  className="w-7 h-7 rounded-full object-cover"
                                />
                              ) : (
                                <div className="w-7 h-7 rounded-full bg-surface-2" />
                              )}
                              <div className="min-w-0">
                                <p className="text-sm font-medium text-text-primary truncate">
                                  {u.display_name ?? u.username}
                                </p>
                                {u.display_name && (
                                  <p className="text-xs text-text-secondary truncate">
                                    @{u.username}
                                  </p>
                                )}
                              </div>
                            </button>
                          ))}
                        </div>
                      )}
                      {query.trim() && !searching && users.length === 0 && (
                        <p className="text-xs text-text-secondary mt-2">
                          No users found
                        </p>
                      )}
                    </div>
                  ) : (
                    <div>
                      {/* Selected user */}
                      <div className="flex items-center gap-3 mb-4 p-2 bg-surface-2/30 rounded">
                        {selectedUser.avatar_url ? (
                          <img
                            src={selectedUser.avatar_url}
                            alt={
                              selectedUser.display_name ??
                              selectedUser.username
                            }
                            className="w-8 h-8 rounded-full object-cover"
                          />
                        ) : (
                          <div className="w-8 h-8 rounded-full bg-surface-2" />
                        )}
                        <div className="flex-1 min-w-0">
                          <p className="text-sm font-medium text-text-primary truncate">
                            {selectedUser.display_name ??
                              selectedUser.username}
                          </p>
                          {selectedUser.display_name && (
                            <p className="text-xs text-text-secondary truncate">
                              @{selectedUser.username}
                            </p>
                          )}
                        </div>
                        <button
                          onClick={() => setSelectedUser(null)}
                          className="text-xs text-text-secondary hover:text-text-primary transition-colors"
                        >
                          Change
                        </button>
                      </div>

                      {/* Note */}
                      <label className="block text-xs font-medium text-text-primary mb-1.5">
                        Add a note (optional)
                      </label>
                      <textarea
                        value={note}
                        onChange={(e) => setNote(e.target.value)}
                        placeholder="Why are you recommending this book?"
                        rows={3}
                        className="w-full px-3 py-2 text-sm border border-border rounded bg-surface-0 text-text-primary placeholder:text-text-secondary focus:outline-none focus:border-text-primary resize-none"
                      />

                      {error && (
                        <p className="text-xs text-red-500 mt-2">{error}</p>
                      )}

                      <button
                        onClick={handleSend}
                        disabled={sending}
                        className="mt-3 w-full text-sm font-medium text-surface-0 bg-text-primary rounded px-4 py-2 hover:opacity-90 transition-opacity disabled:opacity-50"
                      >
                        {sending ? "Sending..." : "Send recommendation"}
                      </button>
                    </div>
                  )}
                </>
              )}
            </div>
          </div>
        </div>
      )}
    </>
  );
}
