"use client";

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
  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
      onClick={(e) => {
        if (e.target === e.currentTarget) onCancel();
      }}
    >
      <div className="bg-surface-0 border border-border rounded-lg p-6 w-full max-w-sm mx-4 shadow-lg">
        <h3 className="text-sm font-semibold text-text-primary mb-2">
          {title}
        </h3>
        <p className="text-xs text-text-secondary mb-4">{message}</p>
        <div className="flex items-center gap-3">
          <button
            onClick={onConfirm}
            className="text-xs px-4 py-2 rounded bg-red-600 text-white hover:bg-red-700 transition-colors"
          >
            {confirmLabel}
          </button>
          <button
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
