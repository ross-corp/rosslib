"use client";

import { useState, useRef, useEffect } from "react";
import Link from "next/link";

interface NavDropdownItem {
  href: string;
  label: string;
}

export default function NavDropdown({
  label,
  items,
}: {
  label: string;
  items: NavDropdownItem[];
}) {
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    function handleClick(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setOpen(false);
      }
    }
    document.addEventListener("mousedown", handleClick);
    return () => document.removeEventListener("mousedown", handleClick);
  }, []);

  return (
    <div ref={ref} className="relative group">
      <button
        onClick={() => setOpen((o) => !o)}
        className="nav-link flex items-center gap-1"
      >
        {label}
        <svg
          xmlns="http://www.w3.org/2000/svg"
          viewBox="0 0 20 20"
          fill="currentColor"
          className="w-3 h-3 transition-transform group-hover:rotate-180"
          style={open ? { transform: "rotate(180deg)" } : undefined}
        >
          <path
            fillRule="evenodd"
            d="M5.22 8.22a.75.75 0 011.06 0L10 11.94l3.72-3.72a.75.75 0 111.06 1.06l-4.25 4.25a.75.75 0 01-1.06 0L5.22 9.28a.75.75 0 010-1.06z"
            clipRule="evenodd"
          />
        </svg>
      </button>
      <div
        className={`absolute top-full left-0 mt-1 min-w-[140px] bg-surface-1 border border-border rounded shadow-lg z-50 py-1 ${
          open ? "block" : "hidden group-hover:block"
        }`}
      >
        {items.map((item) => (
          <Link
            key={item.href}
            href={item.href}
            className="block px-3 py-1.5 font-mono text-sm text-text-secondary hover:text-text-primary hover:bg-surface-2 transition-colors"
            onClick={() => setOpen(false)}
          >
            {item.label}
          </Link>
        ))}
      </div>
    </div>
  );
}
