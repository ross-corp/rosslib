/**
 * Fallback placeholder when a book has no cover image.
 * Renders the title (and optionally author) on a styled background.
 */
export default function BookCoverPlaceholder({
  title,
  author,
  className = "",
}: {
  title: string;
  author?: string | null;
  className?: string;
}) {
  return (
    <div
      className={`bg-surface-2 rounded shadow-sm flex flex-col items-center justify-center p-2 text-center ${className}`}
    >
      <span className="text-[11px] font-medium text-text-secondary leading-tight line-clamp-3">
        {title}
      </span>
      {author && (
        <span className="text-[9px] text-text-tertiary leading-tight mt-1 line-clamp-1">
          {author}
        </span>
      )}
    </div>
  );
}
