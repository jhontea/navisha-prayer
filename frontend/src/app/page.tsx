// src/app/page.tsx

'use client';

import { usePrayerTimes } from '../hooks/usePrayerTimes';
import { useFasting } from '../hooks/useFasting';
import { useEidCountdown } from '../hooks/useEidCountdown';
import NextPrayer from '../components/prayer/NextPrayer';
import PrayerList from '../components/prayer/PrayerList';
import FastingCard from '../components/fasting/FastingCard';
import RamadanBanner from '../components/fasting/RamadanBanner';
import EidCountdown from '../components/eid/EidCountdown';
import LoadingSpinner from '../components/ui/LoadingSpinner';
import ErrorMessage from '../components/ui/ErrorMessage';
import { api } from '../lib/api';

export default function HomePage() {
  const { data: prayerData, error: prayerError, isLoading: prayerLoading, mutate: mutatePrayer } = usePrayerTimes();
  const { data: fastingData, error: fastingError, isLoading: fastingLoading, mutate: mutateFasting } = useFasting();
  const { data: eidData, error: eidError, isLoading: eidLoading } = useEidCountdown();

  const handleFastingToggle = async (type: string) => {
    try {
      await api.updateFastingType(type);
      mutateFasting();
      mutatePrayer();
    } catch (err) {
      console.error('Failed to update fasting type:', err);
    }
  };

  const hasError = prayerError || fastingError || eidError;
  const allLoading = prayerLoading && fastingLoading && eidLoading;

  if (allLoading) {
    return (
      <div className="max-w-4xl mx-auto px-4 py-8">
        <LoadingSpinner />
      </div>
    );
  }

  return (
    <div className="max-w-4xl mx-auto px-4 py-6 space-y-6 animate-fade-in">
      {/* Date header */}
      {prayerData?.data && (
        <div className="text-center mb-2">
          <p className="text-sm text-dark-textMuted">
            {new Date().toLocaleDateString('en-US', {
              weekday: 'long',
              year: 'numeric',
              month: 'long',
              day: 'numeric',
            })}
          </p>
          <p className="text-xs text-islamic-500/70 mt-0.5">
            {prayerData.data.date.hijri} Hijri
          </p>
        </div>
      )}

      {/* Error banner */}
      {hasError && (
        <ErrorMessage
          message="Some data could not be loaded. Showing cached data where available."
        />
      )}

      {/* Next Prayer Sticky */}
      {prayerData?.data && (
        <NextPrayer data={prayerData.data} />
      )}

      {/* Prayer List */}
      {prayerData?.data && (
        <PrayerList data={prayerData.data} />
      )}

      {/* Ramadan Banner */}
      {prayerData?.data && (
        <RamadanBanner dateInfo={prayerData.data.date} />
      )}

      {/* Fasting Section */}
      {fastingData?.data && (
        <FastingCard
          data={fastingData.data}
          onToggleType={handleFastingToggle}
        />
      )}

      {/* Eid Countdown */}
      {eidData?.data && (
        <EidCountdown data={eidData.data} />
      )}
    </div>
  );
}
