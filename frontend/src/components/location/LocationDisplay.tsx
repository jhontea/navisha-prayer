// src/components/location/LocationDisplay.tsx

'use client';

import { useLocation } from '../../hooks/useLocation';

export default function LocationDisplay() {
  const { location, isLoading, isError } = useLocation();

  if (isLoading) {
    return (
      <div className="hidden sm:flex items-center gap-1 text-dark-textMuted text-sm">
        <div className="w-4 h-4 rounded-full border-2 border-dark-border border-t-islamic-500 animate-spin" />
      </div>
    );
  }

  if (isError || !location?.data) {
    return (
      <div className="hidden sm:flex items-center gap-1 text-dark-textMuted text-sm">
        <span>📍</span>
        <span>Unknown</span>
      </div>
    );
  }

  const { city, country } = location.data;

  return (
    <div className="hidden sm:flex items-center gap-1 text-dark-textMuted text-sm">
      <span>📍</span>
      <span>{city}, {country}</span>
    </div>
  );
}
