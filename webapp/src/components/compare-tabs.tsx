"use client";

import { useState, type ReactNode } from "react";

export default function CompareTabs({
  myListsContent,
  friendContent,
}: {
  myListsContent: ReactNode;
  friendContent: ReactNode;
}) {
  const [tab, setTab] = useState<"my-lists" | "friend">("my-lists");

  return (
    <div>
      <div className="flex gap-1 border-b border-border mb-8">
        <button
          onClick={() => setTab("my-lists")}
          className={`px-4 py-2 text-sm font-medium transition-colors ${
            tab === "my-lists"
              ? "text-text-primary border-b-2 border-accent"
              : "text-text-primary hover:text-text-primary"
          }`}
        >
          My Lists
        </button>
        <button
          onClick={() => setTab("friend")}
          className={`px-4 py-2 text-sm font-medium transition-colors ${
            tab === "friend"
              ? "text-text-primary border-b-2 border-accent"
              : "text-text-primary hover:text-text-primary"
          }`}
        >
          Compare with a Friend
        </button>
      </div>

      {tab === "my-lists" ? myListsContent : friendContent}
    </div>
  );
}
