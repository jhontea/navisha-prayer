// src/components/eid/EidCountdown.tsx

'use client';

import { EidCountdownData } from '../../types/api';
import { useCountdownTo } from '../../hooks/useCountdown';

interface EidCountdownProps {
  data: EidCountdownData;
}

export default function EidCountdown({ data }: EidCountdownProps) {
  const eidDate = new Date(data.eid_date);
  const countdown = useCountdownTo(eidDate);

  if (data.is_islamic_today) {
    return (
      <div className="bg-gradient-to-r from-islamic-800/50 via-teal-800/40 to-islamic-800/50 border border-islamic-500/40 rounded-2xl p-8 text-center animate-fade-in">
        <div className="text-5xl mb-4">🎉</div>
        <h2 className="text-3xl font-bold text-islamic-300 mb-2">
          Eid Mubarak!
        </h2>
        <p className="text-dark-textMuted">
          May this blessed day bring you joy and happiness
        </p>
        <p className="text-sm text-islamic-400 mt-2 font-arabic">
          عيد مبارك
        </p>
      </div>
    );
  }

  return (
    <div className="bg-dark-card border border-dark-border rounded-2xl p-6 text-center">
      <div className="flex items-center justify-center gap-2 mb-4">
        <span className="text-2xl">🌙</span>
        <h2 className="text-lg font-semibold text-dark-text">
          Eid al-Fitr Countdown
        </h2>
      </div>

      {/* Large day counter */}
      <div className="mb-4">
        <div className="inline-flex items-baseline gap-1">
          <span className="text-6xl font-bold bg-gradient-to-b from-islamic-400 to-teal-400 bg-clip-text text-transparent">
            {data.days_remaining}
          </span>
          <span className="text-xl text-dark-textMuted font-medium">
            {data.days_remaining === 1 ? 'day' : 'days'}
          </span>
        </div>
      </div>

      {/* Detailed countdown */}
      <div className="flex items-center justify-center gap-2 mb-4">
        <div className="bg-islamic-900/40 rounded-lg px-3 py-2 min-w-[50px]">
          <span className="text-xl font-bold text-islamic-300 font-mono">
            {countdown.hours}
          </span>
          <p className="text-xs text-dark-textMuted">hrs</p>
        </div>
        <span className="text-islamic-500 font-bold text-xl">:</span>
        <div className="bg-islamic-900/40 rounded-lg px-3 py-2 min-w-[50px]">
          <span className="text-xl font-bold text-islamic-300 font-mono">
            {countdown.minutes}
          </span>
          <p className="text-xs text-dark-textMuted">min</p>
        </div>
        <span className="text-islamic-500 font-bold text-xl">:</span>
        <div className="bg-islamic-900/40 rounded-lg px-3 py-2 min-w-[50px]">
          <span className="text-xl font-bold text-islamic-300 font-mono">
            {countdown.seconds}
          </span>
          <p className="text-xs text-dark-textMuted">sec</p>
        </div>
      </div>

      {/* Eid date info */}
      <div className="space-y-1">
        <p className="text-sm text-dark-textMuted">
          {eidDate.toLocaleDateString('en-US', {
            weekday: 'long',
            year: 'numeric',
            month: 'long',
            day: 'numeric',
          })}
        </p>
        <p className="text-xs text-islamic-500/70">
          {data.eid_date_hijri} Hijri
        </p>
      </div>
    </div>
  );
}


