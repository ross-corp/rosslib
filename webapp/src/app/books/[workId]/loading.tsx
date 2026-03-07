export default function BookLoading() {
  return (
    <div className="min-h-screen">
      <main className="max-w-3xl mx-auto px-4 sm:px-6 py-12">
        {/* Book header */}
        <div className="flex gap-8 items-start mb-10">
          {/* Cover placeholder */}
          <div className="shrink-0">
            <div className="w-32 h-48 bg-surface-2 rounded shadow-sm animate-pulse" />
          </div>

          <div className="flex-1 min-w-0">
            {/* Title */}
            <div className="h-7 w-64 bg-surface-2 rounded animate-pulse mb-2" />
            {/* Author */}
            <div className="h-4 w-36 bg-surface-2 rounded animate-pulse mb-1" />
            {/* Publish info */}
            <div className="h-3 w-44 bg-surface-2 rounded animate-pulse mb-3" />
            {/* Rating */}
            <div className="h-4 w-32 bg-surface-2 rounded animate-pulse mb-3" />
            {/* Reader counts */}
            <div className="h-3 w-28 bg-surface-2 rounded animate-pulse mb-4" />
            {/* Status picker placeholder */}
            <div className="h-8 w-40 bg-surface-2 rounded animate-pulse mb-4" />
            {/* Description lines */}
            <div className="space-y-2">
              <div className="h-3 w-full bg-surface-2 rounded animate-pulse" />
              <div className="h-3 w-full bg-surface-2 rounded animate-pulse" />
              <div className="h-3 w-3/4 bg-surface-2 rounded animate-pulse" />
            </div>
            {/* Subject chips */}
            <div className="flex gap-2 mt-4 flex-wrap">
              {Array.from({ length: 4 }).map((_, i) => (
                <div
                  key={i}
                  className="h-6 w-16 bg-surface-2 rounded-full animate-pulse"
                />
              ))}
            </div>
          </div>
        </div>

        {/* Review section */}
        <div className="border-t border-border pt-8 mt-10">
          <div className="h-5 w-28 bg-surface-2 rounded animate-pulse mb-6" />
          <div className="space-y-8">
            {Array.from({ length: 3 }).map((_, i) => (
              <div key={i} className="flex gap-4">
                <div className="w-8 h-8 rounded-full bg-surface-2 animate-pulse shrink-0" />
                <div className="flex-1">
                  <div className="h-4 w-24 bg-surface-2 rounded animate-pulse mb-2" />
                  <div className="h-3 w-20 bg-surface-2 rounded animate-pulse mb-2" />
                  <div className="space-y-1.5">
                    <div className="h-3 w-full bg-surface-2 rounded animate-pulse" />
                    <div className="h-3 w-5/6 bg-surface-2 rounded animate-pulse" />
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </main>
    </div>
  );
}
