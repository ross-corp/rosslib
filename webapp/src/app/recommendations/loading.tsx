export default function RecommendationsLoading() {
  return (
    <div className="min-h-screen">
      <main className="max-w-2xl mx-auto px-4 sm:px-6 py-12">
        {/* Heading */}
        <div className="h-7 w-52 bg-surface-2 rounded animate-pulse mb-6" />

        {/* Received / Sent tabs */}
        <div className="flex gap-1 mb-6 border-b border-border">
          <div className="h-4 w-20 bg-surface-2 rounded animate-pulse my-2" />
          <div className="h-4 w-12 bg-surface-2 rounded animate-pulse my-2 ml-2" />
        </div>

        {/* Status filter sub-tabs */}
        <div className="flex gap-1 mb-6 border-b border-border">
          {["w-16", "w-12", "w-20", "w-8"].map((w, i) => (
            <div
              key={i}
              className={`h-3 ${w} bg-surface-2 rounded animate-pulse my-2`}
            />
          ))}
        </div>

        {/* Recommendation cards */}
        <div className="space-y-4">
          {Array.from({ length: 4 }).map((_, i) => (
            <div
              key={i}
              className="flex gap-4 p-4 border border-border rounded-lg"
            >
              {/* Book cover */}
              <div className="w-16 h-24 bg-surface-2 rounded animate-pulse shrink-0" />

              <div className="flex-1 min-w-0">
                {/* Book title */}
                <div className="h-4 w-40 bg-surface-2 rounded animate-pulse" />
                {/* Author */}
                <div className="h-3 w-24 bg-surface-2 rounded animate-pulse mt-1" />

                {/* Sender row */}
                <div className="flex items-center gap-2 mt-2">
                  <div className="w-5 h-5 rounded-full bg-surface-2 animate-pulse shrink-0" />
                  <div className="h-3 w-20 bg-surface-2 rounded animate-pulse" />
                  <div className="h-3 w-14 bg-surface-2 rounded animate-pulse" />
                </div>

                {/* Note placeholder */}
                <div className="h-3 w-full bg-surface-2 rounded animate-pulse mt-2" />

                {/* Action buttons */}
                <div className="flex items-center gap-2 mt-3">
                  <div className="h-7 w-20 bg-surface-2 rounded animate-pulse" />
                  <div className="h-7 w-20 bg-surface-2 rounded animate-pulse" />
                  <div className="h-7 w-16 bg-surface-2 rounded animate-pulse" />
                </div>
              </div>
            </div>
          ))}
        </div>
      </main>
    </div>
  );
}
