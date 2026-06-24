// src/hooks/usePrayerTimes.ts

import useSWR from 'swr';
import { api } from '../lib/api';
import { PrayerTimesResponse } from '../types/prayer';

export function usePrayerTimes() {
  return useSWR<PrayerTimesResponse>('/api/v1/prayer-times/today', api.getPrayerTimes, {
    refreshInterval: 5 * 60 * 1000, // 5 minutes
    revalidateOnFocus: true,
    dedupingInterval: 60 * 1000, // 1 minute
  });
}
