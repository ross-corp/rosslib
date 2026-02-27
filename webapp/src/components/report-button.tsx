"use client";

import { useState } from "react";
import ReportModal from "@/components/report-modal";

type Props = {
  contentType: "review" | "thread" | "comment" | "link";
  contentId: string;
};

export default function ReportButton({ contentType, contentId }: Props) {
  const [showModal, setShowModal] = useState(false);

  return (
    <>
      <button
        type="button"
        onClick={() => setShowModal(true)}
        className="px-1.5 py-1 rounded text-text-primary hover:text-red-500 hover:bg-red-50 transition-colors"
        title={`Report this ${contentType}`}
      >
        <svg viewBox="0 0 12 12" className="w-3 h-3" fill="none" stroke="currentColor" strokeWidth={1.5}>
          <path d="M2 1v10M2 1h7l-2 3 2 3H2" />
        </svg>
      </button>
      {showModal && (
        <ReportModal
          contentType={contentType}
          contentId={contentId}
          onClose={() => setShowModal(false)}
        />
      )}
    </>
  );
}
