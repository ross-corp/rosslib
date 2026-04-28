export default function SettingsLoading() {
  return (
    <div className="min-h-screen">
      <main className="max-w-5xl mx-auto px-4 sm:px-6 py-10">
        {/* Breadcrumb */}
        <div className="mb-8 flex items-center gap-2">
          <div className="h-4 w-20 bg-surface-2 rounded animate-pulse" />
          <span className="text-text-primary">/</span>
          <div className="h-4 w-16 bg-surface-2 rounded animate-pulse" />
        </div>

        {/* Heading */}
        <div className="h-7 w-28 bg-surface-2 rounded animate-pulse mb-4" />

        {/* Nav pills */}
        <div className="flex gap-2 mb-8 flex-wrap">
          {["w-16", "w-14", "w-16", "w-20", "w-24", "w-14"].map((w, i) => (
            <div
              key={i}
              className={`h-8 ${w} bg-surface-2 rounded-full animate-pulse`}
            />
          ))}
        </div>

        {/* Content placeholder */}
        <div className="space-y-4">
          <div className="h-5 w-40 bg-surface-2 rounded animate-pulse" />
          <div className="h-4 w-full bg-surface-2 rounded animate-pulse" />
          <div className="h-4 w-3/4 bg-surface-2 rounded animate-pulse" />
          <div className="h-10 w-full bg-surface-2 rounded animate-pulse mt-2" />
          <div className="h-10 w-full bg-surface-2 rounded animate-pulse" />
          <div className="h-10 w-full bg-surface-2 rounded animate-pulse" />
        </div>
      </main>
    </div>
  );
}
