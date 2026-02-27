"use client";

import { useState } from "react";

type GoalData = {
  id: string;
  year: number;
  target: number;
};

export default function ReadingGoalForm({
  initialGoal,
}: {
  initialGoal: GoalData | null;
}) {
  const currentYear = new Date().getFullYear();
  const [target, setTarget] = useState(
    initialGoal?.target?.toString() ?? ""
  );
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState("");

  async function handleSubmit(ev: React.FormEvent) {
    ev.preventDefault();
    const num = parseInt(target, 10);
    if (!num || num < 1) {
      setMessage("Enter a number of at least 1.");
      return;
    }

    setSaving(true);
    setMessage("");

    const res = await fetch(`/api/me/goals/${currentYear}`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ target: num }),
    });

    setSaving(false);

    if (!res.ok) {
      const data = await res.json();
      setMessage(data.error || "Something went wrong.");
      return;
    }

    setMessage("Goal saved!");
  }

  return (
    <div className="border-t border-border pt-8 mt-8">
      <h2 className="text-lg font-bold text-text-primary mb-1">
        Reading Goal
      </h2>
      <p className="text-sm text-text-secondary mb-4">
        Set a target number of books to read in {currentYear}.
      </p>

      <form onSubmit={handleSubmit} className="flex items-end gap-3 max-w-sm">
        <div className="flex-1">
          <label
            htmlFor="goal-target"
            className="block text-sm text-text-secondary mb-1"
          >
            Books to read in {currentYear}
          </label>
          <input
            id="goal-target"
            type="number"
            min={1}
            value={target}
            onChange={(e) => setTarget(e.target.value)}
            placeholder="e.g. 25"
            className="w-full px-3 py-2 border border-border rounded text-text-primary placeholder-text-tertiary focus:outline-none focus:ring-2 focus:ring-accent focus:border-transparent text-sm"
          />
        </div>
        <button
          type="submit"
          disabled={saving}
          className="btn-primary disabled:opacity-50"
        >
          {saving ? "Saving..." : "Save"}
        </button>
      </form>

      {message && (
        <p
          className={`text-sm mt-3 ${message.includes("saved") ? "text-green-500" : "text-red-500"}`}
        >
          {message}
        </p>
      )}
    </div>
  );
}
