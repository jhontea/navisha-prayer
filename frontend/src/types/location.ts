// src/types/location.ts

export interface LocationData {
  latitude: number;
  longitude: number;
  city: string;
  country: string;
  source: string;
}

export interface LocationResponse {
  code: number;
  status: string;
  data: LocationData;
}

export interface CityResult {
  name: string;
  country: string;
  latitude: number;
  longitude: number;
}

export interface CitySearchResponse {
  code: number;
  status: string;
  data: CityResult[];
}
