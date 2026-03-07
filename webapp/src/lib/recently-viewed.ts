"use client";

import { useState, useEffect, useCallback } from "react";

export type RecentlyViewedBook = {
  workId: string;
  title: string;
  coverUrl: string | null;
};

const STORAGE_KEY = "rosslib_recently_viewed";
const MAX_ITEMS = 10;

function loadFromStorage(): RecentlyViewedBook[] {
  if (typeof window === "undefined") return [];
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return [];
    const parsed = JSON.parse(raw);
    if (!Array.isArray(parsed)) return [];
    return parsed;
  } catch {
    return [];
  }
}

function saveToStorage(books: RecentlyViewedBook[]) {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(books));
  } catch {
    // localStorage might be full or unavailable
  }
}

export function addToRecentlyViewed(book: RecentlyViewedBook) {
  const current = loadFromStorage();
  const filtered = current.filter((b) => b.workId !== book.workId);
  const updated = [book, ...filtered].slice(0, MAX_ITEMS);
  saveToStorage(updated);
}

export function useRecentlyViewed() {
  const [books, setBooks] = useState<RecentlyViewedBook[]>([]);

  useEffect(() => {
    setBooks(loadFromStorage());
  }, []);

  const add = useCallback((book: RecentlyViewedBook) => {
    addToRecentlyViewed(book);
    setBooks(loadFromStorage());
  }, []);

  return { books, add };
}
