"use client";

import { useState } from "react";
import ImportForm from "@/components/import-form";

const tabs = [
  { key: "goodreads" as const, label: "Goodreads" },
  { key: "storygraph" as const, label: "StoryGraph" },
];

export default function ImportTabs({ username }: { username: string }) {
  const [active, setActive] = useState<"goodreads" | "storygraph">("goodreads");

  return (
    <div>
      <div className="flex gap-1 mb-8 border-b border-border">
        {tabs.map((tab) => (
          <button
            key={tab.key}
            type="button"
            onClick={() => setActive(tab.key)}
            className={`px-4 py-2 text-sm font-medium transition-colors -mb-px ${
              active === tab.key
                ? "border-b-2 border-text-primary text-text-primary"
                : "text-text-primary hover:text-text-primary"
            }`}
          >
            {tab.label}
          </button>
        ))}
      </div>

      <ImportForm key={active} username={username} source={active} />
    </div>
  );
}
