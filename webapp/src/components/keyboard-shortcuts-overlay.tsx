"use client";

import { SHORTCUTS } from "@/lib/keyboard-shortcuts";

type Props = {
  onClose: () => void;
};

export default function KeyboardShortcutsOverlay({ onClose }: Props) {
  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
      data-keyboard-dismiss
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
    >
      <div className="bg-surface-0 border border-border rounded-lg p-6 w-full max-w-sm mx-4 shadow-lg">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-sm font-semibold text-text-primary">
            Keyboard shortcuts
          </h3>
          <button
            onClick={onClose}
            className="text-text-tertiary hover:text-text-primary text-xs transition-colors"
          >
            esc
          </button>
        </div>
        <div className="space-y-2">
          {SHORTCUTS.map((s) => (
            <div
              key={s.key}
              className="flex items-center justify-between text-sm"
            >
              <span className="text-text-secondary">{s.description}</span>
              <kbd className="font-mono text-xs bg-surface-2 border border-border rounded px-1.5 py-0.5 text-text-tertiary min-w-[28px] text-center">
                {s.label}
              </kbd>
            </div>
          ))}
          <div className="flex items-center justify-between text-sm">
            <span className="text-text-secondary">Focus search</span>
            <kbd className="font-mono text-xs bg-surface-2 border border-border rounded px-1.5 py-0.5 text-text-tertiary">
              ⌘K
            </kbd>
          </div>
        </div>
      </div>
    </div>
  );
}
