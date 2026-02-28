import type { Metadata } from "next";
import { Inter, JetBrains_Mono } from "next/font/google";
import Link from "next/link";
import Nav from "@/components/nav";
import SearchFocusHandler from "@/components/search-focus-handler";
import { ToastProvider } from "@/components/toast";
import "./globals.css";

const inter = Inter({ subsets: ["latin"], variable: "--font-inter" });
const jetbrainsMono = JetBrains_Mono({
  subsets: ["latin"],
  variable: "--font-jetbrains-mono",
});

export const metadata: Metadata = {
  title: "rosslib",
  description: "Track your reading. Build your library.",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html
      lang="en"
      className={`${inter.variable} ${jetbrainsMono.variable}`}
      suppressHydrationWarning
    >
      <body
        className="font-sans antialiased bg-surface-0 text-text-primary"
        suppressHydrationWarning
      >
        <ToastProvider>
        <Nav />
        <SearchFocusHandler />
        <main className="max-w-shell mx-auto px-6 py-8">{children}</main>
        <footer className="border-t border-border">
          <div className="max-w-shell mx-auto px-6 py-4 flex items-center justify-between font-mono text-xs text-text-tertiary">
            <span>rosslib</span>
            <div className="flex items-center gap-4">
              <Link href="/feedback" className="hover:text-text-secondary transition-colors">feedback</Link>
              <span>better than goodreads</span>
            </div>
          </div>
        </footer>
        </ToastProvider>
      </body>
    </html>
  );
}
