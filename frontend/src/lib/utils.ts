// src/lib/utils.ts

/**
 * Format seconds into human-readable countdown string e.g. "2h 15m"
 */
export function formatCountdown(totalSeconds: number): string {
  if (totalSeconds <= 0) return 'Now';

  const hours = Math.floor(totalSeconds / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);

  if (hours > 0 && minutes > 0) {
    return `${hours}h ${minutes}m`;
  } else if (hours > 0) {
    return `${hours}h`;
  } else {
    return `${minutes}m`;
  }
}

/**
 * Format seconds into "HH:MM:SS" string
 */
export function formatCountdownHMS(totalSeconds: number): {
  hours: string;
  minutes: string;
  seconds: string;
} {
  const h = Math.max(0, Math.floor(totalSeconds / 3600));
  const m = Math.max(0, Math.floor((totalSeconds % 3600) / 60));
  const s = Math.max(0, totalSeconds % 60);

  return {
    hours: h.toString().padStart(2, '0'),
    minutes: m.toString().padStart(2, '0'),
    seconds: s.toString().padStart(2, '0'),
  };
}

/**
 * Convert 24h time string "HH:MM" to 12h format "H:MM AM/PM"
 */
export function formatTime12h(time24: string): string {
  if (!time24 || !time24.includes(':')) return '--:--';
  const [hStr, mStr] = time24.split(':');
  const h = parseInt(hStr, 10);
  const m = parseInt(mStr, 10);
  if (Number.isNaN(h) || Number.isNaN(m)) return '--:--';
  const ampm = h >= 12 ? 'PM' : 'AM';
  const hour12 = h % 12 || 12;
  return `${hour12}:${m.toString().padStart(2, '0')} ${ampm}`;
}

/**
 * Get prayer icon based on prayer name
 */
export function getPrayerIcon(name: string): string {
  const icons: Record<string, string> = {
    Fajr: '🌅',
    Sunrise: '☀️',
    Dhuhr: '🌞',
    Asr: '🌤️',
    Maghrib: '🌇',
    Isha: '🌙',
  };
  return icons[name] || '🕌';
}

/**
 * Calculate progress percentage between suhoor and iftar
 */
export function calculateFastingProgress(
  suhoorSeconds: number,
  iftarSeconds: number
): number {
  const totalFastingSeconds = suhoorSeconds + iftarSeconds;
  if (totalFastingSeconds <= 0) return 100;
  const elapsed = totalFastingSeconds - iftarSeconds;
  return Math.min(100, Math.max(0, (elapsed / totalFastingSeconds) * 100));
}

/**
 * Format a date string for display
 */
export function formatDate(dateStr: string): string {
  try {
    const [year, month, day] = dateStr.split('-').map(Number);
    const date = new Date(year, month - 1, day);
    return date.toLocaleDateString('en-US', {
      weekday: 'long',
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    });
  } catch {
    return dateStr;
  }
}
