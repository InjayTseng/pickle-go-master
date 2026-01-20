'use client';

import { useQuery, useMutation, useQueryClient, keepPreviousData } from '@tanstack/react-query';
import { apiClient, Event, EventListResponse, RegistrationResponse } from '@/lib/api-client';

export interface EventQueryParams {
  lat?: number;
  lng?: number;
  radius?: number;
  skill_level?: string;
  status?: string;
  limit?: number;
  offset?: number;
}

// Cache time constants (in milliseconds)
const CACHE_TIMES = {
  // Map pins data - frequently accessed, can be slightly stale
  MAP_PINS: {
    staleTime: 60 * 1000, // 1 minute - data is considered fresh
    gcTime: 10 * 60 * 1000, // 10 minutes - keep in cache for background refetch
  },
  // Event details - need to be more up-to-date
  EVENT_DETAIL: {
    staleTime: 30 * 1000, // 30 seconds
    gcTime: 5 * 60 * 1000, // 5 minutes
  },
  // Registrations - need real-time accuracy
  REGISTRATIONS: {
    staleTime: 10 * 1000, // 10 seconds
    gcTime: 2 * 60 * 1000, // 2 minutes
  },
  // User data - can be more aggressively cached
  USER_DATA: {
    staleTime: 2 * 60 * 1000, // 2 minutes
    gcTime: 10 * 60 * 1000, // 10 minutes
  },
};

// Hook for fetching events list with geo filter (Map Pins)
export function useEvents(params: EventQueryParams = {}, enabled = true) {
  return useQuery({
    queryKey: ['events', params],
    queryFn: () => apiClient.listEvents(params),
    enabled,
    staleTime: CACHE_TIMES.MAP_PINS.staleTime,
    gcTime: CACHE_TIMES.MAP_PINS.gcTime,
    refetchOnWindowFocus: false,
    // Keep previous data while fetching new data (smooth UX for map)
    placeholderData: keepPreviousData,
  });
}

// Hook for fetching events with prefetching for map viewport
export function useEventsWithPrefetch(params: EventQueryParams = {}, enabled = true) {
  const queryClient = useQueryClient();

  const query = useQuery({
    queryKey: ['events', params],
    queryFn: () => apiClient.listEvents(params),
    enabled,
    staleTime: CACHE_TIMES.MAP_PINS.staleTime,
    gcTime: CACHE_TIMES.MAP_PINS.gcTime,
    refetchOnWindowFocus: false,
    placeholderData: keepPreviousData,
  });

  // Prefetch adjacent areas when map moves
  const prefetchAdjacentArea = async (newParams: EventQueryParams) => {
    await queryClient.prefetchQuery({
      queryKey: ['events', newParams],
      queryFn: () => apiClient.listEvents(newParams),
      staleTime: CACHE_TIMES.MAP_PINS.staleTime,
    });
  };

  return { ...query, prefetchAdjacentArea };
}

// Hook for fetching a single event by ID
export function useEvent(eventId: string | null) {
  return useQuery({
    queryKey: ['event', eventId],
    queryFn: () => apiClient.getEvent(eventId!),
    enabled: !!eventId,
    staleTime: CACHE_TIMES.EVENT_DETAIL.staleTime,
    gcTime: CACHE_TIMES.EVENT_DETAIL.gcTime,
  });
}

// Hook for fetching event registrations
export function useEventRegistrations(eventId: string | null) {
  return useQuery({
    queryKey: ['eventRegistrations', eventId],
    queryFn: () => apiClient.getEventRegistrations(eventId!),
    enabled: !!eventId,
    staleTime: CACHE_TIMES.REGISTRATIONS.staleTime,
    gcTime: CACHE_TIMES.REGISTRATIONS.gcTime,
    // Refetch on window focus for registrations (important for real-time updates)
    refetchOnWindowFocus: true,
  });
}

// Hook for registering to an event
export function useRegisterForEvent() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (eventId: string) => apiClient.registerForEvent(eventId),
    onSuccess: (data, eventId) => {
      // Invalidate relevant queries
      queryClient.invalidateQueries({ queryKey: ['event', eventId] });
      queryClient.invalidateQueries({ queryKey: ['eventRegistrations', eventId] });
      queryClient.invalidateQueries({ queryKey: ['events'] });
      queryClient.invalidateQueries({ queryKey: ['myRegistrations'] });
    },
  });
}

// Hook for cancelling registration
export function useCancelRegistration() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (eventId: string) => apiClient.cancelRegistration(eventId),
    onSuccess: (data, eventId) => {
      // Invalidate relevant queries
      queryClient.invalidateQueries({ queryKey: ['event', eventId] });
      queryClient.invalidateQueries({ queryKey: ['eventRegistrations', eventId] });
      queryClient.invalidateQueries({ queryKey: ['events'] });
      queryClient.invalidateQueries({ queryKey: ['myRegistrations'] });
    },
  });
}

// Hook for getting user's registrations
export function useMyRegistrations() {
  return useQuery({
    queryKey: ['myRegistrations'],
    queryFn: () => apiClient.getMyRegistrations(),
    staleTime: CACHE_TIMES.USER_DATA.staleTime,
    gcTime: CACHE_TIMES.USER_DATA.gcTime,
  });
}

// Hook for getting user's hosted events
export function useMyEvents() {
  return useQuery({
    queryKey: ['myEvents'],
    queryFn: () => apiClient.getMyEvents(),
    staleTime: CACHE_TIMES.USER_DATA.staleTime,
    gcTime: CACHE_TIMES.USER_DATA.gcTime,
  });
}

// Export cache times for use in providers
export { CACHE_TIMES };

// Helper to determine event pin color based on status and capacity
export function getEventPinColor(event: Event): 'green' | 'red' | 'gray' {
  if (event.status === 'completed' || event.status === 'cancelled') {
    return 'gray';
  }
  if (event.confirmed_count >= event.capacity) {
    return 'red';
  }
  return 'green';
}

// Helper to determine if an event is full
export function isEventFull(event: Event): boolean {
  return event.confirmed_count >= event.capacity;
}

// Helper to get available spots
export function getAvailableSpots(event: Event): number {
  return Math.max(0, event.capacity - event.confirmed_count);
}
