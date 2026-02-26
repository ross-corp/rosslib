"use client";

import { useState } from "react";
import EditionSelector from "@/components/edition-selector";

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
  workId,
  openLibraryId,
  title,
  defaultCoverUrl,
  initialEditionCoverUrl,
  initialEditionKey,
  editions,
  totalEditions,
  hasBook,
}: {
  workId: string;
  openLibraryId: string;
  title: string;
  defaultCoverUrl: string | null;
  initialEditionCoverUrl: string | null;
  initialEditionKey: string | null;
  editions: Edition[];
  totalEditions: number;
  hasBook: boolean;
}) {
  const [editionCoverUrl, setEditionCoverUrl] = useState<string | null>(
    initialEditionCoverUrl
  );
  const [editionKey, setEditionKey] = useState<string | null>(initialEditionKey);

  const displayCover = editionCoverUrl || defaultCoverUrl;

  return (
    <div className="shrink-0">
      {displayCover ? (
        <img
          src={displayCover}
          alt={title}
          className="w-32 rounded shadow-sm object-cover"
        />
      ) : (
        <div className="w-32 h-48 bg-surface-2 rounded" />
      )}

      {hasBook && editions.length > 0 && (
        <div className="mt-2">
          <EditionSelector
            workId={workId}
            openLibraryId={openLibraryId}
            editions={editions}
            totalEditions={totalEditions}
            currentEditionKey={editionKey}
            defaultCoverUrl={defaultCoverUrl}
            onEditionChanged={(key, cover) => {
              setEditionKey(key);
              setEditionCoverUrl(cover);
            }}
          />
        </div>
      )}
    </div>
  );
}
