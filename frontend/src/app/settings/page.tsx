// src/app/settings/page.tsx

'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';
import { useLocation } from '../../hooks/useLocation';
import { useFasting } from '../../hooks/useFasting';
import LocationSearch from '../../components/location/LocationSearch';
import GpsButton from '../../components/location/GpsButton';
import FastingToggle from '../../components/fasting/FastingToggle';
import LoadingSpinner from '../../components/ui/LoadingSpinner';
import ErrorMessage from '../../components/ui/ErrorMessage';
import { CityResult } from '../../types/location';
import { api } from '../../lib/api';

export default function SettingsPage() {
  const { location, isLoading, isError, setGPS, setCity, mutate } = useLocation();
  const { data: fastingData, mutate: mutateFasting } = useFasting();
  const [fastingType, setFastingType] = useState('none');
  const [fastingEnabled, setFastingEnabled] = useState(false);
  const [saved, setSaved] = useState(false);

  // Sync local fasting toggle state with the actual backend status
  useEffect(() => {
    if (fastingData?.data) {
      setFastingType(fastingData.data.fasting_type);
      setFastingEnabled(fastingData.data.is_fasting);
    }
  }, [fastingData]);

  const handleCitySelect = async (city: CityResult) => {
    try {
      await setCity(city.name, city.country);
      setSaved(true);
      setTimeout(() => setSaved(false), 2000);
    } catch (err) {
      console.error('Failed to set city:', err);
    }
  };

  const handleGpsLocation = async (lat: number, lon: number) => {
    try {
      await setGPS(lat, lon);
      setSaved(true);
      setTimeout(() => setSaved(false), 2000);
    } catch (err) {
      console.error('Failed to set GPS location:', err);
    }
  };

  const handleFastingToggle = async (type: string) => {
    try {
      if (fastingType === type && fastingEnabled) {
        // Disable
        await api.updateFastingType('none');
        setFastingEnabled(false);
        setFastingType('none');
      } else {
        await api.updateFastingType(type);
        setFastingType(type);
        setFastingEnabled(true);
      }
      mutateFasting();
    } catch (err) {
      console.error('Failed to update fasting type:', err);
    }
  };

  return (
    <div className="max-w-2xl mx-auto px-4 py-6 space-y-8 animate-fade-in">
      {/* Header */}
      <div className="flex items-center gap-4">
        <Link
          href="/"
          className="p-2 rounded-lg hover:bg-dark-cardHover text-dark-textMuted hover:text-islamic-400 transition-colors"
          aria-label="Back to home"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            className="h-5 w-5"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            strokeWidth={2}
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              d="M15 19l-7-7 7-7"
            />
          </svg>
        </Link>
        <h1 className="text-2xl font-bold text-dark-text">Settings</h1>
      </div>

      {/* Success message */}
      {saved && (
        <div className="bg-islamic-900/30 border border-islamic-700/40 rounded-xl p-4 text-center animate-fade-in">
          <p className="text-islamic-300 text-sm">✓ Settings saved successfully</p>
        </div>
      )}

      {/* Location Settings */}
      <section className="space-y-4">
        <h2 className="text-lg font-semibold text-dark-text flex items-center gap-2">
          <span>📍</span> Location
        </h2>

        {/* Current location display */}
        {isLoading ? (
          <LoadingSpinner />
        ) : isError ? (
          <ErrorMessage message="Failed to load location settings" />
        ) : (
          <div className="bg-dark-card border border-dark-border rounded-xl p-4">
            <p className="text-sm text-dark-textMuted mb-1">Current Location</p>
            <p className="text-dark-text font-medium">
              {location?.data
                ? `${location.data.city}, ${location.data.country}`
                : 'Not set'}
            </p>
            {location?.data?.source && (
              <p className="text-xs text-dark-textMuted mt-1">
                Source: {location.data.source}
              </p>
            )}
          </div>
        )}

        {/* City search */}
        <div className="space-y-3">
          <label className="text-sm text-dark-textMuted">Search by City</label>
          <LocationSearch onSelect={handleCitySelect} />
        </div>

        {/* GPS button */}
        <div className="space-y-3">
          <label className="text-sm text-dark-textMuted">Or use GPS</label>
          <GpsButton onLocation={handleGpsLocation} />
        </div>
      </section>

      {/* Fasting Settings */}
      <section className="space-y-4">
        <h2 className="text-lg font-semibold text-dark-text flex items-center gap-2">
          <span>🌙</span> Fasting Preferences
        </h2>
        <p className="text-sm text-dark-textMuted">
          Select the type of fasting you observe. This helps us show relevant
          schedules and reminders.
        </p>
        <FastingToggle
          currentType={fastingType}
          isEnabled={fastingEnabled}
          onToggle={handleFastingToggle}
        />
      </section>

      {/* About */}
      <section className="space-y-3 pt-4 border-t border-dark-border">
        <h2 className="text-lg font-semibold text-dark-text">About</h2>
        <div className="bg-dark-card border border-dark-border rounded-xl p-4 space-y-2">
          <p className="text-sm text-dark-textMuted">
            <span className="text-dark-text font-medium">Navisha Prayer</span> v1.0.0
          </p>
          <p className="text-xs text-dark-textMuted">
            Prayer times are calculated using the Aladhan API. Fasting schedules
            are derived from prayer times.
          </p>
          <p className="text-xs text-dark-textMuted">
            Default calculation method: Muslim World League (MWL)
          </p>
        </div>
      </section>
    </div>
  );
}
