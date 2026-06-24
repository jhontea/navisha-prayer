// src/components/prayer/NextPrayer.tsx

'use client';

import { PrayerTimesData } from '../../types/prayer';
import { formatTime12h, getPrayerIcon } from '../../lib/utils';
import { useCountdown } from '../../hooks/useCountdown';

interface NextPrayerProps {
  data: PrayerTimesData;
}

export default function NextPrayer({ data }: NextPrayerProps) {
  const nextPrayer = data.prayers.find((p) => p.is_next);

  if (!nextPrayer) {
    return (
      <div className="bg-dark-card border border-dark-border rounded-xl p-4 text-center">
        <p className="text-dark-textMuted">All prayers for today have passed</p>
      </div>
    );
  }

  // Calculate seconds until next prayer.
  const now = new Date();
  const currentSeconds = now.getHours() * 3600 + now.getMinutes() * 60 + now.getSeconds();
  const [h, m] = nextPrayer.time.split(':').map(Number);
  const prayerSeconds = h * 3600 + m * 60;
  // If the next prayer time is earlier in the clock than "now" (e.g. it's
  // evening and the next prayer is tomorrow's Fajr), it falls on the next day,
  // so add a full day. Without this the diff goes negative and the countdown
  // would clamp to 00:00:00.
  let diff = prayerSeconds - currentSeconds;
  if (diff < 0) diff += 24 * 3600;
  const remainingSeconds = diff;

  const countdown = useCountdown(remainingSeconds);

  return (
    <div className="bg-gradient-to-r from-islamic-900/60 via-teal-900/40 to-islamic-900/60 border border-islamic-700/40 rounded-2xl p-5 shadow-xl shadow-islamic-900/20">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <div className="w-14 h-14 rounded-full bg-islamic-800/50 flex items-center justify-center text-3xl">
            {getPrayerIcon(nextPrayer.name)}
          </div>
          <div>
            <p className="text-sm text-islamic-400 font-medium">Next Prayer</p>
            <h2 className="text-2xl font-bold text-dark-text">
              {nextPrayer.name}
            </h2>
            <p className="text-islamic-300 font-medium">
              {formatTime12h(nextPrayer.time)}
            </p>
          </div>
        </div>

        <div className="text-right">
          <p className="text-xs text-dark-textMuted mb-1">in</p>
          <div className="flex items-center gap-1">
            <span className="bg-islamic-800/60 text-islamic-300 font-mono text-lg font-bold px-2 py-1 rounded-lg">
              {countdown.hours}
            </span>
            <span className="text-islamic-500 font-bold">:</span>
            <span className="bg-islamic-800/60 text-islamic-300 font-mono text-lg font-bold px-2 py-1 rounded-lg">
              {countdown.minutes}
            </span>
            <span className="text-islamic-500 font-bold">:</span>
            <span className="bg-islamic-800/60 text-islamic-300 font-mono text-lg font-bold px-2 py-1 rounded-lg">
              {countdown.seconds}
            </span>
          </div>
        </div>
      </div>
    </div>
  );
}
