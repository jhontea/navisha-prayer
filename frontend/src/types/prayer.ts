// src/types/prayer.ts

export interface DateInfo {
  hijri: string;
  hijri_month_number: number;
  gregorian: string;
}

export interface PrayerEntry {
  name: string;
  time: string;
  is_next: boolean;
}

export interface PrayerTimesData {
  date: DateInfo;
  latitude: number;
  longitude: number;
  method: number;
  prayers: PrayerEntry[];
}

export interface PrayerTimesResponse {
  code: number;
  status: string;
  data: PrayerTimesData;
}

export interface MonthlyPrayerTimesResponse {
  code: number;
  status: string;
  data: PrayerTimesData[];
}

export interface NextPrayerResponse {
  code: number;
  status: string;
  data: PrayerEntry & {
    time_remaining_seconds: number;
  };
}
