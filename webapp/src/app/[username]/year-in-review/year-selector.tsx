"use client";

import { useRouter } from "next/navigation";

export default function YearInReviewSelector({
  username,
  currentYear,
  availableYears,
}: {
  username: string;
  currentYear: number;
  availableYears: number[];
}) {
  const router = useRouter();
  const thisYear = new Date().getFullYear();

  // Merge available years with recent years, dedupe, sort descending
  const yearSet = new Set(availableYears);
  yearSet.add(thisYear);
  const years = Array.from(yearSet).sort((a, b) => b - a);

  return (
    <select
      value={currentYear}
      onChange={(e) => {
        const year = e.target.value;
        router.push(`/${username}/year-in-review?year=${year}`);
      }}
      className="bg-surface-1 border border-border rounded px-2 py-1 text-sm text-text-primary focus:outline-none focus:border-accent"
    >
      {years.map((y) => (
        <option key={y} value={y}>
          {y}
        </option>
      ))}
    </select>
  );
}
