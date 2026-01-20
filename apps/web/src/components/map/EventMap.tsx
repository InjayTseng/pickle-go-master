'use client';

import React, { useState, useCallback, useRef, useEffect } from 'react';
import { GoogleMap, useLoadScript } from '@react-google-maps/api';
import { Event } from '@/lib/api-client';
import { EventPin } from './EventPin';
import { EventCard } from './EventCard';
import { Spinner } from '@/components/ui/spinner';

interface EventMapProps {
  events: Event[];
  center: { lat: number; lng: number };
  zoom?: number;
  onBoundsChange?: (bounds: google.maps.LatLngBounds) => void;
  onCenterChange?: (center: { lat: number; lng: number }) => void;
  loading?: boolean;
}

const libraries: ('places' | 'geometry')[] = ['places'];

const mapContainerStyle = {
  width: '100%',
  height: '100%',
};

const defaultMapOptions: google.maps.MapOptions = {
  disableDefaultUI: true,
  zoomControl: true,
  zoomControlOptions: {
    position: typeof google !== 'undefined' ? google.maps.ControlPosition.RIGHT_CENTER : 6,
  },
  mapTypeControl: false,
  streetViewControl: false,
  fullscreenControl: false,
  gestureHandling: 'greedy', // Enables one-finger scroll on mobile
  styles: [
    {
      featureType: 'poi',
      elementType: 'labels',
      stylers: [{ visibility: 'off' }],
    },
  ],
};

export function EventMap({
  events,
  center,
  zoom = 14,
  onBoundsChange,
  onCenterChange,
  loading = false,
}: EventMapProps) {
  const [selectedEvent, setSelectedEvent] = useState<Event | null>(null);
  const mapRef = useRef<google.maps.Map | null>(null);

  const { isLoaded, loadError } = useLoadScript({
    googleMapsApiKey: process.env.NEXT_PUBLIC_GOOGLE_MAPS_API_KEY || '',
    libraries,
  });

  const onLoad = useCallback((map: google.maps.Map) => {
    mapRef.current = map;
  }, []);

  const onUnmount = useCallback(() => {
    mapRef.current = null;
  }, []);

  const handleMapClick = useCallback(() => {
    setSelectedEvent(null);
  }, []);

  const handleBoundsChanged = useCallback(() => {
    if (mapRef.current && onBoundsChange) {
      const bounds = mapRef.current.getBounds();
      if (bounds) {
        onBoundsChange(bounds);
      }
    }
  }, [onBoundsChange]);

  const handleCenterChanged = useCallback(() => {
    if (mapRef.current && onCenterChange) {
      const center = mapRef.current.getCenter();
      if (center) {
        onCenterChange({ lat: center.lat(), lng: center.lng() });
      }
    }
  }, [onCenterChange]);

  const handlePinClick = useCallback((event: Event) => {
    setSelectedEvent(event);
    // Center map on the selected event
    if (mapRef.current) {
      mapRef.current.panTo({ lat: event.location.lat, lng: event.location.lng });
    }
  }, []);

  const handleCardClose = useCallback(() => {
    setSelectedEvent(null);
  }, []);

  // Close card when clicking outside
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        setSelectedEvent(null);
      }
    };
    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, []);

  if (loadError) {
    return (
      <div className="w-full h-full flex items-center justify-center bg-gray-100">
        <div className="text-center text-muted-foreground">
          <p className="text-lg font-medium">Unable to load map</p>
          <p className="text-sm">Please check your internet connection</p>
        </div>
      </div>
    );
  }

  if (!isLoaded) {
    return (
      <div className="w-full h-full flex items-center justify-center bg-gray-100">
        <Spinner className="h-8 w-8" />
      </div>
    );
  }

  return (
    <div className="relative w-full h-full">
      <GoogleMap
        mapContainerStyle={mapContainerStyle}
        center={center}
        zoom={zoom}
        options={defaultMapOptions}
        onLoad={onLoad}
        onUnmount={onUnmount}
        onClick={handleMapClick}
        onBoundsChanged={handleBoundsChanged}
        onCenterChanged={handleCenterChanged}
      >
        {/* Event Pins */}
        {events.map((event) => (
          <EventPin
            key={event.id}
            event={event}
            onClick={handlePinClick}
            isSelected={selectedEvent?.id === event.id}
          />
        ))}
      </GoogleMap>

      {/* Loading overlay */}
      {loading && (
        <div className="absolute inset-0 bg-white/50 flex items-center justify-center pointer-events-none">
          <Spinner className="h-8 w-8" />
        </div>
      )}

      {/* Selected Event Card - Mobile Bottom Sheet Style */}
      {selectedEvent && (
        <div className="absolute bottom-4 left-4 right-4 z-10 md:bottom-auto md:top-4 md:left-auto md:right-4 md:w-80">
          <EventCard
            event={selectedEvent}
            onClose={handleCardClose}
            compact
          />
        </div>
      )}

      {/* Current Location Button */}
      <button
        className="absolute bottom-20 right-4 md:bottom-4 bg-white rounded-full p-3 shadow-lg hover:bg-gray-50 active:bg-gray-100 transition-colors"
        onClick={() => {
          if (navigator.geolocation) {
            navigator.geolocation.getCurrentPosition(
              (position) => {
                const newCenter = {
                  lat: position.coords.latitude,
                  lng: position.coords.longitude,
                };
                if (mapRef.current) {
                  mapRef.current.panTo(newCenter);
                }
                onCenterChange?.(newCenter);
              },
              () => {
                // Silently fail
              }
            );
          }
        }}
        aria-label="Go to my location"
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="24"
          height="24"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
        >
          <circle cx="12" cy="12" r="3" />
          <path d="M12 2v2" />
          <path d="M12 20v2" />
          <path d="M2 12h2" />
          <path d="M20 12h2" />
        </svg>
      </button>
    </div>
  );
}

export default EventMap;
