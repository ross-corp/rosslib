"use client";

import { useEffect, useRef, useState } from "react";

export type TagValue = {
  id: string;
  name: string;
  slug: string;
};

export type TagKey = {
  id: string;
  name: string;
  slug: string;
  mode: "select_one" | "select_multiple";
  values: TagValue[];
};

type BookTag = {
  key_id: string;
  value_id: string;
};

export default function BookTagPicker({
  openLibraryId,
  tagKeys,
}: {
  openLibraryId: string;
  tagKeys: TagKey[];
}) {
  const [open, setOpen] = useState(false);
  // keyId → set of assigned value IDs
  const [assignments, setAssignments] = useState<Record<string, Set<string>>>({});
  const [loaded, setLoaded] = useState(false);
  const [saving, setSaving] = useState<string | null>(null); // valueId being toggled
  // Per-key free-form input state
  const [freeInput, setFreeInput] = useState<Record<string, string>>({});
  const [addingFree, setAddingFree] = useState<string | null>(null); // keyId
  const containerRef = useRef<HTMLDivElement>(null);

  // Lazily load current tags on first open.
  useEffect(() => {
    if (!open || loaded) return;
    fetch(`/api/me/books/${openLibraryId}/tags`)
      .then((r) => (r.ok ? r.json() : []))
      .then((data: BookTag[]) => {
        const map: Record<string, Set<string>> = {};
        for (const t of data) {
          if (!map[t.key_id]) map[t.key_id] = new Set();
          map[t.key_id].add(t.value_id);
        }
        setAssignments(map);
        setLoaded(true);
      });
  }, [open, loaded, openLibraryId]);

  // Close on outside click.
  useEffect(() => {
    if (!open) return;
    function handle(e: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    }
    document.addEventListener("mousedown", handle);
    return () => document.removeEventListener("mousedown", handle);
  }, [open]);

  function isAssigned(keyId: string, valueId: string) {
    return assignments[keyId]?.has(valueId) ?? false;
  }

  async function toggle(key: TagKey, valueId: string) {
    setSaving(valueId);
    const assigned = isAssigned(key.id, valueId);

    if (assigned) {
      if (key.mode === "select_one") {
        // Clear the entire key.
        await fetch(`/api/me/books/${openLibraryId}/tags/${key.id}`, { method: "DELETE" });
        setAssignments((prev) => ({ ...prev, [key.id]: new Set() }));
      } else {
        // Remove just this value.
        await fetch(`/api/me/books/${openLibraryId}/tags/${key.id}/values/${valueId}`, {
          method: "DELETE",
        });
        setAssignments((prev) => {
          const next = new Set(prev[key.id] ?? []);
          next.delete(valueId);
          return { ...prev, [key.id]: next };
        });
      }
    } else {
      await fetch(`/api/me/books/${openLibraryId}/tags/${key.id}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ value_id: valueId }),
      });
      setAssignments((prev) => {
        const existing = key.mode === "select_one" ? new Set<string>() : new Set(prev[key.id] ?? []);
        existing.add(valueId);
        return { ...prev, [key.id]: existing };
      });
    }

    setSaving(null);
  }

  async function addFreeValue(key: TagKey) {
    const name = (freeInput[key.id] ?? "").trim();
    if (!name) return;
    setAddingFree(key.id);

    const res = await fetch(`/api/me/books/${openLibraryId}/tags/${key.id}`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ value_name: name }),
    });

    setAddingFree(null);
    if (res.ok) {
      // Refresh tag keys to pick up the new value — simplest approach is to
      // reload. We don't have a hook to update tagKeys here, so we just update
      // the assignment optimistically with the value_name as a temporary label.
      // A full reload would require lifting state; for now we just note it's set.
      setFreeInput((prev) => ({ ...prev, [key.id]: "" }));
      // Reload assignments so the new value appears as checked.
      const updated = await fetch(`/api/me/books/${openLibraryId}/tags`).then((r) =>
        r.ok ? r.json() : []
      );
      const map: Record<string, Set<string>> = {};
      for (const t of updated as BookTag[]) {
        if (!map[t.key_id]) map[t.key_id] = new Set();
        map[t.key_id].add(t.value_id);
      }
      setAssignments(map);
    }
  }

  if (tagKeys.length === 0) return null;

  const totalSet = Object.values(assignments).reduce((sum, s) => sum + s.size, 0);

  return (
    <div ref={containerRef} className="relative">
      <button
        onClick={() => setOpen((v) => !v)}
        className={`text-[10px] px-2 py-0.5 rounded border transition-colors ${
          totalSet > 0
            ? "border-stone-400 text-stone-600"
            : "border-stone-200 text-stone-400 hover:border-stone-400 hover:text-stone-600"
        }`}
      >
        {totalSet > 0 ? `${totalSet} label${totalSet > 1 ? "s" : ""}` : "label"}
      </button>

      {open && (
        <div className="absolute left-0 top-full mt-1 z-20 bg-white border border-stone-200 rounded shadow-md min-w-[200px]">
          {!loaded ? (
            <div className="px-3 py-2 text-xs text-stone-400">Loading...</div>
          ) : (
            tagKeys.map((key) => (
              <div key={key.id} className="border-b border-stone-100 last:border-0">
                <div className="px-3 pt-2 pb-1 flex items-center gap-1.5">
                  <span className="text-[10px] font-semibold text-stone-400 uppercase tracking-wider">
                    {key.name}
                  </span>
                  <span className="text-[9px] text-stone-300">
                    {key.mode === "select_multiple" ? "multi" : "single"}
                  </span>
                </div>

                <div className="pb-1">
                  {key.values.map((val) => {
                    const checked = isAssigned(key.id, val.id);
                    const busy = saving === val.id || addingFree === key.id;
                    return (
                      <button
                        key={val.id}
                        onClick={() => toggle(key, val.id)}
                        disabled={busy}
                        className={`w-full text-left px-3 py-1.5 text-xs transition-colors hover:bg-stone-50 disabled:opacity-50 ${
                          checked ? "text-stone-900 font-medium" : "text-stone-600"
                        }`}
                      >
                        <span className="inline-block w-3">
                          {key.mode === "select_multiple"
                            ? checked ? "☑" : "☐"
                            : checked ? "✓" : " "}
                        </span>{" "}
                        {val.name}
                      </button>
                    );
                  })}
                  {key.values.length === 0 && (
                    <p className="px-3 py-1 text-xs text-stone-400">No values defined</p>
                  )}
                </div>

                {/* Free-form input */}
                <div className="px-2 pb-2">
                  <form
                    onSubmit={(e) => { e.preventDefault(); addFreeValue(key); }}
                    className="flex gap-1"
                  >
                    <input
                      type="text"
                      placeholder="Add value..."
                      value={freeInput[key.id] ?? ""}
                      onChange={(e) =>
                        setFreeInput((prev) => ({ ...prev, [key.id]: e.target.value }))
                      }
                      className="flex-1 text-[10px] border border-stone-200 rounded px-1.5 py-1 focus:outline-none focus:border-stone-400 min-w-0"
                    />
                    <button
                      type="submit"
                      disabled={addingFree === key.id || !(freeInput[key.id] ?? "").trim()}
                      className="text-[10px] px-1.5 py-1 rounded border border-stone-200 text-stone-500 hover:border-stone-400 disabled:opacity-40 shrink-0"
                    >
                      {addingFree === key.id ? "..." : "+"}
                    </button>
                  </form>
                </div>
              </div>
            ))
          )}
        </div>
      )}
    </div>
  );
}
