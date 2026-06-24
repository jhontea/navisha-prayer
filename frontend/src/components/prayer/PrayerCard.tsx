// src/components/prayer/PrayerCard.tsx

import { PrayerEntry } from '../../types/prayer';
import { formatTime12h, getPrayerIcon } from '../../lib/utils';

interface PrayerCardProps {
  prayer: PrayerEntry;
  countdownSeconds?: number;
}

export default function PrayerCard({ prayer, countdownSeconds }: PrayerCardProps) {
  const isNext = prayer.is_next;
  const isPast = !isNext && countdownSeconds === undefined;

  return (
    <div
      role="article"
      aria-label={`${prayer.name} prayer at ${formatTime12h(prayer.time)}`}
      className={`
        relative rounded-xl p-4 transition-all duration-300 animate-slide-up
        ${isNext
          ? 'bg-gradient-to-br from-islamic-900/80 to-teal-900/60 border-2 border-islamic-500/50 shadow-lg shadow-islamic-500/10'
          : 'bg-dark-card border border-dark-border hover:border-islamic-800/50'
        }
      `}
    >
      {isNext && (
        <div aria-hidden="true" className="absolute -top-2 left-3 px-2 py-0.5 bg-islamic-600 text-white text-xs font-medium rounded-full">
          Next
        </div>
      )}

      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <span className="text-2xl">{getPrayerIcon(prayer.name)}</span>
          <div>
            <h3
              className={`font-semibold ${
                isNext ? 'text-islamic-300' : 'text-dark-text'
              }`}
            >
              {prayer.name}
            </h3>
            {isPast && (
              <span className="text-xs text-islamic-500">✓ Past</span>
            )}
          </div>
        </div>

        <div className="text-right">
          <p
            className={`text-lg font-bold ${
              isNext ? 'text-islamic-400' : 'text-dark-text'
            }`}
          >
            {formatTime12h(prayer.time)}
          </p>
          {countdownSeconds !== undefined && !isNext && countdownSeconds > 0 && (
            <p className="text-xs text-dark-textMuted mt-0.5">
              in {Math.floor(countdownSeconds / 3600)}h{' '}
              {Math.floor((countdownSeconds % 3600) / 60)}m
            </p>
          )}
        </div>
      </div>
    </div>
  );
}
