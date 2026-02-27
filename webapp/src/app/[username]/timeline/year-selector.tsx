"use client";

import { useRouter } from "next/navigation";

export default function TimelineYearSelector({
  username,
  currentYear,
}: {
  username: string;
  currentYear: number;
}) {
  const router = useRouter();
  const thisYear = new Date().getFullYear();

  // Show years from 5 years ago to current year
  const years: number[] = [];
  for (let y = thisYear; y >= thisYear - 5; y--) {
    years.push(y);
  }

  return (
    <select
      value={currentYear}
      onChange={(e) => {
        const year = e.target.value;
        router.push(`/${username}/timeline?year=${year}`);
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
