import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Sideby | Find your people",
  description: "A thoughtful student matching and chat space.",
};

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
