"use client";

import { useEffect, useState } from "react";

interface KeyboardShortcutHintProps {
  keys: { mac: string; other: string };
  className?: string;
}

export default function KeyboardShortcutHint({
  keys,
  className,
}: KeyboardShortcutHintProps) {
  const [label, setLabel] = useState(keys.other);

  useEffect(() => {
    const isMac = /mac|iphone|ipad|ipod/i.test(navigator.userAgent);
    setLabel(isMac ? keys.mac : keys.other);
  }, [keys.mac, keys.other]);

  return <kbd className={className}>{label}</kbd>;
}
