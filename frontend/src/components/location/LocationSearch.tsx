// src/components/location/LocationSearch.tsx

'use client';

import { useState, useCallback, useRef, useEffect } from 'react';
import { api } from '../../lib/api';
import { CityResult } from '../../types/location';

interface LocationSearchProps {
  onSelect: (city: CityResult) => void;
}

export default function LocationSearch({ onSelect }: LocationSearchProps) {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<CityResult[]>([]);
  const [isSearching, setIsSearching] = useState(false);
  const [showDropdown, setShowDropdown] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Close dropdown when clicking outside
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setShowDropdown(false);
      }
    }
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  // Clean up any pending debounce timer on unmount
  useEffect(() => {
    return () => {
      if (debounceRef.current) clearTimeout(debounceRef.current);
    };
  }, []);

  const runSearch = useCallback(async (value: string) => {
    setIsSearching(true);
    try {
      const response = await api.searchCities(value);
      if (response && Array.isArray(response.data)) {
        setResults(response.data);
      } else {
        setResults([]);
      }
      setShowDropdown(true);
    } catch {
      setResults([]);
    } finally {
      setIsSearching(false);
    }
  }, []);

  const handleSearch = useCallback(
    (value: string) => {
      setQuery(value);

      // Cancel the previous pending request
      if (debounceRef.current) clearTimeout(debounceRef.current);

      if (value.length < 2) {
        setResults([]);
        setShowDropdown(false);
        setIsSearching(false);
        return;
      }

      // Debounce the API call by 300ms to avoid spamming on each keystroke
      debounceRef.current = setTimeout(() => {
        runSearch(value);
      }, 300);
    },
    [runSearch]
  );

  const handleSelect = (city: CityResult) => {
    setQuery(`${city.name}, ${city.country}`);
    setShowDropdown(false);
    onSelect(city);
  };

  return (
    <div className="relative" ref={dropdownRef}>
      <div className="relative">
        <span className="absolute left-3 top-1/2 -translate-y-1/2 text-dark-textMuted">
          🔍
        </span>
        <input
          type="text"
          value={query}
          onChange={(e) => handleSearch(e.target.value)}
          onFocus={() => results.length > 0 && setShowDropdown(true)}
          placeholder="Search city..."
          className="w-full pl-10 pr-4 py-3 bg-dark-card border border-dark-border rounded-xl text-dark-text placeholder-dark-textMuted focus:outline-none focus:border-islamic-600/50 focus:ring-1 focus:ring-islamic-600/30 transition-all"
        />
        {isSearching && (
          <div className="absolute right-3 top-1/2 -translate-y-1/2">
            <div className="w-4 h-4 rounded-full border-2 border-dark-border border-t-islamic-500 animate-spin" />
          </div>
        )}
      </div>

      {/* Dropdown */}
      {showDropdown && results.length > 0 && (
        <div className="absolute z-50 w-full mt-2 bg-dark-card border border-dark-border rounded-xl shadow-xl max-h-60 overflow-y-auto animate-slide-up">
          {results.map((city, index) => (
            <button
              key={`${city.name}-${city.country}-${index}`}
              onClick={() => handleSelect(city)}
              className="w-full px-4 py-3 text-left hover:bg-dark-cardHover transition-colors flex items-center gap-3 border-b border-dark-border last:border-b-0"
            >
              <span className="text-lg">📍</span>
              <div>
                <p className="text-dark-text font-medium">{city.name}</p>
                <p className="text-xs text-dark-textMuted">{city.country}</p>
              </div>
            </button>
          ))}
        </div>
      )}

      {showDropdown && query.length >= 2 && results.length === 0 && !isSearching && (
        <div className="absolute z-50 w-full mt-2 bg-dark-card border border-dark-border rounded-xl p-4 text-center">
          <p className="text-dark-textMuted text-sm">No cities found</p>
        </div>
      )}
    </div>
  );
}
