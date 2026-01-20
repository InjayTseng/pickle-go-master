'use client';

import { useState, useCallback, useMemo } from 'react';
import { EventMap } from '@/components/map/EventMap';
import { SkillLevelFilterCompact, SkillLevel } from '@/components/map/SkillLevelFilter';
import { useEvents } from '@/hooks/useEvents';
import { useGeolocation } from '@/hooks/useGeolocation';
import { Button } from '@/components/ui/button';
import { Spinner } from '@/components/ui/spinner';
import { MapPin, List, RefreshCw } from 'lucide-react';

export default function HomePage() {
  const [skillFilter, setSkillFilter] = useState<SkillLevel>('all');
  const [mapCenter, setMapCenter] = useState<{ lat: number; lng: number } | null>(null);

  // Get user's geolocation
  const { coordinates, loading: geoLoading, error: geoError, refresh: refreshLocation } = useGeolocation();

  // Use map center if available, otherwise use geolocation
  const queryCenter = mapCenter || coordinates;

  // Fetch events based on location and filters
  const {
    data: eventsData,
    isLoading: eventsLoading,
    error: eventsError,
    refetch: refetchEvents,
  } = useEvents(
    {
      lat: queryCenter.lat,
      lng: queryCenter.lng,
      radius: 10000, // 10km
      skill_level: skillFilter === 'all' ? undefined : skillFilter,
      limit: 50,
    },
    !geoLoading // Only fetch when geolocation is ready
  );

  const events = useMemo(() => eventsData?.events || [], [eventsData]);

  const handleCenterChange = useCallback((center: { lat: number; lng: number }) => {
    setMapCenter(center);
  }, []);

  const handleSkillFilterChange = useCallback((value: SkillLevel) => {
    setSkillFilter(value);
  }, []);

  const handleRefresh = useCallback(() => {
    refreshLocation();
    refetchEvents();
  }, [refreshLocation, refetchEvents]);

  const isLoading = geoLoading || eventsLoading;

  return (
    <main className="flex flex-col h-[calc(100vh-64px)]">
      {/* Filter Bar */}
      <div className="bg-white border-b px-4 py-3 flex items-center gap-3 shadow-sm">
        <div className="flex-1 overflow-hidden">
          <SkillLevelFilterCompact
            value={skillFilter}
            onChange={handleSkillFilterChange}
          />
        </div>
        <Button
          variant="outline"
          size="icon"
          onClick={handleRefresh}
          disabled={isLoading}
          className="shrink-0"
        >
          {isLoading ? (
            <Spinner className="h-4 w-4" />
          ) : (
            <RefreshCw className="h-4 w-4" />
          )}
        </Button>
      </div>

      {/* Map Container */}
      <div className="flex-1 relative">
        {geoLoading ? (
          <div className="w-full h-full flex items-center justify-center bg-gray-100">
            <div className="text-center">
              <Spinner className="h-8 w-8 mx-auto mb-4" />
              <p className="text-muted-foreground">Getting your location...</p>
            </div>
          </div>
        ) : (
          <EventMap
            events={events}
            center={queryCenter}
            zoom={14}
            onCenterChange={handleCenterChange}
            loading={eventsLoading}
          />
        )}

        {/* Location Error Banner */}
        {geoError && (
          <div className="absolute top-2 left-2 right-2 bg-amber-50 border border-amber-200 rounded-lg px-4 py-2 text-sm text-amber-800 flex items-center gap-2">
            <MapPin className="h-4 w-4 shrink-0" />
            <span className="flex-1">{geoError}</span>
            <Button
              variant="ghost"
              size="sm"
              onClick={refreshLocation}
              className="shrink-0"
            >
              Retry
            </Button>
          </div>
        )}

        {/* Events Count Badge */}
        <div className="absolute top-2 right-2 bg-white rounded-full px-3 py-1 shadow-md text-sm font-medium">
          {isLoading ? (
            <Spinner className="h-4 w-4" />
          ) : (
            <span>{events.length} events nearby</span>
          )}
        </div>

        {/* Empty State */}
        {!isLoading && events.length === 0 && (
          <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
            <div className="bg-white rounded-lg shadow-lg p-6 text-center max-w-sm mx-4 pointer-events-auto">
              <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-gray-100">
                <MapPin className="h-6 w-6 text-gray-400" />
              </div>
              <h3 className="font-semibold text-lg mb-2">No events found</h3>
              <p className="text-muted-foreground text-sm mb-4">
                There are no events in this area yet.
                {skillFilter !== 'all' && ' Try changing the skill level filter.'}
              </p>
              {skillFilter !== 'all' && (
                <Button
                  variant="outline"
                  onClick={() => setSkillFilter('all')}
                >
                  Show all levels
                </Button>
              )}
            </div>
          </div>
        )}
      </div>
    </main>
  );
}
