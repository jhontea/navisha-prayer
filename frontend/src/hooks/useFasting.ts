// src/hooks/useFasting.ts

import useSWR from 'swr';
import { api } from '../lib/api';
import { FastingResponse } from '../types/fasting';

export function useFasting() {
  return useSWR<FastingResponse>('/api/v1/fasting/status', api.getFastingStatus, {
    refreshInterval: 60 * 1000, // 1 minute
    revalidateOnFocus: true,
    dedupingInterval: 30 * 1000, // 30 seconds
  });
}
