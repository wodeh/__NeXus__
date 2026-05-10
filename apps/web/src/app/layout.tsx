import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "NeXus PMS — Property Management System",
  description: "Hospitality property management platform",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
