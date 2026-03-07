"use client";

import { useRouter } from "next/navigation";

const SORT_OPTIONS = [
  { value: "newest", label: "Newest first" },
  { value: "oldest", label: "Oldest first" },
  { value: "highest_rating", label: "Highest rating" },
  { value: "lowest_rating", label: "Lowest rating" },
];

export default function ReviewSortDropdown({
  username,
  currentSort,
}: {
  username: string;
  currentSort: string;
}) {
  const router = useRouter();

  function handleChange(e: React.ChangeEvent<HTMLSelectElement>) {
    const sort = e.target.value;
    const params = sort !== "newest" ? `?sort=${sort}` : "";
    router.push(`/${username}/reviews${params}`);
  }

  return (
    <select
      value={currentSort}
      onChange={handleChange}
      className="text-xs border border-border rounded px-2 py-1.5 text-text-primary bg-surface-0 focus:outline-none focus:ring-1 focus:ring-border-strong"
    >
      {SORT_OPTIONS.map((opt) => (
        <option key={opt.value} value={opt.value}>
          {opt.label}
        </option>
      ))}
    </select>
  );
}
