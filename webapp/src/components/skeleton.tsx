type SkeletonVariant = "text" | "circular" | "rectangular";

type SkeletonProps = {
  width?: string;
  height?: string;
  variant?: SkeletonVariant;
  className?: string;
};

export default function Skeleton({
  width,
  height,
  variant = "text",
  className = "",
}: SkeletonProps) {
  const base = "bg-surface-2 animate-skeleton-pulse";
  const shape =
    variant === "circular"
      ? "rounded-full"
      : variant === "rectangular"
        ? "rounded"
        : "rounded";

  return (
    <div
      className={`${base} ${shape} ${className}`}
      style={{ width, height }}
    />
  );
}

export function BookGridSkeleton({ count = 6 }: { count?: number }) {
  return (
    <div className="grid grid-cols-3 sm:grid-cols-4 md:grid-cols-5 lg:grid-cols-6 gap-5">
      {Array.from({ length: count }, (_, i) => (
        <div key={i} className="space-y-2">
          <Skeleton variant="rectangular" className="aspect-[2/3] w-full" />
          <Skeleton variant="text" className="h-3 w-3/4" />
          <Skeleton variant="text" className="h-3 w-1/2" />
        </div>
      ))}
    </div>
  );
}

export function ProfileSkeleton() {
  return (
    <div>
      {/* Header */}
      <div className="mb-10">
        <div className="flex items-start justify-between mb-4">
          <div className="flex items-center gap-4">
            <Skeleton variant="circular" className="w-14 h-14 shrink-0" />
            <div className="space-y-2">
              <Skeleton variant="text" className="h-6 w-40" />
              <Skeleton variant="text" className="h-4 w-24" />
            </div>
          </div>
        </div>
        <Skeleton variant="text" className="h-4 w-64 mb-4" />
        <div className="flex items-center gap-4">
          <Skeleton variant="text" className="h-4 w-16" />
          <Skeleton variant="text" className="h-4 w-20" />
          <Skeleton variant="text" className="h-4 w-20" />
          <Skeleton variant="text" className="h-4 w-16" />
        </div>
      </div>

      {/* Content */}
      <div className="lg:grid lg:grid-cols-3 lg:gap-8">
        <div className="lg:col-span-2 space-y-10">
          {/* Currently Reading section */}
          <section>
            <Skeleton variant="text" className="h-3 w-32 mb-3" />
            <div className="flex gap-4">
              {Array.from({ length: 4 }, (_, i) => (
                <Skeleton
                  key={i}
                  variant="rectangular"
                  className="w-20 h-28 shrink-0"
                />
              ))}
            </div>
          </section>

          {/* Reading Stats section */}
          <section>
            <Skeleton variant="text" className="h-3 w-28 mb-3" />
            <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
              {Array.from({ length: 4 }, (_, i) => (
                <Skeleton
                  key={i}
                  variant="rectangular"
                  className="h-16 w-full"
                />
              ))}
            </div>
          </section>
        </div>

        {/* Sidebar */}
        <div className="mt-10 lg:mt-0 space-y-4">
          <Skeleton variant="text" className="h-3 w-28" />
          {Array.from({ length: 3 }, (_, i) => (
            <Skeleton
              key={i}
              variant="rectangular"
              className="h-16 w-full"
            />
          ))}
        </div>
      </div>
    </div>
  );
}

export function ReviewSkeleton({ count = 3 }: { count?: number }) {
  return (
    <div className="space-y-8">
      {Array.from({ length: count }, (_, i) => (
        <div key={i} className="flex gap-4">
          <Skeleton variant="circular" className="w-8 h-8 shrink-0" />
          <div className="flex-1 space-y-2">
            <Skeleton variant="text" className="h-4 w-28" />
            <Skeleton variant="text" className="h-3 w-20" />
            <Skeleton variant="text" className="h-4 w-full" />
            <Skeleton variant="text" className="h-4 w-3/4" />
          </div>
        </div>
      ))}
    </div>
  );
}

export function BookDetailSkeleton() {
  return (
    <div className="min-h-screen">
      <main className="max-w-3xl mx-auto px-4 sm:px-6 py-12">
        {/* Book header */}
        <div className="flex gap-8 items-start mb-10">
          <Skeleton variant="rectangular" className="w-32 h-48 shrink-0" />
          <div className="flex-1 space-y-3">
            <Skeleton variant="text" className="h-7 w-64" />
            <Skeleton variant="text" className="h-4 w-36" />
            <Skeleton variant="text" className="h-3 w-48" />
            <Skeleton variant="text" className="h-4 w-24" />
            <div className="pt-2 space-y-2">
              <Skeleton variant="text" className="h-4 w-full" />
              <Skeleton variant="text" className="h-4 w-full" />
              <Skeleton variant="text" className="h-4 w-3/4" />
            </div>
            <div className="flex gap-1.5 pt-2">
              {Array.from({ length: 4 }, (_, i) => (
                <Skeleton
                  key={i}
                  variant="text"
                  className="h-6 w-16 rounded-full"
                />
              ))}
            </div>
          </div>
        </div>

        {/* Reviews section */}
        <div className="border-t border-border pt-8">
          <Skeleton variant="text" className="h-4 w-24 mb-6" />
          <ReviewSkeleton count={2} />
        </div>
      </main>
    </div>
  );
}

export function FeedSkeleton() {
  return (
    <div className="min-h-screen">
      <main className="max-w-2xl mx-auto px-4 sm:px-6 py-12">
        <Skeleton variant="text" className="h-7 w-16 mb-8" />
        <div className="space-y-1">
          {Array.from({ length: 5 }, (_, i) => (
            <div key={i} className="py-3 space-y-2">
              <div className="flex items-center gap-2">
                <Skeleton variant="circular" className="w-6 h-6 shrink-0" />
                <Skeleton variant="text" className="h-4 w-48" />
              </div>
              <Skeleton variant="text" className="h-3 w-24 ml-8" />
            </div>
          ))}
        </div>
      </main>
    </div>
  );
}

export function SearchSkeleton() {
  return (
    <div className="min-h-screen">
      <main className="max-w-5xl mx-auto px-4 sm:px-6 py-10">
        {/* Search input */}
        <Skeleton variant="rectangular" className="h-10 w-80 mb-6" />

        {/* Tabs */}
        <div className="flex gap-1 mb-6 border-b border-border">
          <Skeleton variant="text" className="h-9 w-16" />
          <Skeleton variant="text" className="h-9 w-16" />
          <Skeleton variant="text" className="h-9 w-16" />
        </div>

        {/* Results */}
        <div className="space-y-4">
          {Array.from({ length: 5 }, (_, i) => (
            <div
              key={i}
              className="flex items-start gap-4 py-3"
            >
              <Skeleton
                variant="rectangular"
                className="w-12 h-16 shrink-0"
              />
              <div className="flex-1 space-y-2">
                <Skeleton variant="text" className="h-4 w-48" />
                <Skeleton variant="text" className="h-3 w-32" />
                <Skeleton variant="text" className="h-3 w-24" />
              </div>
            </div>
          ))}
        </div>
      </main>
    </div>
  );
}
