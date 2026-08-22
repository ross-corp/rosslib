"use client";

import { useEffect, useId, useRef } from "react";

type Props = {
  title: string;
  message: string;
  confirmLabel?: string;
  onConfirm: () => void;
  onCancel: () => void;
};

export default function ConfirmDialog({
  title,
  message,
  confirmLabel = "Remove",
  onConfirm,
  onCancel,
}: Props) {
  const titleId = useId();
  const descId = useId();
  const cancelRef = useRef<HTMLButtonElement>(null);

  useEffect(() => {
    cancelRef.current?.focus();
  }, []);

  useEffect(() => {
    const onKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") onCancel();
    };
    document.addEventListener("keydown", onKeyDown);
    return () => document.removeEventListener("keydown", onKeyDown);
  }, [onCancel]);

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
      onClick={(e) => {
        if (e.target === e.currentTarget) onCancel();
      }}
    >
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby={titleId}
        aria-describedby={descId}
        className="bg-surface-0 border border-border rounded-lg p-6 w-full max-w-sm mx-4 shadow-lg"
      >
        <h3
          id={titleId}
          className="text-sm font-semibold text-text-primary mb-2"
        >
          {title}
        </h3>
        <p id={descId} className="text-xs text-text-secondary mb-4">
          {message}
        </p>
        <div className="flex items-center gap-3">
          <button
            onClick={onConfirm}
            className="text-xs px-4 py-2 rounded bg-danger text-badge-text hover:bg-danger-bg-hover transition-colors"
          >
            {confirmLabel}
          </button>
          <button
            ref={cancelRef}
            onClick={onCancel}
            className="text-xs text-text-secondary hover:text-text-primary transition-colors"
          >
            Cancel
          </button>
        </div>
      </div>
    </div>
  );
}
