import Link from "next/link";
import React from "react";

type Props = {
  text: string;
};

export default function ReviewText({ text }: Props) {
  if (!text) return null;

  // Split by [[Wikilink]] or [Markdown](Link)
  // use non-greedy matching for content inside brackets/parens
  // The capturing group means the delimiters are included in the result array
  const parts = text.split(/(\[\[.+?\]\]|\[.+?\]\(.+?\))/g);

  return (
    <span className="whitespace-pre-wrap">
      {parts.map((part, i) => {
        // Handle [[Title]] -> Search
        if (part.startsWith("[[") && part.endsWith("]]")) {
          const content = part.slice(2, -2);
          return (
            <Link
              key={i}
              href={`/search?q=${encodeURIComponent(content)}`}
              className="text-blue-600 hover:underline"
            >
              {content}
            </Link>
          );
        }

        // Handle [Title](URL)
        // Check for specific structure to be safe
        const match = part.match(/^\[(.+?)\]\((.+?)\)$/);
        if (match) {
          const [, label, url] = match;
          return (
            <Link key={i} href={url} className="text-blue-600 hover:underline">
              {label}
            </Link>
          );
        }

        // Plain text
        return <span key={i}>{part}</span>;
      })}
    </span>
  );
}
