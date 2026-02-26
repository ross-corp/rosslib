"use client";

import { useState } from "react";
import EditionPicker from "@/components/edition-picker";

type Edition = {
  key: string;
  title: string;
  publisher: string | null;
  publish_date: string;
  page_count: number | null;
  isbn: string | null;
  cover_url: string | null;
  format: string;
  language: string;
};

export default function BookCoverWithEdition({
  defaultCoverUrl,
  title,
  openLibraryId,
  workId,
  editions,
  totalEditions,
  selectedEditionKey,
  selectedEditionCoverUrl,
  showEditionPicker,
}: {
  defaultCoverUrl: string | null;
  title: string;
  openLibraryId: string;
  workId: string;
  editions: Edition[];
  totalEditions: number;
  selectedEditionKey: string | null;
  selectedEditionCoverUrl: string | null;
  showEditionPicker: boolean;
}) {
  const [coverUrl, setCoverUrl] = useState(
    selectedEditionCoverUrl || defaultCoverUrl
  );
  const [editionKey, setEditionKey] = useState(selectedEditionKey);

  function handleEditionChanged(key: string | null, cover: string | null) {
    setEditionKey(key);
    setCoverUrl(cover || defaultCoverUrl);
  }

  return (
    <div className="shrink-0">
      {coverUrl ? (
        <img
          src={coverUrl}
          alt={title}
          className="w-32 rounded shadow-sm object-cover"
        />
      ) : (
        <div className="w-32 h-48 bg-surface-2 rounded" />
      )}
      {showEditionPicker && editions.length > 0 && (
        <div className="mt-2">
          <EditionPicker
            openLibraryId={openLibraryId}
            workId={workId}
            initialEditions={editions}
            totalEditions={totalEditions}
            currentEditionKey={editionKey}
            currentEditionCoverUrl={coverUrl !== defaultCoverUrl ? coverUrl : null}
            onEditionChanged={handleEditionChanged}
          />
        </div>
      )}
    </div>
  );
}
