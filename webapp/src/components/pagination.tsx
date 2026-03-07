import Link from "next/link";

export default function Pagination({
  prevHref,
  nextHref,
  label,
}: {
  prevHref: string | null;
  nextHref: string | null;
  label?: string;
}) {
  if (!prevHref && !nextHref) return null;

  return (
    <nav className="flex items-center justify-between mt-8">
      {prevHref ? (
        <Link
          href={prevHref}
          className="text-sm px-4 py-2 rounded border border-border text-text-primary hover:bg-surface-2 transition-colors"
        >
          Previous
        </Link>
      ) : (
        <span />
      )}
      {label && (
        <span className="text-xs text-text-secondary">{label}</span>
      )}
      {nextHref ? (
        <Link
          href={nextHref}
          className="text-sm px-4 py-2 rounded border border-border text-text-primary hover:bg-surface-2 transition-colors"
        >
          Next
        </Link>
      ) : (
        <span />
      )}
    </nav>
  );
}
