'use client';

import { useState, useEffect, useCallback } from 'react';

export interface GeolocationState {
  latitude: number | null;
  longitude: number | null;
  error: string | null;
  loading: boolean;
}

export interface GeolocationOptions {
  enableHighAccuracy?: boolean;
  timeout?: number;
  maximumAge?: number;
}

// Default center: Taipei 101
const DEFAULT_LATITUDE = 25.0330;
const DEFAULT_LONGITUDE = 121.5654;

export function useGeolocation(options: GeolocationOptions = {}) {
  const [state, setState] = useState<GeolocationState>({
    latitude: null,
    longitude: null,
    error: null,
    loading: true,
  });

  const {
    enableHighAccuracy = true,
    timeout = 10000,
    maximumAge = 60000, // Cache for 1 minute
  } = options;

  const getCurrentPosition = useCallback(() => {
    if (!navigator.geolocation) {
      setState({
        latitude: DEFAULT_LATITUDE,
        longitude: DEFAULT_LONGITUDE,
        error: 'Geolocation is not supported by this browser',
        loading: false,
      });
      return;
    }

    setState(prev => ({ ...prev, loading: true, error: null }));

    navigator.geolocation.getCurrentPosition(
      (position) => {
        setState({
          latitude: position.coords.latitude,
          longitude: position.coords.longitude,
          error: null,
          loading: false,
        });
      },
      (error) => {
        let errorMessage = 'Unable to get location';
        switch (error.code) {
          case error.PERMISSION_DENIED:
            errorMessage = 'Location permission denied';
            break;
          case error.POSITION_UNAVAILABLE:
            errorMessage = 'Location information unavailable';
            break;
          case error.TIMEOUT:
            errorMessage = 'Location request timed out';
            break;
        }
        // Fall back to default location
        setState({
          latitude: DEFAULT_LATITUDE,
          longitude: DEFAULT_LONGITUDE,
          error: errorMessage,
          loading: false,
        });
      },
      {
        enableHighAccuracy,
        timeout,
        maximumAge,
      }
    );
  }, [enableHighAccuracy, timeout, maximumAge]);

  useEffect(() => {
    getCurrentPosition();
  }, [getCurrentPosition]);

  const refresh = useCallback(() => {
    getCurrentPosition();
  }, [getCurrentPosition]);

  return {
    ...state,
    refresh,
    // Provide coordinates with fallback to default
    coordinates: {
      lat: state.latitude ?? DEFAULT_LATITUDE,
      lng: state.longitude ?? DEFAULT_LONGITUDE,
    },
    hasPermission: !state.error || !state.error.includes('permission'),
  };
}

// Hook for watching position changes
export function useWatchGeolocation(options: GeolocationOptions = {}) {
  const [state, setState] = useState<GeolocationState>({
    latitude: null,
    longitude: null,
    error: null,
    loading: true,
  });

  const {
    enableHighAccuracy = true,
    timeout = 10000,
    maximumAge = 0, // Always get fresh position
  } = options;

  useEffect(() => {
    if (!navigator.geolocation) {
      setState({
        latitude: DEFAULT_LATITUDE,
        longitude: DEFAULT_LONGITUDE,
        error: 'Geolocation is not supported by this browser',
        loading: false,
      });
      return;
    }

    const watchId = navigator.geolocation.watchPosition(
      (position) => {
        setState({
          latitude: position.coords.latitude,
          longitude: position.coords.longitude,
          error: null,
          loading: false,
        });
      },
      (error) => {
        let errorMessage = 'Unable to get location';
        switch (error.code) {
          case error.PERMISSION_DENIED:
            errorMessage = 'Location permission denied';
            break;
          case error.POSITION_UNAVAILABLE:
            errorMessage = 'Location information unavailable';
            break;
          case error.TIMEOUT:
            errorMessage = 'Location request timed out';
            break;
        }
        setState(prev => ({
          ...prev,
          latitude: prev.latitude ?? DEFAULT_LATITUDE,
          longitude: prev.longitude ?? DEFAULT_LONGITUDE,
          error: errorMessage,
          loading: false,
        }));
      },
      {
        enableHighAccuracy,
        timeout,
        maximumAge,
      }
    );

    return () => {
      navigator.geolocation.clearWatch(watchId);
    };
  }, [enableHighAccuracy, timeout, maximumAge]);

  return {
    ...state,
    coordinates: {
      lat: state.latitude ?? DEFAULT_LATITUDE,
      lng: state.longitude ?? DEFAULT_LONGITUDE,
    },
    hasPermission: !state.error || !state.error.includes('permission'),
  };
}
