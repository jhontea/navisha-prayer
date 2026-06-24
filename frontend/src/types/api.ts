// src/types/api.ts

export interface EidCountdownData {
  is_islamic_today: boolean;
  hijri_date: string;
  eid_date: string;
  eid_date_hijri: string;
  days_remaining: number;
  message: string;
}

export interface EidCountdownResponse {
  code: number;
  status: string;
  data: EidCountdownData;
}

export interface HealthResponse {
  code: number;
  status: string;
  data: {
    service: string;
    version: string;
  };
}

export interface ApiError {
  code: number;
  status: string;
  message: string;
}
