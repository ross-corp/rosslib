"use client";

import Link from "next/link";
import { useRecentlyViewed } from "@/lib/recently-viewed";

export default function RecentlyViewedBooks() {
  const { books } = useRecentlyViewed();

  if (books.length === 0) return null;

  return (
    <div className="mb-8">
      <h2 className="text-lg font-semibold text-text-primary mb-4">
        Recently Viewed
      </h2>
      <div className="flex gap-4 overflow-x-auto pb-2">
        {books.map((book) => {
          return (
            <Link
              key={book.workId}
              href={`/books/${book.workId}`}
              className="group flex flex-col items-center text-center shrink-0"
            >
              {book.coverUrl ? (
                <img
                  src={book.coverUrl}
                  alt={book.title}
                  width={96}
                  height={144}
                  className="w-24 h-36 object-cover rounded shadow-sm bg-surface-2 group-hover:shadow-md transition-shadow"
                />
              ) : (
                <div className="w-24 h-36 bg-surface-2 rounded shadow-sm flex items-center justify-center">
                  <span className="text-xs text-text-tertiary px-2 text-center leading-tight">
                    {book.title}
                  </span>
                </div>
              )}
              <span className="text-xs font-medium text-text-primary mt-2 line-clamp-2 max-w-[6rem]">
                {book.title}
              </span>
            </Link>
          );
        })}
      </div>
    </div>
  );
}
