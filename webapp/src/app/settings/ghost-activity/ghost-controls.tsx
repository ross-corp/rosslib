"use client";

import { useState } from "react";
import Link from "next/link";

type GhostStatus = {
  username: string;
  display_name: string;
  user_id: string;
  books_read: number;
  currently_reading: number;
  want_to_read: number;
  following_count: number;
  followers_count: number;
};

type SimulateResult = {
  ghost: string;
  actions: string[];
};

export default function GhostControls({
  initialGhosts,
}: {
  initialGhosts: GhostStatus[];
}) {
  const [ghosts, setGhosts] = useState(initialGhosts);
  const [seeding, setSeeding] = useState(false);
  const [simulating, setSimulating] = useState(false);
  const [seedMessage, setSeedMessage] = useState("");
  const [simulateResults, setSimulateResults] = useState<SimulateResult[]>([]);

  async function seed() {
    setSeeding(true);
    setSeedMessage("");
    try {
      const res = await fetch("/api/admin/ghosts/seed", { method: "POST" });
      const data = await res.json();
      if (res.ok) {
        setSeedMessage(`Seeded: ${data.created?.join(", ") || "done"}`);
        await refreshStatus();
      } else {
        setSeedMessage(`Error: ${data.error || "failed"}`);
      }
    } catch {
      setSeedMessage("Error: network failure");
    }
    setSeeding(false);
  }

  async function simulate() {
    setSimulating(true);
    setSimulateResults([]);
    try {
      const res = await fetch("/api/admin/ghosts/simulate", { method: "POST" });
      const data = await res.json();
      if (res.ok && data.results) {
        setSimulateResults(data.results);
        await refreshStatus();
      }
    } catch {
      // ignore
    }
    setSimulating(false);
  }

  async function refreshStatus() {
    try {
      const res = await fetch("/api/admin/ghosts/status");
      if (res.ok) {
        const data = await res.json();
        setGhosts(data);
      }
    } catch {
      // ignore
    }
  }

  return (
    <div className="space-y-8">
      <div className="flex items-center gap-3">
        <button
          onClick={seed}
          disabled={seeding}
          className="text-sm px-4 py-2 rounded border border-accent bg-accent text-white hover:bg-surface-3 transition-colors disabled:opacity-50"
        >
          {seeding ? "Seeding..." : "Seed ghosts"}
        </button>
        <button
          onClick={simulate}
          disabled={simulating || ghosts.length === 0}
          className="text-sm px-4 py-2 rounded border border-border text-text-primary hover:border-border hover:text-text-primary transition-colors disabled:opacity-50"
        >
          {simulating ? "Simulating..." : "Simulate round"}
        </button>
      </div>

      {seedMessage && (
        <p className="text-sm text-text-primary">{seedMessage}</p>
      )}

      {ghosts.length === 0 ? (
        <p className="text-sm text-text-primary">
          No ghosts seeded yet. Click &quot;Seed ghosts&quot; to create them.
        </p>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          {ghosts.map((g) => (
            <div
              key={g.username}
              className="border border-border rounded-lg p-4 space-y-2"
            >
              <div className="flex items-center justify-between">
                <Link
                  href={`/${g.username}`}
                  className="font-medium text-text-primary hover:underline"
                >
                  {g.display_name}
                </Link>
                <span className="text-xs text-text-primary">@{g.username}</span>
              </div>
              <div className="grid grid-cols-3 gap-2 text-sm text-text-primary">
                <div>
                  <span className="font-medium text-text-primary">{g.books_read}</span> read
                </div>
                <div>
                  <span className="font-medium text-text-primary">{g.currently_reading}</span> reading
                </div>
                <div>
                  <span className="font-medium text-text-primary">{g.want_to_read}</span> want
                </div>
              </div>
              <div className="text-xs text-text-primary">
                {g.following_count} following &middot; {g.followers_count} followers
              </div>
            </div>
          ))}
        </div>
      )}

      {simulateResults.length > 0 && (
        <div className="space-y-4">
          <h2 className="text-lg font-semibold text-text-primary">Simulation results</h2>
          {simulateResults.map((r) => (
            <div key={r.ghost} className="border border-border rounded-lg p-3">
              <p className="text-sm font-medium text-text-primary mb-1">@{r.ghost}</p>
              {r.actions.length === 0 ? (
                <p className="text-xs text-text-primary">No actions taken</p>
              ) : (
                <ul className="text-xs text-text-primary space-y-0.5">
                  {r.actions.map((a, i) => (
                    <li key={i}>{a}</li>
                  ))}
                </ul>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
