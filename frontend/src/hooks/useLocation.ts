// src/hooks/useLocation.ts

import useSWR, { useSWRConfig } from 'swr';
import { api } from '../lib/api';
import { LocationResponse } from '../types/location';

export function useLocation() {
  const { data, error, mutate } = useSWR<LocationResponse>(
    '/api/v1/config/location',
    api.getLocation,
    {
      revalidateOnFocus: true,
      dedupingInterval: 30 * 1000,
    }
  );
  const { mutate: globalMutate } = useSWRConfig();

  /**
   * Invalidate every SWR cache key that depends on the user's location
   * (prayer times, fasting schedule, Eid countdown) so the UI updates
   * immediately after the location changes — no page reload needed.
   */
  const revalidateLocationDependent = () =>
    globalMutate(
      (key) =>
        typeof key === 'string' &&
        (key.startsWith('/api/v1/prayer-times') ||
          key.startsWith('/api/v1/fasting') ||
          key.startsWith('/api/v1/eid')),
      undefined,
      { revalidate: true }
    );

  const setGPS = async (latitude: number, longitude: number) => {
    await api.setLocationGPS(latitude, longitude);
    await mutate();
    await revalidateLocationDependent();
  };

  const setCity = async (city: string, country: string) => {
    await api.setLocationCity(city, country);
    await mutate();
    await revalidateLocationDependent();
  };

  return {
    location: data,
    isLoading: !error && !data,
    isError: error,
    mutate,
    setGPS,
    setCity,
  };
}
