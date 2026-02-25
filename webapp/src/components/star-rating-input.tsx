"use client";

import { useState } from "react";

type Props = {
  value: number | null;
  onChange: (rating: number | null) => void;
  disabled?: boolean;
};

export default function StarRatingInput({ value, onChange, disabled }: Props) {
  const [hover, setHover] = useState<number | null>(null);

  const display = hover ?? value ?? 0;

  return (
    <span className="inline-flex items-center gap-0.5">
      {[1, 2, 3, 4, 5].map((star) => (
        <button
          key={star}
          type="button"
          disabled={disabled}
          onMouseEnter={() => setHover(star)}
          onMouseLeave={() => setHover(null)}
          onClick={() => onChange(star === value ? null : star)}
          className="text-lg leading-none transition-colors disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer select-none"
          aria-label={`${star} star${star !== 1 ? "s" : ""}`}
        >
          <span className={star <= display ? "text-amber-500" : "text-text-primary"}>
            ★
          </span>
        </button>
      ))}
    </span>
  );
}
