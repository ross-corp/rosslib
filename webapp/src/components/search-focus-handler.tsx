"use client";

import { useEffect } from "react";

export default function SearchFocusHandler() {
  useEffect(() => {
    function handleKeyDown(e: KeyboardEvent) {
      if (!(e.key === "k" && (e.metaKey || e.ctrlKey))) return;

      const tag = (e.target as HTMLElement)?.tagName;
      if (
        tag === "INPUT" ||
        tag === "TEXTAREA" ||
        (e.target as HTMLElement)?.isContentEditable
      ) {
        return;
      }

      e.preventDefault();

      const search =
        document.querySelector<HTMLInputElement>('input[type="search"]') ??
        document.querySelector<HTMLInputElement>('input[name="q"]');
      search?.focus();
    }

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, []);

  return null;
}
