// src/hooks/useEidCountdown.ts

import useSWR from 'swr';
import { api } from '../lib/api';
import { EidCountdownResponse } from '../types/api';

export function useEidCountdown() {
  return useSWR<EidCountdownResponse>('/api/v1/eid/countdown', api.getEidCountdown, {
    refreshInterval: 5 * 60 * 1000, // 5 minutes
    revalidateOnFocus: true,
    dedupingInterval: 60 * 1000, // 1 minute
  });
}
