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
      <div role="tablist" className="flex gap-1 border-b border-border mb-8">
        <button
          role="tab"
          id="tab-my-lists"
          aria-selected={tab === "my-lists"}
          aria-controls="tabpanel-my-lists"
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
          role="tab"
          id="tab-friend"
          aria-selected={tab === "friend"}
          aria-controls="tabpanel-friend"
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

      <div
        role="tabpanel"
        id={tab === "my-lists" ? "tabpanel-my-lists" : "tabpanel-friend"}
        aria-labelledby={tab === "my-lists" ? "tab-my-lists" : "tab-friend"}
      >
        {tab === "my-lists" ? myListsContent : friendContent}
      </div>
    </div>
  );
}
