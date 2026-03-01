"use client";

import { useState } from "react";

export default function SeriesDescription({
  seriesId,
  initialDescription,
  isLoggedIn,
}: {
  seriesId: string;
  initialDescription: string;
  isLoggedIn: boolean;
}) {
  const [description, setDescription] = useState(initialDescription);
  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState(initialDescription);
  const [saving, setSaving] = useState(false);

  async function handleSave() {
    setSaving(true);
    try {
      const res = await fetch(`/api/series/${seriesId}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ description: draft }),
      });
      if (res.ok) {
        const data = await res.json();
        setDescription(data.description ?? "");
        setEditing(false);
      }
    } finally {
      setSaving(false);
    }
  }

  if (editing) {
    return (
      <div className="mb-6">
        <textarea
          value={draft}
          onChange={(e) => setDraft(e.target.value)}
          rows={3}
          className="w-full bg-surface-2 border border-border rounded-lg px-3 py-2 text-sm text-text-primary placeholder-text-tertiary focus:outline-none focus:ring-1 focus:ring-text-tertiary resize-vertical"
          placeholder="Add a series description..."
          autoFocus
        />
        <div className="flex gap-2 mt-2">
          <button
            onClick={handleSave}
            disabled={saving}
            className="text-xs font-medium px-3 py-1.5 rounded bg-text-primary text-bg-primary hover:opacity-90 transition-opacity disabled:opacity-50"
          >
            {saving ? "Saving..." : "Save"}
          </button>
          <button
            onClick={() => {
              setDraft(description);
              setEditing(false);
            }}
            className="text-xs font-medium px-3 py-1.5 rounded text-text-tertiary hover:text-text-secondary transition-colors"
          >
            Cancel
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="mb-6">
      {description ? (
        <p className="text-sm text-text-secondary leading-relaxed">
          {description}
        </p>
      ) : isLoggedIn ? (
        <p className="text-sm text-text-tertiary italic">No description yet.</p>
      ) : null}
      {isLoggedIn && (
        <button
          onClick={() => {
            setDraft(description);
            setEditing(true);
          }}
          className="text-xs text-text-tertiary hover:text-text-secondary mt-1 transition-colors"
        >
          {description ? "Edit description" : "Add description"}
        </button>
      )}
    </div>
  );
}
