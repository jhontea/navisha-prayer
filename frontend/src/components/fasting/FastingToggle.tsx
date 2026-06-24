// src/components/fasting/FastingToggle.tsx

'use client';

interface FastingToggleProps {
  currentType: string;
  isEnabled: boolean;
  onToggle: (type: string) => void;
}

const FASTING_TYPES = [
  { key: 'ramadan', label: 'Ramadan', description: 'Obligatory fast during Ramadan', emoji: '🌙' },
  { key: 'syawal', label: 'Syawal', description: '6 days of fasting in Syawal', emoji: '📅' },
  { key: 'sawm', label: 'Sawm', description: 'Monday & Thursday voluntary fast', emoji: '☀️' },
];

export default function FastingToggle({ currentType, isEnabled, onToggle }: FastingToggleProps) {
  return (
    <div className="space-y-3">
      {FASTING_TYPES.map((type) => {
        const isActive = currentType === type.key && isEnabled;
        return (
          <button
            key={type.key}
            onClick={() => onToggle(type.key)}
            role="button"
            aria-pressed={isActive}
            aria-label={`${type.label} fasting: ${isActive ? 'enabled' : 'disabled'}`}
            className={`
              w-full flex items-center justify-between p-4 rounded-xl border transition-all
              ${isActive
                ? 'bg-islamic-900/40 border-islamic-600/50'
                : 'bg-dark-card border-dark-border hover:border-islamic-800/40'
              }
            `}
          >
            <div className="flex items-center gap-3">
              <span aria-hidden="true" className="text-2xl">{type.emoji}</span>
              <div className="text-left">
                <p className={`font-medium ${isActive ? 'text-islamic-300' : 'text-dark-text'}`}>
                  {type.label}
                </p>
                <p className="text-xs text-dark-textMuted">{type.description}</p>
              </div>
            </div>

            <div
              role="switch"
              aria-checked={isActive}
              aria-hidden="true"
              className={`
                w-12 h-6 rounded-full transition-colors relative
                ${isActive ? 'bg-islamic-600' : 'bg-dark-border'}
              `}
            >
              <div
                className={`
                  w-5 h-5 rounded-full bg-white absolute top-0.5 transition-all
                  ${isActive ? 'left-6' : 'left-0.5'}
                `}
              />
            </div>
          </button>
        );
      })}
    </div>
  );
}
