import Link from "next/link";
import React from "react";

type Props = {
  text: string;
};

export type Segment =
  | { type: "text"; content: string }
  | { type: "wikilink"; content: string }
  | { type: "link"; label: string; url: string };

export function parseSegments(text: string): Segment[] {
  if (!text) return [];

  // Split by [[Wikilink]] or [Markdown](Link)
  // use non-greedy matching for content inside brackets/parens
  // The capturing group means the delimiters are included in the result array
  const parts = text.split(/(\[\[.+?\]\]|\[.+?\]\(.+?\))/g);
  const segments: Segment[] = [];

  for (const part of parts) {
    if (!part) continue;

    // Handle [[Title]] -> Search
    if (part.startsWith("[[") && part.endsWith("]]")) {
      const content = part.slice(2, -2);
      segments.push({ type: "wikilink", content });
    }
    // Handle [Title](URL)
    else {
      const match = part.match(/^\[(.+?)\]\((.+?)\)$/);
      if (match) {
        const [, label, url] = match;
        segments.push({ type: "link", label, url });
      } else {
        // Plain text
        segments.push({ type: "text", content: part });
      }
    }
  }

  return segments;
}

export default function ReviewText({ text }: Props) {
  if (!text) return null;

  const segments = parseSegments(text);

  return (
    <span className="whitespace-pre-wrap">
      {segments.map((segment, i) => {
        if (segment.type === "wikilink") {
          return (
            <Link
              key={i}
              href={`/search?q=${encodeURIComponent(segment.content)}`}
              className="text-link hover:underline"
            >
              {segment.content}
            </Link>
          );
        } else if (segment.type === "link") {
          return (
            <Link
              key={i}
              href={segment.url}
              className="text-link hover:underline"
            >
              {segment.label}
            </Link>
          );
        }
        return <span key={i}>{segment.content}</span>;
      })}
    </span>
  );
}
