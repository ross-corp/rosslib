export default function ReviewsLoading() {
  return (
    <div className="min-h-screen">
      <main className="max-w-2xl mx-auto px-4 sm:px-6 py-12">
        {/* Back link */}
        <div className="mb-8">
          <div className="h-4 w-20 bg-surface-2 rounded animate-pulse" />
          <div className="flex items-center justify-between mt-2">
            <div className="h-7 w-36 bg-surface-2 rounded animate-pulse" />
            <div className="h-8 w-28 bg-surface-2 rounded animate-pulse" />
          </div>
        </div>

        {/* Review cards */}
        <div className="space-y-8">
          {Array.from({ length: 5 }).map((_, i) => (
            <div key={i} className="flex gap-4">
              {/* Cover */}
              <div className="w-16 h-24 bg-surface-2 rounded animate-pulse shrink-0" />

              <div className="flex-1 min-w-0">
                {/* Title */}
                <div className="h-4 w-40 bg-surface-2 rounded animate-pulse" />
                {/* Author */}
                <div className="h-3 w-24 bg-surface-2 rounded animate-pulse mt-1" />

                {/* Rating + date */}
                <div className="flex items-center gap-3 mt-1.5">
                  <div className="h-3.5 w-20 bg-surface-2 rounded animate-pulse" />
                  <div className="h-3 w-28 bg-surface-2 rounded animate-pulse" />
                </div>

                {/* Review text lines */}
                <div className="mt-2 space-y-1.5">
                  <div className="h-3 w-full bg-surface-2 rounded animate-pulse" />
                  <div className="h-3 w-full bg-surface-2 rounded animate-pulse" />
                  <div className="h-3 w-3/4 bg-surface-2 rounded animate-pulse" />
                </div>

                {/* Date + like count */}
                <div className="flex items-center gap-3 mt-2">
                  <div className="h-3 w-24 bg-surface-2 rounded animate-pulse" />
                  <div className="h-3 w-8 bg-surface-2 rounded animate-pulse" />
                </div>
              </div>
            </div>
          ))}
        </div>
      </main>
    </div>
  );
}
