"use client";

import { useState } from "react";
import StarRatingInput from "@/components/star-rating-input";

type ReadingSession = {
  id: string;
  date_started: string | null;
  date_finished: string | null;
  rating: number | null;
  notes: string | null;
  created: string;
};

type Props = {
  openLibraryId: string;
  initialSessions: ReadingSession[];
};

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}

function renderStars(rating: number): string {
  return Array.from({ length: 5 }, (_, i) =>
    i < rating ? "\u2605" : "\u2606"
  ).join("");
}

export default function ReadingHistory({ openLibraryId, initialSessions }: Props) {
  const [sessions, setSessions] = useState<ReadingSession[]>(initialSessions);
  const [showForm, setShowForm] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  // Form state
  const [formDateStarted, setFormDateStarted] = useState("");
  const [formDateFinished, setFormDateFinished] = useState("");
  const [formRating, setFormRating] = useState<number | null>(null);
  const [formNotes, setFormNotes] = useState("");

  function resetForm() {
    setFormDateStarted("");
    setFormDateFinished("");
    setFormRating(null);
    setFormNotes("");
    setShowForm(false);
    setEditingId(null);
  }

  function startEdit(session: ReadingSession) {
    setEditingId(session.id);
    setFormDateStarted(session.date_started ? session.date_started.slice(0, 10) : "");
    setFormDateFinished(session.date_finished ? session.date_finished.slice(0, 10) : "");
    setFormRating(session.rating);
    setFormNotes(session.notes ?? "");
    setShowForm(true);
  }

  async function handleSave() {
    setSaving(true);

    const body: Record<string, unknown> = {};
    if (formDateStarted) body.date_started = formDateStarted;
    else body.date_started = "";
    if (formDateFinished) body.date_finished = formDateFinished;
    else body.date_finished = "";
    if (formRating != null) body.rating = formRating;
    if (formNotes) body.notes = formNotes;
    else body.notes = "";

    if (editingId) {
      // Update existing session
      const res = await fetch(`/api/me/sessions/${editingId}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      });
      if (res.ok) {
        setSessions((prev) =>
          prev.map((s) =>
            s.id === editingId
              ? {
                  ...s,
                  date_started: formDateStarted || null,
                  date_finished: formDateFinished || null,
                  rating: formRating,
                  notes: formNotes || null,
                }
              : s
          )
        );
        resetForm();
      }
    } else {
      // Create new session
      const res = await fetch(`/api/me/books/${openLibraryId}/sessions`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      });
      if (res.ok) {
        const data = await res.json();
        setSessions((prev) => [
          {
            id: data.id,
            date_started: formDateStarted || null,
            date_finished: formDateFinished || null,
            rating: formRating,
            notes: formNotes || null,
            created: new Date().toISOString(),
          },
          ...prev,
        ]);
        resetForm();
      }
    }

    setSaving(false);
  }

  async function handleDelete(sessionId: string) {
    if (!confirm("Delete this reading session?")) return;

    const res = await fetch(`/api/me/sessions/${sessionId}`, {
      method: "DELETE",
    });
    if (res.ok) {
      setSessions((prev) => prev.filter((s) => s.id !== sessionId));
      if (editingId === sessionId) resetForm();
    }
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-sm font-semibold text-text-primary uppercase tracking-wider">
          Reading History
          {sessions.length > 0 && ` (${sessions.length})`}
        </h2>
        {!showForm && (
          <button
            type="button"
            onClick={() => {
              resetForm();
              setShowForm(true);
            }}
            className="text-xs px-3 py-1.5 rounded bg-accent text-text-inverted hover:bg-accent-hover transition-colors"
          >
            Log a read
          </button>
        )}
      </div>

      {/* Add/Edit form */}
      {showForm && (
        <div className="border border-border rounded p-4 mb-4 space-y-3">
          <p className="text-xs font-medium text-text-primary">
            {editingId ? "Edit reading session" : "Log a new read"}
          </p>

          <div className="flex flex-wrap gap-4 text-xs">
            <label className="flex items-center gap-1.5 text-text-primary">
              Date started
              <input
                type="date"
                value={formDateStarted}
                onChange={(e) => setFormDateStarted(e.target.value)}
                disabled={saving}
                className="border border-border rounded px-2 py-1 text-xs text-text-primary focus:outline-none focus:ring-1 focus:ring-border-strong disabled:opacity-50"
              />
            </label>
            <label className="flex items-center gap-1.5 text-text-primary">
              Date finished
              <input
                type="date"
                value={formDateFinished}
                onChange={(e) => setFormDateFinished(e.target.value)}
                disabled={saving}
                className="border border-border rounded px-2 py-1 text-xs text-text-primary focus:outline-none focus:ring-1 focus:ring-border-strong disabled:opacity-50"
              />
            </label>
          </div>

          <div className="flex items-center gap-3">
            <span className="text-xs text-text-primary">Rating</span>
            <StarRatingInput value={formRating} onChange={setFormRating} disabled={saving} />
          </div>

          <textarea
            value={formNotes}
            onChange={(e) => setFormNotes(e.target.value)}
            disabled={saving}
            placeholder="Notes about this read (optional)"
            rows={2}
            maxLength={2000}
            className="w-full border border-border rounded px-3 py-2 text-sm text-text-primary placeholder:text-text-tertiary focus:outline-none focus:ring-1 focus:ring-border-strong resize-y disabled:opacity-50"
          />

          <div className="flex items-center gap-3">
            <button
              type="button"
              onClick={handleSave}
              disabled={saving}
              className="text-xs px-3 py-1.5 rounded bg-accent text-text-inverted hover:bg-accent-hover disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              {saving ? "Saving..." : editingId ? "Update" : "Save"}
            </button>
            <button
              type="button"
              onClick={resetForm}
              disabled={saving}
              className="text-xs text-text-primary hover:text-text-secondary transition-colors disabled:opacity-50"
            >
              Cancel
            </button>
          </div>
        </div>
      )}

      {/* Sessions list */}
      {sessions.length === 0 && !showForm ? (
        <p className="text-sm text-text-tertiary">
          No reading sessions logged yet.
        </p>
      ) : (
        <div className="space-y-3">
          {sessions.map((session) => (
            <div
              key={session.id}
              className="flex items-start justify-between gap-4 border-b border-border pb-3 last:border-0"
            >
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-3 text-xs text-text-primary">
                  {session.date_started && (
                    <span>Started {formatDate(session.date_started)}</span>
                  )}
                  {session.date_started && session.date_finished && (
                    <span className="text-text-tertiary">&rarr;</span>
                  )}
                  {session.date_finished && (
                    <span>Finished {formatDate(session.date_finished)}</span>
                  )}
                  {!session.date_started && !session.date_finished && (
                    <span className="text-text-tertiary">
                      Logged {formatDate(session.created)}
                    </span>
                  )}
                </div>
                {session.rating != null && session.rating > 0 && (
                  <span className="text-sm tracking-tight text-amber-500">
                    {renderStars(session.rating)}
                  </span>
                )}
                {session.notes && (
                  <p className="text-sm text-text-primary mt-1 line-clamp-2">
                    {session.notes}
                  </p>
                )}
              </div>

              <div className="flex items-center gap-2 shrink-0">
                <button
                  type="button"
                  onClick={() => startEdit(session)}
                  className="text-xs text-text-tertiary hover:text-text-primary transition-colors"
                >
                  Edit
                </button>
                <button
                  type="button"
                  onClick={() => handleDelete(session.id)}
                  className="text-xs text-red-400 hover:text-red-600 transition-colors"
                >
                  Delete
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
