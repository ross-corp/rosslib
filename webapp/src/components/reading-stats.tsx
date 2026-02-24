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
        <p className="text-2xl font-bold text-stone-900">{booksRead}</p>
        <p className="text-xs text-stone-400">books read</p>
      </div>
      <div>
        <p className="text-2xl font-bold text-stone-900">{booksThisYear}</p>
        <p className="text-xs text-stone-400">this year</p>
      </div>
      <div>
        <p className="text-2xl font-bold text-stone-900">
          {averageRating != null ? averageRating.toFixed(1) : "—"}
        </p>
        <p className="text-xs text-stone-400">avg rating</p>
      </div>
      <div>
        <p className="text-2xl font-bold text-stone-900">{reviewsCount}</p>
        <p className="text-xs text-stone-400">reviews</p>
      </div>
    </div>
  );
}
