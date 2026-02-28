"use client";

import { useState } from "react";

type Tab = "bug" | "feature";

export default function FeedbackForm() {
  const [tab, setTab] = useState<Tab>("bug");
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [steps, setSteps] = useState("");
  const [severity, setSeverity] = useState("medium");
  const [submitting, setSubmitting] = useState(false);
  const [success, setSuccess] = useState(false);
  const [error, setError] = useState("");

  function reset() {
    setTitle("");
    setDescription("");
    setSteps("");
    setSeverity("medium");
    setError("");
  }

  function switchTab(t: Tab) {
    setTab(t);
    reset();
    setSuccess(false);
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setSubmitting(true);
    setError("");
    setSuccess(false);

    const body: Record<string, string> = {
      type: tab,
      title,
      description,
    };
    if (tab === "bug") {
      body.steps_to_reproduce = steps;
      body.severity = severity;
    }

    try {
      const res = await fetch("/api/feedback", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      });
      if (!res.ok) {
        const data = await res.json();
        setError(data.error || "Something went wrong");
        return;
      }
      setSuccess(true);
      reset();
    } catch {
      setError("Network error. Please try again.");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div>
      <div className="flex gap-1 mb-6">
        <button
          onClick={() => switchTab("bug")}
          className={`px-4 py-2 text-sm font-medium rounded transition-colors ${
            tab === "bug"
              ? "bg-accent text-text-inverted"
              : "bg-surface-2 text-text-secondary hover:text-text-primary"
          }`}
        >
          Bug Report
        </button>
        <button
          onClick={() => switchTab("feature")}
          className={`px-4 py-2 text-sm font-medium rounded transition-colors ${
            tab === "feature"
              ? "bg-accent text-text-inverted"
              : "bg-surface-2 text-text-secondary hover:text-text-primary"
          }`}
        >
          Feature Request
        </button>
      </div>

      {success && (
        <div className="mb-4 p-3 rounded bg-semantic-success-bg border border-semantic-success-border text-semantic-success text-sm">
          Thank you! Your {tab === "bug" ? "bug report" : "feature request"} has
          been submitted.
        </div>
      )}

      {error && (
        <div className="mb-4 p-3 rounded bg-semantic-error-bg border border-semantic-error-border text-semantic-error text-sm">
          {error}
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-text-primary mb-1">
            Title
          </label>
          <input
            type="text"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            required
            className="w-full px-3 py-2 text-sm bg-surface-2 border border-border rounded text-text-primary placeholder-text-tertiary focus:outline-none focus:ring-1 focus:ring-accent"
            placeholder={
              tab === "bug"
                ? "Brief summary of the bug"
                : "Brief summary of the feature"
            }
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-text-primary mb-1">
            Description
          </label>
          <textarea
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            required
            rows={4}
            className="w-full px-3 py-2 text-sm bg-surface-2 border border-border rounded text-text-primary placeholder-text-tertiary focus:outline-none focus:ring-1 focus:ring-accent resize-y"
            placeholder={
              tab === "bug"
                ? "What happened? What did you expect?"
                : "Describe the feature you'd like to see"
            }
          />
        </div>

        {tab === "bug" && (
          <>
            <div>
              <label className="block text-sm font-medium text-text-primary mb-1">
                Steps to Reproduce
              </label>
              <textarea
                value={steps}
                onChange={(e) => setSteps(e.target.value)}
                rows={3}
                className="w-full px-3 py-2 text-sm bg-surface-2 border border-border rounded text-text-primary placeholder-text-tertiary focus:outline-none focus:ring-1 focus:ring-accent resize-y"
                placeholder="1. Go to...&#10;2. Click on...&#10;3. See error"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-text-primary mb-1">
                Severity
              </label>
              <select
                value={severity}
                onChange={(e) => setSeverity(e.target.value)}
                className="px-3 py-2 text-sm bg-surface-2 border border-border rounded text-text-primary focus:outline-none focus:ring-1 focus:ring-accent"
              >
                <option value="low">Low</option>
                <option value="medium">Medium</option>
                <option value="high">High</option>
              </select>
            </div>
          </>
        )}

        <button
          type="submit"
          disabled={submitting}
          className="btn-primary"
        >
          {submitting
            ? "Submitting..."
            : tab === "bug"
              ? "Submit Bug Report"
              : "Submit Feature Request"}
        </button>
      </form>
    </div>
  );
}
