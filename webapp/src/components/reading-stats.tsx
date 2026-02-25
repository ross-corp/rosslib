export default function ReadingStats({
  booksRead,
  reviewsCount,
  booksThisYear,
  averageRating,
}: {
  booksRead: number;
  reviewsCount: number;
  booksThisYear: number;
  averageRating: number | null;
}) {
  return (
    <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
      <div>
        <p className="text-2xl font-bold text-text-primary">{booksRead}</p>
        <p className="text-xs text-text-tertiary">books read</p>
      </div>
      <div>
        <p className="text-2xl font-bold text-text-primary">{booksThisYear}</p>
        <p className="text-xs text-text-tertiary">this year</p>
      </div>
      <div>
        <p className="text-2xl font-bold text-text-primary">
          {averageRating != null ? averageRating.toFixed(1) : "—"}
        </p>
        <p className="text-xs text-text-tertiary">avg rating</p>
      </div>
      <div>
        <p className="text-2xl font-bold text-text-primary">{reviewsCount}</p>
        <p className="text-xs text-text-tertiary">reviews</p>
      </div>
    </div>
  );
}
