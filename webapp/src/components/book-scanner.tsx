"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import Link from "next/link";
import StatusPicker, { type StatusValue } from "@/components/shelf-picker";

type BookResult = {
  key: string;
  title: string;
  authors: string[] | null;
  publish_year: number | null;
  isbn: string[] | null;
  cover_url: string | null;
  edition_count: number;
  average_rating: number | null;
  rating_count: number;
  already_read_count: number;
  subjects: string[] | null;
};

type ScanResult = {
  isbn: string;
  book: BookResult;
};

type ScanMode = "camera" | "upload" | "manual";

export default function BookScanner({
  statusValues,
  statusKeyId,
  bookStatusMap,
}: {
  statusValues: StatusValue[];
  statusKeyId: string | null;
  bookStatusMap: Record<string, string>;
}) {
  const [mode, setMode] = useState<ScanMode>("upload");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [hint, setHint] = useState<string | null>(null);
  const [result, setResult] = useState<ScanResult | null>(null);
  const [scannedBooks, setScannedBooks] = useState<ScanResult[]>([]);
  const [hasBarcodeApi, setHasBarcodeApi] = useState(false);
  const [cameraActive, setCameraActive] = useState(false);
  const [manualIsbn, setManualIsbn] = useState("");

  const videoRef = useRef<HTMLVideoElement>(null);
  const streamRef = useRef<MediaStream | null>(null);
  const scanIntervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  // Check for BarcodeDetector API support
  useEffect(() => {
    if ("BarcodeDetector" in window) {
      setHasBarcodeApi(true);
      setMode("camera");
    }
  }, []);

  const stopCamera = useCallback(() => {
    if (scanIntervalRef.current) {
      clearInterval(scanIntervalRef.current);
      scanIntervalRef.current = null;
    }
    if (streamRef.current) {
      streamRef.current.getTracks().forEach((t) => t.stop());
      streamRef.current = null;
    }
    setCameraActive(false);
  }, []);

  // Cleanup camera on unmount
  useEffect(() => {
    return () => stopCamera();
  }, [stopCamera]);

  async function handleIsbnDetected(isbn: string) {
    // Skip if we already scanned this ISBN
    if (scannedBooks.some((b) => b.isbn === isbn)) return;

    setLoading(true);
    setError(null);
    setHint(null);

    try {
      const res = await fetch(`/api/books/lookup?isbn=${encodeURIComponent(isbn)}`);
      if (!res.ok) {
        const data = await res.json().catch(() => ({}));
        setError(data.error || "Book not found for this ISBN");
        setLoading(false);
        return;
      }
      const book: BookResult = await res.json();
      const scanResult: ScanResult = { isbn, book };
      setResult(scanResult);
      setScannedBooks((prev) => [...prev, scanResult]);
    } catch {
      setError("Failed to look up book");
    }
    setLoading(false);
  }

  async function startCamera() {
    setError(null);
    try {
      const stream = await navigator.mediaDevices.getUserMedia({
        video: { facingMode: "environment" },
      });
      streamRef.current = stream;
      if (videoRef.current) {
        videoRef.current.srcObject = stream;
        await videoRef.current.play();
      }
      setCameraActive(true);

      // eslint-disable-next-line
      const detector = new (window as any).BarcodeDetector({
        formats: ["ean_13", "ean_8"],
      });

      scanIntervalRef.current = setInterval(async () => {
        if (!videoRef.current || videoRef.current.readyState < 2) return;
        try {
          // eslint-disable-next-line
          const barcodes: any[] = await detector.detect(videoRef.current);
          for (const barcode of barcodes) {
            const isbn = barcode.rawValue;
            if (isbn && (isbn.length === 13 || isbn.length === 10)) {
              handleIsbnDetected(isbn);
              return;
            }
          }
        } catch {
          // Detection frame failed, continue
        }
      }, 500);
    } catch {
      setError("Could not access camera. Please check permissions or use file upload instead.");
    }
  }

  async function handleFileUpload(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file) return;

    setLoading(true);
    setError(null);
    setHint(null);
    setResult(null);

    const formData = new FormData();
    formData.append("image", file);

    try {
      const res = await fetch("/api/books/scan", {
        method: "POST",
        body: formData,
      });
      const data = await res.json();
      if (!res.ok) {
        setError(data.error || "Scan failed");
        if (data.hint) setHint(data.hint);
        setLoading(false);
        return;
      }
      const scanResult: ScanResult = { isbn: data.isbn, book: data.book };
      setResult(scanResult);
      setScannedBooks((prev) => [...prev, scanResult]);
    } catch {
      setError("Failed to upload image");
    }
    setLoading(false);

    // Reset file input
    e.target.value = "";
  }

  async function handleManualSubmit(e: React.FormEvent) {
    e.preventDefault();
    const isbn = manualIsbn.trim();
    if (!isbn) return;
    await handleIsbnDetected(isbn);
    setManualIsbn("");
  }

  function getOlId(key: string): string {
    return key.replace("/works/", "");
  }

  function getCurrentStatusValueId(book: BookResult): string | null {
    const olId = getOlId(book.key);
    return bookStatusMap[olId] ?? null;
  }

  return (
    <div className="space-y-6">
      {/* Mode selector */}
      <div className="flex gap-1 border-b border-border">
        {hasBarcodeApi && (
          <button
            onClick={() => {
              stopCamera();
              setMode("camera");
              setError(null);
            }}
            className={`px-4 py-2 text-sm font-medium border-b-2 -mb-px transition-colors ${
              mode === "camera"
                ? "border-accent text-text-primary"
                : "border-transparent text-text-primary hover:text-text-primary"
            }`}
          >
            Camera
          </button>
        )}
        <button
          onClick={() => {
            stopCamera();
            setMode("upload");
            setError(null);
          }}
          className={`px-4 py-2 text-sm font-medium border-b-2 -mb-px transition-colors ${
            mode === "upload"
              ? "border-accent text-text-primary"
              : "border-transparent text-text-primary hover:text-text-primary"
          }`}
        >
          Upload Photo
        </button>
        <button
          onClick={() => {
            stopCamera();
            setMode("manual");
            setError(null);
          }}
          className={`px-4 py-2 text-sm font-medium border-b-2 -mb-px transition-colors ${
            mode === "manual"
              ? "border-accent text-text-primary"
              : "border-transparent text-text-primary hover:text-text-primary"
          }`}
        >
          Enter ISBN
        </button>
      </div>

      {/* Camera mode */}
      {mode === "camera" && (
        <div className="space-y-4">
          <div className="relative bg-surface-2 rounded-lg overflow-hidden max-w-md aspect-[4/3]">
            <video
              ref={videoRef}
              className="w-full h-full object-cover"
              playsInline
              muted
            />
            {!cameraActive && (
              <div className="absolute inset-0 flex items-center justify-center">
                <button
                  onClick={startCamera}
                  className="px-4 py-2 bg-accent text-text-inverted text-sm rounded hover:bg-surface-3 transition-colors"
                >
                  Start Camera
                </button>
              </div>
            )}
            {cameraActive && (
              <div className="absolute inset-0 pointer-events-none border-2 border-dashed border-border rounded-lg m-4 flex items-center justify-center">
                <span className="bg-surface-0/80 text-text-primary text-xs px-2 py-1 rounded">
                  Point at barcode
                </span>
              </div>
            )}
          </div>
          {cameraActive && (
            <button
              onClick={stopCamera}
              className="text-xs text-text-primary hover:text-text-primary transition-colors"
            >
              Stop camera
            </button>
          )}
          {loading && (
            <p className="text-sm text-text-primary">Looking up book...</p>
          )}
        </div>
      )}

      {/* Upload mode */}
      {mode === "upload" && (
        <div className="space-y-4">
          <label className="block max-w-md">
            <div className="border-2 border-dashed border-border rounded-lg p-8 text-center hover:border-border transition-colors cursor-pointer">
              <svg
                className="mx-auto h-10 w-10 text-text-primary mb-3"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                strokeWidth={1.5}
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  d="M6.827 6.175A2.31 2.31 0 015.186 7.23c-.38.054-.757.112-1.134.175C2.999 7.58 2.25 8.507 2.25 9.574V18a2.25 2.25 0 002.25 2.25h15A2.25 2.25 0 0021.75 18V9.574c0-1.067-.75-1.994-1.802-2.169a47.865 47.865 0 00-1.134-.175 2.31 2.31 0 01-1.64-1.055l-.822-1.316a2.192 2.192 0 00-1.736-1.039 48.774 48.774 0 00-5.232 0 2.192 2.192 0 00-1.736 1.039l-.821 1.316z"
                />
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  d="M16.5 12.75a4.5 4.5 0 11-9 0 4.5 4.5 0 019 0zM18.75 10.5h.008v.008h-.008V10.5z"
                />
              </svg>
              <p className="text-sm text-text-primary mb-1">
                {loading ? "Scanning..." : "Upload a photo of the barcode"}
              </p>
              <p className="text-xs text-text-primary">JPEG, PNG, or GIF</p>
            </div>
            <input
              type="file"
              accept="image/*"
              capture="environment"
              onChange={handleFileUpload}
              disabled={loading}
              className="hidden"
            />
          </label>
        </div>
      )}

      {/* Manual ISBN mode */}
      {mode === "manual" && (
        <form onSubmit={handleManualSubmit} className="flex gap-2 max-w-md">
          <input
            type="text"
            value={manualIsbn}
            onChange={(e) => setManualIsbn(e.target.value)}
            placeholder="Enter ISBN (10 or 13 digits)"
            className="flex-1 px-3 py-2 text-sm border border-border rounded text-text-primary placeholder-text-tertiary focus:outline-none focus:ring-2 focus:ring-accent focus:border-transparent"
          />
          <button
            type="submit"
            disabled={loading || !manualIsbn.trim()}
            className="px-4 py-2 bg-accent text-text-inverted text-sm rounded hover:bg-surface-3 transition-colors disabled:opacity-50"
          >
            {loading ? "..." : "Look up"}
          </button>
        </form>
      )}

      {/* Error display */}
      {error && (
        <div className="bg-semantic-error-bg border border-semantic-error-border rounded-lg p-4 max-w-md">
          <p className="text-sm text-semantic-error">{error}</p>
          {hint && <p className="text-xs text-semantic-error mt-1">{hint}</p>}
        </div>
      )}

      {/* Current result */}
      {result && (
        <div className="border border-border rounded-lg p-4 max-w-lg">
          <div className="flex gap-4">
            {result.book.cover_url ? (
              <img
                src={result.book.cover_url}
                alt={result.book.title}
                className="w-16 h-24 object-cover rounded shadow-sm flex-shrink-0"
              />
            ) : (
              <div className="w-16 h-24 bg-surface-2 rounded flex-shrink-0 flex items-center justify-center text-text-primary text-xs">
                No cover
              </div>
            )}
            <div className="flex-1 min-w-0">
              <Link
                href={`/books/${getOlId(result.book.key)}`}
                className="text-sm font-medium text-text-primary hover:underline"
              >
                {result.book.title}
              </Link>
              {result.book.authors && result.book.authors.length > 0 && (
                <p className="text-xs text-text-primary mt-0.5">
                  {result.book.authors.join(", ")}
                </p>
              )}
              {result.book.publish_year && (
                <p className="text-xs text-text-primary mt-0.5">
                  {result.book.publish_year}
                </p>
              )}
              <p className="text-xs text-text-primary mt-1">
                ISBN: {result.isbn}
              </p>
              {statusValues.length > 0 && statusKeyId && (
                <div className="mt-2">
                  <StatusPicker
                    openLibraryId={getOlId(result.book.key)}
                    title={result.book.title}
                    coverUrl={result.book.cover_url}
                    statusValues={statusValues}
                    statusKeyId={statusKeyId}
                    currentStatusValueId={getCurrentStatusValueId(result.book)}
                  />
                </div>
              )}
            </div>
          </div>
        </div>
      )}

      {/* Previously scanned books */}
      {scannedBooks.length > 1 && (
        <div>
          <h2 className="text-sm font-medium text-text-primary mb-3">
            Scanned Books ({scannedBooks.length})
          </h2>
          <div className="space-y-3">
            {scannedBooks
              .filter((b) => b.isbn !== result?.isbn)
              .reverse()
              .map((scan) => (
                <div
                  key={scan.isbn}
                  className="flex items-center gap-3 border border-border rounded-lg p-3"
                >
                  {scan.book.cover_url ? (
                    <img
                      src={scan.book.cover_url}
                      alt={scan.book.title}
                      className="w-10 h-14 object-cover rounded shadow-sm flex-shrink-0"
                    />
                  ) : (
                    <div className="w-10 h-14 bg-surface-2 rounded flex-shrink-0" />
                  )}
                  <div className="flex-1 min-w-0">
                    <Link
                      href={`/books/${getOlId(scan.book.key)}`}
                      className="text-sm text-text-primary hover:underline truncate block"
                    >
                      {scan.book.title}
                    </Link>
                    {scan.book.authors && (
                      <p className="text-xs text-text-primary truncate">
                        {scan.book.authors.join(", ")}
                      </p>
                    )}
                  </div>
                  {statusValues.length > 0 && statusKeyId && (
                    <StatusPicker
                      openLibraryId={getOlId(scan.book.key)}
                      title={scan.book.title}
                      coverUrl={scan.book.cover_url}
                      statusValues={statusValues}
                      statusKeyId={statusKeyId}
                      currentStatusValueId={getCurrentStatusValueId(scan.book)}
                    />
                  )}
                </div>
              ))}
          </div>
        </div>
      )}
    </div>
  );
}
