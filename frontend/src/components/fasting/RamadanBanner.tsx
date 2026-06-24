// src/components/fasting/RamadanBanner.tsx

'use client';

import { DateInfo } from '../../types/prayer';

interface RamadanBannerProps {
  dateInfo: DateInfo;
}

export default function RamadanBanner({ dateInfo }: RamadanBannerProps) {
  const isRamadan = dateInfo.hijri_month_number === 9;

  if (!isRamadan) return null;

  return (
    <div className="bg-gradient-to-r from-islamic-800/40 via-teal-800/30 to-islamic-800/40 border border-islamic-600/30 rounded-2xl p-5 text-center animate-fade-in">
      <div className="text-3xl mb-2">🌙</div>
      <h2 className="text-xl font-bold text-islamic-300 mb-1">
        Ramadan Mubarak!
      </h2>
      <p className="text-dark-textMuted text-sm">
        May this blessed month bring you peace and blessings
      </p>
    </div>
  );
}
