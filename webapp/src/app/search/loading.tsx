export default function SearchLoading() {
  return (
    <div className="min-h-screen">
      <main className="max-w-5xl mx-auto px-4 sm:px-6 py-10">
        {/* Search bar placeholder */}
        <div className="mb-6 max-w-md">
          <div className="h-9 bg-surface-2 rounded animate-pulse" />
        </div>

        {/* Tab selector */}
        <div className="flex gap-1 mb-6 border-b border-border">
          <div className="h-4 w-14 bg-surface-2 rounded animate-pulse px-4 py-2 mb-2" />
          <div className="h-4 w-16 bg-surface-2 rounded animate-pulse px-4 py-2 mb-2" />
          <div className="h-4 w-14 bg-surface-2 rounded animate-pulse px-4 py-2 mb-2" />
        </div>

        {/* Popular section heading */}
        <div className="h-4 w-24 bg-surface-2 rounded animate-pulse mb-2" />
        <div className="h-6 w-40 bg-surface-2 rounded animate-pulse mb-4" />

        {/* Grid of 8 skeleton book cards */}
        <div className="grid grid-cols-3 sm:grid-cols-4 md:grid-cols-6 gap-4">
          {Array.from({ length: 8 }).map((_, i) => (
            <div key={i} className="flex flex-col items-center">
              <div className="w-24 h-36 bg-surface-2 rounded shadow-sm animate-pulse" />
              <div className="h-3 w-16 bg-surface-2 rounded animate-pulse mt-2" />
              <div className="h-3 w-12 bg-surface-2 rounded animate-pulse mt-1" />
            </div>
          ))}
        </div>
      </main>
    </div>
  );
}
