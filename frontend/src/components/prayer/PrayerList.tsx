// src/components/prayer/PrayerList.tsx

'use client';

import { PrayerTimesData } from '../../types/prayer';
import PrayerCard from './PrayerCard';

interface PrayerListProps {
  data: PrayerTimesData;
}

export default function PrayerList({ data }: PrayerListProps) {
  // Calculate countdown seconds for each prayer
  const now = new Date();
  const currentSeconds = now.getHours() * 3600 + now.getMinutes() * 60 + now.getSeconds();

  function getCountdownSeconds(time: string): number | undefined {
    const [h, m] = time.split(':').map(Number);
    const prayerSeconds = h * 3600 + m * 60;
    const diff = prayerSeconds - currentSeconds;
    if (diff > 0) return diff;
    return undefined; // past prayer
  }

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-lg font-semibold text-dark-text">Prayer Times</h2>
        <span className="text-sm text-dark-textMuted">
          {data.date.hijri} Hijri
        </span>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
        {data.prayers.map((prayer) => (
          <PrayerCard
            key={prayer.name}
            prayer={prayer}
            countdownSeconds={getCountdownSeconds(prayer.time)}
          />
        ))}
      </div>
    </div>
  );
}
