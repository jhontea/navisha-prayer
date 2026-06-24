// src/components/location/GpsButton.tsx

'use client';

import { useState } from 'react';

interface GpsButtonProps {
  onLocation: (lat: number, lon: number) => void;
}

export default function GpsButton({ onLocation }: GpsButtonProps) {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleClick = () => {
    if (!navigator.geolocation) {
      setError('Geolocation is not supported by your browser');
      return;
    }

    setIsLoading(true);
    setError(null);

    navigator.geolocation.getCurrentPosition(
      (position) => {
        onLocation(position.coords.latitude, position.coords.longitude);
        setIsLoading(false);
      },
      (err) => {
        switch (err.code) {
          case err.PERMISSION_DENIED:
            setError('Location permission denied');
            break;
          case err.POSITION_UNAVAILABLE:
            setError('Location unavailable');
            break;
          case err.TIMEOUT:
            setError('Location request timed out');
            break;
          default:
            setError('Failed to get location');
        }
        setIsLoading(false);
      },
      {
        enableHighAccuracy: true,
        timeout: 10000,
        maximumAge: 300000, // 5 minutes
      }
    );
  };

  return (
    <div>
      <button
        onClick={handleClick}
        disabled={isLoading}
        className="w-full flex items-center justify-center gap-2 px-4 py-3 bg-teal-900/30 border border-teal-800/40 rounded-xl text-teal-300 hover:bg-teal-900/50 hover:border-teal-700/50 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
      >
        {isLoading ? (
          <>
            <div className="w-4 h-4 rounded-full border-2 border-teal-700 border-t-teal-300 animate-spin" />
            <span>Getting location...</span>
          </>
        ) : (
          <>
            <span className="text-lg">📡</span>
            <span>Use My Location</span>
          </>
        )}
      </button>
      {error && (
        <p className="text-xs text-red-400 mt-2 text-center">{error}</p>
      )}
    </div>
  );
}
