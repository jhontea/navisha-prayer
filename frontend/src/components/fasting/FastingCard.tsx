// src/components/fasting/FastingCard.tsx

'use client';

import { FastingData } from '../../types/fasting';
import { formatTime12h, formatCountdown, calculateFastingProgress } from '../../lib/utils';
import { useCountdown } from '../../hooks/useCountdown';

interface FastingCardProps {
  data: FastingData;
  onToggleType: (type: string) => void;
}

export default function FastingCard({ data, onToggleType }: FastingCardProps) {
  const suhoorCountdown = useCountdown(data.suhoor_seconds);
  const iftarCountdown = useCountdown(data.iftar_seconds);

  const fastingTypes = [
    { key: 'ramadan', label: 'Ramadan', emoji: '🌙' },
    { key: 'syawal', label: 'Syawal', emoji: '📅' },
    { key: 'sawm', label: 'Sawm (Mon/Thu)', emoji: '☀️' },
  ];

  const progress = data.is_fasting
    ? calculateFastingProgress(data.suhoor_seconds, data.iftar_seconds)
    : 0;

  return (
    <div className="bg-dark-card border border-dark-border rounded-2xl p-5 space-y-5">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold text-dark-text flex items-center gap-2">
          <span>🕐</span> Fasting Schedule
        </h2>
        {data.is_fasting && (
          <span className="px-2 py-1 bg-islamic-900/60 text-islamic-400 text-xs font-medium rounded-full capitalize">
            {data.fasting_type}
          </span>
        )}
      </div>

      {/* Suhoor & Iftar times */}
      <div className="grid grid-cols-2 gap-4">
        {/* Suhoor */}
        <div className="bg-teal-900/20 border border-teal-800/30 rounded-xl p-4">
          <div className="flex items-center gap-2 mb-2">
            <span className="text-lg">🌅</span>
            <span className="text-sm text-dark-textMuted">Suhoor ends</span>
          </div>
          <p className="text-xl font-bold text-teal-300">
            {formatTime12h(data.suhoor_time)}
          </p>
          {data.suhoor_seconds > 0 ? (
            <p className="text-sm text-teal-400 mt-1">
              in {formatCountdown(data.suhoor_seconds)}
            </p>
          ) : (
            <p className="text-sm text-dark-textMuted mt-1">Passed</p>
          )}
        </div>

        {/* Iftar */}
        <div className="bg-islamic-900/20 border border-islamic-800/30 rounded-xl p-4">
          <div className="flex items-center gap-2 mb-2">
            <span className="text-lg">🌇</span>
            <span className="text-sm text-dark-textMuted">Iftar at</span>
          </div>
          <p className="text-xl font-bold text-islamic-300">
            {formatTime12h(data.iftar_time)}
          </p>
          {data.iftar_seconds > 0 ? (
            <p className="text-sm text-islamic-400 mt-1">
              in {formatCountdown(data.iftar_seconds)}
            </p>
          ) : (
            <p className="text-sm text-dark-textMuted mt-1">Passed</p>
          )}
        </div>
      </div>

      {/* Progress bar */}
      {data.is_fasting && (
        <div className="space-y-2">
          <div className="flex justify-between text-xs text-dark-textMuted">
            <span>Suhoor</span>
            <span>{Math.round(progress)}%</span>
            <span>Iftar</span>
          </div>
          <div className="h-2 bg-dark-border rounded-full overflow-hidden">
            <div
              className="h-full bg-gradient-to-r from-teal-500 to-islamic-500 rounded-full transition-all duration-1000"
              style={{ width: `${progress}%` }}
            />
          </div>
        </div>
      )}

      {/* Fasting toggles */}
      <div className="space-y-3 pt-2 border-t border-dark-border">
        <p className="text-sm text-dark-textMuted font-medium">Fasting Type</p>
        <div className="flex flex-wrap gap-2">
          {fastingTypes.map((type) => (
            <button
              key={type.key}
              onClick={() => onToggleType(type.key)}
              className={`
                flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-medium transition-all
                ${data.fasting_type === type.key && data.is_fasting
                  ? 'bg-islamic-800/60 text-islamic-300 border border-islamic-600/50'
                  : 'bg-dark-cardHover text-dark-textMuted border border-dark-border hover:border-islamic-800/50'
                }
              `}
            >
              <span>{type.emoji}</span>
              <span>{type.label}</span>
              {data.fasting_type === type.key && data.is_fasting && (
                <span className="w-2 h-2 rounded-full bg-islamic-400 animate-pulse" />
              )}
            </button>
          ))}
        </div>
      </div>
    </div>
  );
}
