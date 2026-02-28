"use client";

import { useState, useCallback } from "react";
import { useKeyboardShortcuts } from "@/lib/keyboard-shortcuts";
import KeyboardShortcutsOverlay from "@/components/keyboard-shortcuts-overlay";

export default function KeyboardShortcuts({
  showHint,
}: {
  showHint: boolean;
}) {
  const [open, setOpen] = useState(false);

  const toggle = useCallback(() => setOpen((o) => !o), []);

  useKeyboardShortcuts(toggle);

  return (
    <>
      {open && (
        <KeyboardShortcutsOverlay onClose={() => setOpen(false)} />
      )}
      {showHint && !open && (
        <div className="fixed bottom-4 right-4 z-30">
          <button
            onClick={() => setOpen(true)}
            className="font-mono text-[10px] text-text-tertiary hover:text-text-secondary bg-surface-1 border border-border rounded px-2 py-1 transition-colors"
          >
            Press{" "}
            <kbd className="font-mono bg-surface-2 border border-border rounded px-1 py-0.5">
              ?
            </kbd>{" "}
            for shortcuts
          </button>
        </div>
      )}
    </>
  );
}
