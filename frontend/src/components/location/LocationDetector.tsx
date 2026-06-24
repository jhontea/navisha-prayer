// src/components/location/LocationDetector.tsx

'use client';

import { useEffect, useRef } from 'react';
import { useSWRConfig } from 'swr';
import { api } from '../../lib/api';

const AUTO_DETECT_KEY = 'navisha:gps-auto-detect';

/**
 * LocationDetector attempts to resolve the user's real location on first load.
 * 
 * Strategy:
 * 1. Try IP-based geolocation (works on desktop, no permission needed)
 * 2. Fall back to browser Geolocation API (more accurate on mobile)
 * 
 * Only triggers when the stored location is still the server "default"
 * (i.e. the user has never set GPS or a city). This prevents overriding a
 * location the user explicitly chose. The attempt is also recorded in
 * sessionStorage so we don't re-prompt repeatedly within the same session.
 * 
 * Renders nothing.
 */
export default function LocationDetector() {
  const hasRun = useRef(false);
  const { mutate } = useSWRConfig();

  useEffect(() => {
    if (hasRun.current) return;
    hasRun.current = true;

    // Skip if we've already attempted auto-detection this session.
    if (sessionStorage.getItem(AUTO_DETECT_KEY) === 'done') return;

    const detect = async () => {
      try {
        const location = await api.getLocation();

        // Only auto-detect when the location has never been set by the user.
        if (location?.data?.source && location.data.source !== 'default') {
          sessionStorage.setItem(AUTO_DETECT_KEY, 'done');
          return;
        }

        // Strategy 1: Try IP-based geolocation (works on desktop, no permission needed)
        try {
          const ipController = new AbortController();
          const ipTimer = setTimeout(() => ipController.abort(), 8000);
          const ipResponse = await fetch('https://ipapi.co/json/', {
            signal: ipController.signal,
          }).finally(() => clearTimeout(ipTimer));
          if (ipResponse.ok) {
            const ipData = await ipResponse.json();
            if (ipData.latitude && ipData.longitude) {
              console.log('IP geolocation:', ipData.city, ipData.latitude, ipData.longitude);
              await api.setLocationGPS(ipData.latitude, ipData.longitude);
              mutate(() => true);
              sessionStorage.setItem(AUTO_DETECT_KEY, 'done');
              return;
            }
          }
        } catch (err) {
          console.warn('IP geolocation failed, trying browser geolocation:', err);
        }

        // Strategy 2: Fall back to browser Geolocation API (more accurate on mobile)
        if (typeof navigator !== 'undefined' && navigator.geolocation) {
          navigator.geolocation.getCurrentPosition(
            async (position) => {
              try {
                await api.setLocationGPS(
                  position.coords.latitude,
                  position.coords.longitude
                );
                mutate(() => true);
              } catch (err) {
                console.error('Auto GPS: failed to save location', err);
              } finally {
                sessionStorage.setItem(AUTO_DETECT_KEY, 'done');
              }
            },
            (err) => {
              console.warn('Auto GPS: could not get position', err.message);
              sessionStorage.setItem(AUTO_DETECT_KEY, 'done');
            },
            {
              enableHighAccuracy: true,
              timeout: 10000,
              maximumAge: 300000, // 5 minutes
            }
          );
        } else {
          sessionStorage.setItem(AUTO_DETECT_KEY, 'done');
        }
      } catch (err) {
        console.error('Auto GPS: failed to read location', err);
        sessionStorage.setItem(AUTO_DETECT_KEY, 'done');
      }
    };

    detect();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return null;
}
