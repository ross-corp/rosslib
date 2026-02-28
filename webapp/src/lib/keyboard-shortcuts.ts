"use client";

import { useEffect, useCallback } from "react";

export type Shortcut = {
  key: string;
  label: string;
  description: string;
};

export const SHORTCUTS: Shortcut[] = [
  { key: "/", label: "/", description: "Focus search" },
  { key: "Escape", label: "Esc", description: "Close modal / dropdown" },
  { key: "?", label: "?", description: "Show keyboard shortcuts" },
];

function isEditableTarget(target: EventTarget | null): boolean {
  if (!target || !(target instanceof HTMLElement)) return false;
  const tag = target.tagName;
  return tag === "INPUT" || tag === "TEXTAREA" || target.isContentEditable;
}

export function useKeyboardShortcuts(onToggleHelp: () => void) {
  const handleKeyDown = useCallback(
    (e: KeyboardEvent) => {
      // Escape always works, even in inputs
      if (e.key === "Escape") {
        // Close any open modal overlay
        const overlay = document.querySelector<HTMLElement>(
          "[data-keyboard-dismiss]"
        );
        if (overlay) {
          overlay.click();
          return;
        }
        // Blur focused input
        if (document.activeElement instanceof HTMLElement) {
          document.activeElement.blur();
        }
        return;
      }

      // Skip other shortcuts when typing in an input
      if (isEditableTarget(e.target)) return;

      if (e.key === "/") {
        e.preventDefault();
        const search =
          document.querySelector<HTMLInputElement>('input[type="search"]') ??
          document.querySelector<HTMLInputElement>('input[name="q"]');
        search?.focus();
        return;
      }

      if (e.key === "?") {
        e.preventDefault();
        onToggleHelp();
        return;
      }
    },
    [onToggleHelp]
  );

  useEffect(() => {
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [handleKeyDown]);
}
