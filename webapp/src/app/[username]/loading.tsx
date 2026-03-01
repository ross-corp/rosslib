export default function ProfileLoading() {
  return (
    <>
      {/* Header */}
      <div className="mb-10">
        <div className="flex items-start justify-between mb-4">
          <div className="flex items-center gap-4">
            {/* Avatar */}
            <div className="w-14 h-14 rounded-full bg-surface-2 animate-pulse shrink-0" />
            <div>
              {/* Display name */}
              <div className="h-7 w-40 bg-surface-2 rounded animate-pulse" />
              {/* Username */}
              <div className="h-4 w-24 bg-surface-2 rounded animate-pulse mt-1" />
            </div>
          </div>
        </div>

        {/* Bio */}
        <div className="h-4 w-64 bg-surface-2 rounded animate-pulse mb-4" />

        {/* Stats row */}
        <div className="flex items-center gap-4 flex-wrap">
          {Array.from({ length: 5 }).map((_, i) => (
            <div key={i} className="h-4 w-16 bg-surface-2 rounded animate-pulse" />
          ))}
        </div>
        {/* Member since */}
        <div className="h-3 w-32 bg-surface-2 rounded animate-pulse mt-1" />
      </div>

      {/* Main content grid */}
      <div className="lg:grid lg:grid-cols-3 lg:gap-8">
        {/* Left column */}
        <div className="lg:col-span-2 space-y-10">
          {/* Currently reading section */}
          <div>
            <div className="h-5 w-36 bg-surface-2 rounded animate-pulse mb-4" />
            <div className="flex gap-4">
              {Array.from({ length: 3 }).map((_, i) => (
                <div key={i} className="w-24 h-36 bg-surface-2 rounded shadow-sm animate-pulse" />
              ))}
            </div>
          </div>

          {/* Want to read section */}
          <div>
            <div className="h-5 w-28 bg-surface-2 rounded animate-pulse mb-4" />
            <div className="flex gap-4">
              {Array.from({ length: 4 }).map((_, i) => (
                <div key={i} className="w-20 h-[120px] bg-surface-2 rounded shadow-sm animate-pulse" />
              ))}
            </div>
          </div>

          {/* Recent reviews section */}
          <div>
            <div className="h-5 w-32 bg-surface-2 rounded animate-pulse mb-4" />
            <div className="space-y-4">
              {Array.from({ length: 2 }).map((_, i) => (
                <div key={i} className="flex gap-3">
                  <div className="w-12 h-[72px] bg-surface-2 rounded shadow-sm animate-pulse shrink-0" />
                  <div className="flex-1">
                    <div className="h-4 w-32 bg-surface-2 rounded animate-pulse mb-1" />
                    <div className="h-3 w-20 bg-surface-2 rounded animate-pulse mb-2" />
                    <div className="h-3 w-full bg-surface-2 rounded animate-pulse" />
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* Right column — activity */}
        <div>
          <div className="h-5 w-28 bg-surface-2 rounded animate-pulse mb-4" />
          <div className="space-y-4">
            {Array.from({ length: 4 }).map((_, i) => (
              <div key={i}>
                <div className="h-3 w-full bg-surface-2 rounded animate-pulse mb-1" />
                <div className="h-3 w-3/4 bg-surface-2 rounded animate-pulse" />
              </div>
            ))}
          </div>
        </div>
      </div>
    </>
  );
}
