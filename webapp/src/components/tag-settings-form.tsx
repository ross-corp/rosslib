"use client";

import { useState } from "react";

type TagValue = {
  id: string;
  name: string;
  slug: string;
};

type TagKey = {
  id: string;
  name: string;
  slug: string;
  mode: "select_one" | "select_multiple";
  values: TagValue[];
};

export default function TagSettingsForm({
  initialTagKeys,
}: {
  initialTagKeys: TagKey[];
}) {
  const [tagKeys, setTagKeys] = useState<TagKey[]>(initialTagKeys);

  const [newKeyName, setNewKeyName] = useState("");
  const [newKeyMode, setNewKeyMode] = useState<"select_one" | "select_multiple">("select_one");
  const [addingKey, setAddingKey] = useState(false);

  const [newValueInputs, setNewValueInputs] = useState<Record<string, string>>({});
  const [addingValue, setAddingValue] = useState<string | null>(null);

  async function createKey(e: React.FormEvent) {
    e.preventDefault();
    if (!newKeyName.trim()) return;
    setAddingKey(true);
    const res = await fetch("/api/me/tag-keys", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name: newKeyName.trim(), mode: newKeyMode }),
    });
    setAddingKey(false);
    if (res.ok) {
      const key: TagKey = await res.json();
      key.values = key.values ?? [];
      setTagKeys((prev) => [...prev, key]);
      setNewKeyName("");
    }
  }

  async function deleteKey(keyId: string) {
    const res = await fetch(`/api/me/tag-keys/${keyId}`, { method: "DELETE" });
    if (res.ok) {
      setTagKeys((prev) => prev.filter((k) => k.id !== keyId));
    }
  }

  async function createValue(keyId: string) {
    const name = (newValueInputs[keyId] ?? "").trim();
    if (!name) return;
    setAddingValue(keyId);
    const res = await fetch(`/api/me/tag-keys/${keyId}/values`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name }),
    });
    setAddingValue(null);
    if (res.ok) {
      const val: TagValue = await res.json();
      setTagKeys((prev) =>
        prev.map((k) =>
          k.id === keyId ? { ...k, values: [...k.values, val] } : k
        )
      );
      setNewValueInputs((prev) => ({ ...prev, [keyId]: "" }));
    }
  }

  async function deleteValue(keyId: string, valueId: string) {
    const res = await fetch(`/api/me/tag-keys/${keyId}/values/${valueId}`, {
      method: "DELETE",
    });
    if (res.ok) {
      setTagKeys((prev) =>
        prev.map((k) =>
          k.id === keyId
            ? { ...k, values: k.values.filter((v) => v.id !== valueId) }
            : k
        )
      );
    }
  }

  return (
    <div className="space-y-6">
      {tagKeys.map((key) => (
        <div key={key.id} className="border border-border rounded-lg p-4">
          <div className="flex items-center justify-between mb-3">
            <div>
              <h2 className="font-semibold text-text-primary">{key.name}</h2>
              <p className="text-xs text-text-primary mt-0.5">
                {key.mode === "select_multiple" ? "Select multiple" : "Select one"}
              </p>
            </div>
            <button
              onClick={() => deleteKey(key.id)}
              className="text-xs text-text-primary hover:text-red-500 transition-colors"
            >
              Delete
            </button>
          </div>

          <div className="flex flex-wrap gap-2 mb-3">
            {key.values.map((val) => (
              <span
                key={val.id}
                className="inline-flex items-center gap-1 text-sm px-2.5 py-0.5 rounded-full border border-border text-text-primary"
              >
                {val.name}
                <button
                  onClick={() => deleteValue(key.id, val.id)}
                  className="text-text-primary hover:text-red-400 transition-colors leading-none"
                  aria-label={`Remove ${val.name}`}
                >
                  ×
                </button>
              </span>
            ))}
            {key.values.length === 0 && (
              <span className="text-sm text-text-primary">No values yet</span>
            )}
          </div>

          <form
            onSubmit={(e) => { e.preventDefault(); createValue(key.id); }}
            className="flex gap-2"
          >
            <input
              type="text"
              placeholder="Add value..."
              value={newValueInputs[key.id] ?? ""}
              onChange={(e) =>
                setNewValueInputs((prev) => ({ ...prev, [key.id]: e.target.value }))
              }
              className="flex-1 text-sm border border-border rounded px-2.5 py-1.5 focus:outline-none focus:border-border"
            />
            <button
              type="submit"
              disabled={addingValue === key.id || !(newValueInputs[key.id] ?? "").trim()}
              className="text-sm px-3 py-1.5 rounded border border-border text-text-primary hover:border-border-strong hover:text-text-primary transition-colors disabled:opacity-50"
            >
              {addingValue === key.id ? "Adding..." : "Add"}
            </button>
          </form>
        </div>
      ))}

      <div className="border border-border border-dashed rounded-lg p-4">
        <p className="text-xs font-semibold text-text-primary uppercase tracking-wider mb-3">
          New label category
        </p>
        <form onSubmit={createKey} className="space-y-3">
          <input
            type="text"
            placeholder="Name (e.g. Gifted from, Read in)"
            value={newKeyName}
            onChange={(e) => setNewKeyName(e.target.value)}
            className="w-full text-sm border border-border rounded px-2.5 py-1.5 focus:outline-none focus:border-border"
          />
          <div className="flex gap-3">
            <label className="flex items-center gap-1.5 text-sm text-text-primary cursor-pointer">
              <input
                type="radio"
                name="mode"
                value="select_one"
                checked={newKeyMode === "select_one"}
                onChange={() => setNewKeyMode("select_one")}
              />
              Select one
            </label>
            <label className="flex items-center gap-1.5 text-sm text-text-primary cursor-pointer">
              <input
                type="radio"
                name="mode"
                value="select_multiple"
                checked={newKeyMode === "select_multiple"}
                onChange={() => setNewKeyMode("select_multiple")}
              />
              Select multiple
            </label>
          </div>
          <button
            type="submit"
            disabled={addingKey || !newKeyName.trim()}
            className="text-sm px-3 py-1.5 rounded border border-accent bg-accent text-white hover:bg-surface-3 transition-colors disabled:opacity-50"
          >
            {addingKey ? "Creating..." : "Create category"}
          </button>
        </form>
      </div>
    </div>
  );
}
