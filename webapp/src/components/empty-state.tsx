import Link from "next/link";

export default function EmptyState({
  message,
  actionLabel,
  actionHref,
}: {
  message: string;
  actionLabel?: string;
  actionHref?: string;
}) {
  return (
    <div className="text-center py-16">
      <p className="text-text-primary text-sm">{message}</p>
      {actionLabel && actionHref && (
        <Link
          href={actionHref}
          className="inline-block mt-4 text-sm text-text-primary hover:text-text-primary border border-border px-4 py-2 rounded transition-colors"
        >
          {actionLabel}
        </Link>
      )}
    </div>
  );
}
