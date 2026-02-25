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
      <div className="flex gap-1 border-b border-stone-200 mb-8">
        <button
          onClick={() => setTab("my-lists")}
          className={`px-4 py-2 text-sm font-medium transition-colors ${
            tab === "my-lists"
              ? "text-stone-900 border-b-2 border-stone-900"
              : "text-stone-400 hover:text-stone-600"
          }`}
        >
          My Lists
        </button>
        <button
          onClick={() => setTab("friend")}
          className={`px-4 py-2 text-sm font-medium transition-colors ${
            tab === "friend"
              ? "text-stone-900 border-b-2 border-stone-900"
              : "text-stone-400 hover:text-stone-600"
          }`}
        >
          Compare with a Friend
        </button>
      </div>

      {tab === "my-lists" ? myListsContent : friendContent}
    </div>
  );
}
