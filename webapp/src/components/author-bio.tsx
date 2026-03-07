"use client";

import { useState } from "react";

const MAX_LENGTH = 500;

export default function AuthorBio({ bio }: { bio: string }) {
  const needsTruncation = bio.length > MAX_LENGTH;
  const [expanded, setExpanded] = useState(false);

  const displayText =
    needsTruncation && !expanded ? bio.slice(0, MAX_LENGTH) + "\u2026" : bio;

  return (
    <div className="text-text-primary text-sm leading-relaxed">
      <p className="whitespace-pre-wrap">{displayText}</p>
      {needsTruncation && (
        <button
          onClick={() => setExpanded(!expanded)}
          className="text-accent hover:text-accent-hover text-sm mt-1 transition-colors"
        >
          {expanded ? "Show less" : "Read more"}
        </button>
      )}
    </div>
  );
}
