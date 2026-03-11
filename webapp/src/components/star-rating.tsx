type Props = {
  rating: number;
  count?: number;
  className?: string;
};

export default function StarRating({ rating, count, className }: Props) {
  const filled = Math.round(rating);
  const stars = Array.from({ length: 5 }, (_, i) =>
    i < filled ? "⭐" : "☆"
  ).join("");

  return (
    <span className={className} aria-label={`${rating.toFixed(2)} out of 5 stars`}>
      <span className="tracking-tight" aria-hidden="true">{stars}</span>{" "}
      <span className="text-text-primary">{rating.toFixed(2)}</span>
      {count != null && count > 0 && (
        <span className="text-text-primary"> ({count.toLocaleString()})</span>
      )}
    </span>
  );
}
