// src/app/layout.tsx

import type { Metadata } from 'next';
import './globals.css';
import Header from '../components/ui/Header';
import Footer from '../components/ui/Footer';
import LocationDetector from '../components/location/LocationDetector';

export const metadata: Metadata = {
  title: 'Navisha Prayer - Prayer Times, Fasting & Eid Countdown',
  description:
    'A modern Islamic web application for prayer times, fasting schedules, and Eid al-Fitr countdown.',
  keywords: ['prayer times', 'islam', 'fasting', 'ramadan', 'eid', 'qibla'],
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en" className="dark">
      <body className="min-h-screen flex flex-col bg-dark-bg text-dark-text antialiased">
        <LocationDetector />
        <Header />
        <main className="flex-1">{children}</main>
        <Footer />
      </body>
    </html>
  );
}
