"use client";

export type ReadingGoal = {
  id: string;
  year: number;
  target: number;
  progress: number;
};

export default function ReadingGoalCard({ goal }: { goal: ReadingGoal }) {
  const pct = Math.min(100, Math.round((goal.progress / goal.target) * 100));
  const complete = goal.progress >= goal.target;

  return (
    <div className="p-4 rounded border border-border">
      <div className="flex items-baseline justify-between mb-2">
        <h3 className="text-sm font-medium text-text-primary">
          {goal.year} Reading Goal
        </h3>
        <span className="text-xs text-text-tertiary">{pct}%</span>
      </div>
      <div className="w-full h-2 rounded-full bg-surface-2 overflow-hidden mb-2">
        <div
          className={`h-full rounded-full transition-all ${complete ? "bg-progress" : "bg-accent"}`}
          style={{ width: `${pct}%` }}
        />
      </div>
      <p className="text-sm text-text-secondary">
        <span className="font-semibold text-text-primary">{goal.progress}</span>{" "}
        of{" "}
        <span className="font-semibold text-text-primary">{goal.target}</span>{" "}
        books read in {goal.year}
      </p>
    </div>
  );
}
