// src/types/fasting.ts

export interface FastingData {
  is_fasting: boolean;
  fasting_type: string;
  suhoor_time: string;
  iftar_time: string;
  suhoor_in: string;
  iftar_in: string;
  suhoor_seconds: number;
  iftar_seconds: number;
}

export interface FastingResponse {
  code: number;
  status: string;
  data: FastingData;
}
