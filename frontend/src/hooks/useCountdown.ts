// src/hooks/useCountdown.ts

import { useState, useEffect, useCallback, useRef } from 'react';
import { formatCountdownHMS } from '../lib/utils';

interface CountdownResult {
  hours: string;
  minutes: string;
  seconds: string;
  totalSeconds: number;
  isExpired: boolean;
}

export function useCountdown(targetSeconds: number): CountdownResult {
  const calculateRemaining = useCallback(() => {
    return Math.max(0, targetSeconds);
  }, [targetSeconds]);

  const [remaining, setRemaining] = useState<number>(calculateRemaining);
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  useEffect(() => {
    setRemaining(calculateRemaining());
  }, [calculateRemaining]);

  useEffect(() => {
    if (remaining <= 0) return;

    intervalRef.current = setInterval(() => {
      setRemaining((prev) => {
        if (prev <= 1) {
          if (intervalRef.current) {
            clearInterval(intervalRef.current);
            intervalRef.current = null;
          }
          return 0;
        }
        return prev - 1;
      });
  }, 1000);

    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
    };
  }, [remaining]);

  return {
    ...formatCountdownHMS(remaining),
    totalSeconds: remaining,
    isExpired: remaining <= 0,
  };
}

/**
 * Hook that counts down to a specific target Date
 */
export function useCountdownTo(targetDate: Date): CountdownResult {
  const calculateRemaining = useCallback(() => {
    const now = new Date();
    const diff = Math.max(0, Math.floor((targetDate.getTime() - now.getTime()) / 1000));
    return diff;
  }, [targetDate]);

  const [remaining, setRemaining] = useState<number>(calculateRemaining);
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  useEffect(() => {
    setRemaining(calculateRemaining());
  }, [calculateRemaining]);

  useEffect(() => {
    if (remaining <= 0) return;

    intervalRef.current = setInterval(() => {
      setRemaining((prev) => {
        if (prev <= 1) {
          if (intervalRef.current) {
            clearInterval(intervalRef.current);
            intervalRef.current = null;
          }
          return 0;
        }
        return prev - 1;
      });
  }, 1000);

    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
    };
  }, [remaining]);

  return {
    ...formatCountdownHMS(remaining),
    totalSeconds: remaining,
    isExpired: remaining <= 0,
  };
}
