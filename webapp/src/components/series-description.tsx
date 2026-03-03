"use client";

import { useState } from "react";

export default function SeriesDescription({
  seriesId,
  initialName,
  initialDescription,
  isLoggedIn,
}: {
  seriesId: string;
  initialName: string;
  initialDescription: string;
  isLoggedIn: boolean;
}) {
  const [name, setName] = useState(initialName);
  const [description, setDescription] = useState(initialDescription);
  const [editing, setEditing] = useState(false);
  const [draftName, setDraftName] = useState(initialName);
  const [draftDescription, setDraftDescription] = useState(initialDescription);
  const [saving, setSaving] = useState(false);

  async function handleSave() {
    const trimmedName = draftName.trim();
    if (!trimmedName) return;
    setSaving(true);
    try {
      const body: Record<string, string> = {};
      if (trimmedName !== name) body.name = trimmedName;
      if (draftDescription !== description) body.description = draftDescription;
      if (Object.keys(body).length === 0) {
        setEditing(false);
        return;
      }
      const res = await fetch(`/api/series/${seriesId}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      });
      if (res.ok) {
        const data = await res.json();
        setName(data.name ?? name);
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
        <input
          type="text"
          value={draftName}
          onChange={(e) => setDraftName(e.target.value)}
          className="w-full bg-surface-2 border border-border rounded-lg px-3 py-2 text-xl font-bold text-text-primary placeholder-text-tertiary focus:outline-none focus:ring-1 focus:ring-text-tertiary mb-3"
          placeholder="Series name"
          autoFocus
        />
        <textarea
          value={draftDescription}
          onChange={(e) => setDraftDescription(e.target.value)}
          rows={3}
          className="w-full bg-surface-2 border border-border rounded-lg px-3 py-2 text-sm text-text-primary placeholder-text-tertiary focus:outline-none focus:ring-1 focus:ring-text-tertiary resize-vertical"
          placeholder="Add a series description..."
        />
        <div className="flex gap-2 mt-2">
          <button
            onClick={handleSave}
            disabled={saving || !draftName.trim()}
            className="text-xs font-medium px-3 py-1.5 rounded bg-text-primary text-bg-primary hover:opacity-90 transition-opacity disabled:opacity-50"
          >
            {saving ? "Saving..." : "Save"}
          </button>
          <button
            onClick={() => {
              setDraftName(name);
              setDraftDescription(description);
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
      <h1 className="text-2xl font-bold text-text-primary mb-2">{name}</h1>
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
            setDraftName(name);
            setDraftDescription(description);
            setEditing(true);
          }}
          className="text-xs text-text-tertiary hover:text-text-secondary mt-1 transition-colors"
        >
          Edit series
        </button>
      )}
    </div>
  );
}
