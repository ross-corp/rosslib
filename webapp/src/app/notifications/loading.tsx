export default function NotificationsLoading() {
  return (
    <div className="min-h-screen">
      <main className="max-w-2xl mx-auto px-4 sm:px-6 py-12">
        {/* Header */}
        <div className="flex items-center justify-between mb-8">
          <div className="flex items-center gap-4">
            <div className="h-7 w-36 bg-surface-2 rounded animate-pulse" />
            <div className="h-7 w-28 bg-surface-2 rounded animate-pulse" />
          </div>
        </div>

        {/* Notification rows */}
        <div className="divide-y divide-border">
          {Array.from({ length: 5 }).map((_, i) => (
            <div key={i} className="flex items-start gap-3 py-4">
              <div className="w-8 h-8 rounded-full bg-surface-2 animate-pulse shrink-0" />
              <div className="flex-1">
                <div className="h-4 w-48 bg-surface-2 rounded animate-pulse mb-2" />
                <div className="h-3 w-full bg-surface-2 rounded animate-pulse mb-1" />
                <div className="h-3 w-20 bg-surface-2 rounded animate-pulse mt-2" />
              </div>
            </div>
          ))}
        </div>
      </main>
    </div>
  );
}
