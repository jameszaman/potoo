import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "Potoo",
  description: "Developer-first notification orchestration platform",
  icons: {
    icon: [
      { url: "/favicons/favicon-16x16.png", sizes: "16x16", type: "image/png" },
      { url: "/favicons/favicon-32x32.png", sizes: "32x32", type: "image/png" },
      { url: "/favicons/favicon-96x96.png", sizes: "96x96", type: "image/png" },
    ],
    apple: [
      { url: "/favicons/apple-touch-icon-120x120.png", sizes: "120x120" },
      { url: "/favicons/apple-touch-icon-152x152.png", sizes: "152x152" },
      { url: "/favicons/apple-touch-icon-167x167.png", sizes: "167x167" },
      { url: "/apple-icon.png", sizes: "180x180" },
    ],
    other: [
      { rel: "msapplication-TileImage", url: "/favicons/mstile-150x150.png" },
    ],
  },
  manifest: "/site.webmanifest",
  other: {
    "msapplication-TileColor": "#0F172A",
  },
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" className={`${geistSans.variable} ${geistMono.variable} h-full antialiased`}>
      <body className="min-h-full bg-gray-50 text-gray-900">{children}</body>
    </html>
  );
}
