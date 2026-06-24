// src/lib/api.ts

import type {
  PrayerTimesResponse,
  NextPrayerResponse,
  MonthlyPrayerTimesResponse,
} from '../types/prayer';
import type { FastingResponse } from '../types/fasting';
import type { EidCountdownResponse, HealthResponse } from '../types/api';
import type { LocationResponse, CitySearchResponse } from '../types/location';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

// Default request timeout (ms). Prevents the UI from hanging forever when the
// backend is slow, down, or unreachable.
const REQUEST_TIMEOUT = 12000;

/**
 * fetchWithTimeout wraps fetch with an AbortController so a single request can
 * never hang indefinitely. Throws on timeout, network failure, or non-2xx.
 */
async function fetchWithTimeout(
  url: string,
  options: RequestInit = {},
  timeout = REQUEST_TIMEOUT
): Promise<Response> {
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), timeout);

  try {
    return await fetch(url, { ...options, signal: controller.signal });
  } catch (err) {
    if (err instanceof DOMException && err.name === 'AbortError') {
      throw new Error(
        `Request to ${url} timed out after ${timeout}ms. Is the backend running?`
      );
    }
    // Network failure — backend is down or unreachable
    throw new Error(
      `Cannot reach the server at ${API_BASE_URL}. Is the backend running?`
    );
  } finally {
    clearTimeout(timer);
  }
}

export async function fetcher<T>(url: string): Promise<T> {
  const res = await fetchWithTimeout(`${API_BASE_URL}${url}`, {
    headers: {
      'Content-Type': 'application/json',
    },
  });

  if (!res.ok) {
    const error = new Error('An error occurred while fetching the data.');
    throw error;
  }

  return res.json() as Promise<T>;
}

/**
 * postJSON sends a POST request with a JSON body, guarded by the same timeout.
 */
async function postJSON<T>(path: string, body: unknown): Promise<T> {
  const res = await fetchWithTimeout(`${API_BASE_URL}${path}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });

  if (!res.ok) {
    throw new Error(`Request failed with status ${res.status}`);
  }

  return res.json() as Promise<T>;
}

export const api = {
  // Prayer Times
  getPrayerTimes: () => fetcher<PrayerTimesResponse>('/api/v1/prayer-times/today'),
  getPrayerTimesByDate: (date: string) =>
    fetcher<PrayerTimesResponse>(`/api/v1/prayer-times/date/${date}`),
  getMonthlyPrayerTimes: (month: number, year: number) =>
    fetcher<MonthlyPrayerTimesResponse>(
      `/api/v1/prayer-times/monthly?month=${month}&year=${year}`
    ),
  getNextPrayer: () => fetcher<NextPrayerResponse>('/api/v1/prayer-times/next'),

  // Fasting
  getFastingStatus: () => fetcher<FastingResponse>('/api/v1/fasting/status'),
  getFastingSchedule: () => fetcher<FastingResponse>('/api/v1/fasting/schedule'),
  updateFastingType: (type: string) =>
    postJSON('/api/v1/fasting/type', { type }),

  // Eid
  getEidCountdown: () => fetcher<EidCountdownResponse>('/api/v1/eid/countdown'),

  // Location
  getLocation: () => fetcher<LocationResponse>('/api/v1/config/location'),
  setLocationGPS: (latitude: number, longitude: number) =>
    postJSON('/api/v1/config/location/gps', { latitude, longitude }),
  setLocationCity: (city: string, country: string) =>
    postJSON('/api/v1/config/location/city', { city, country }),
  searchCities: async (query: string): Promise<CitySearchResponse> =>
    fetcher<CitySearchResponse>(`/api/v1/config/location/search?q=${encodeURIComponent(query)}`),

  // Health
  getHealth: () => fetcher<HealthResponse>('/api/v1/health'),
};
