"use client";

import { useState } from "react";

export default function AuthorPhoto({
  src,
  name,
  className = "",
}: {
  src: string | null;
  name: string;
  className?: string;
}) {
  const [failed, setFailed] = useState(false);

  if (!src || failed) {
    return (
      <div
        className={`bg-surface-2 rounded flex items-center justify-center text-3xl font-semibold text-text-primary ${className}`}
      >
        {name.charAt(0)}
      </div>
    );
  }

  return (
    <img
      src={src}
      alt={name}
      className={`rounded shadow-sm object-cover bg-surface-2 ${className}`}
      onError={() => setFailed(true)}
    />
  );
}
