'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
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

// Hook for fetching events list with geo filter
export function useEvents(params: EventQueryParams = {}, enabled = true) {
  return useQuery({
    queryKey: ['events', params],
    queryFn: () => apiClient.listEvents(params),
    enabled,
    staleTime: 30000, // Consider data stale after 30 seconds
    refetchOnWindowFocus: false,
  });
}

// Hook for fetching a single event by ID
export function useEvent(eventId: string | null) {
  return useQuery({
    queryKey: ['event', eventId],
    queryFn: () => apiClient.getEvent(eventId!),
    enabled: !!eventId,
    staleTime: 30000,
  });
}

// Hook for fetching event registrations
export function useEventRegistrations(eventId: string | null) {
  return useQuery({
    queryKey: ['eventRegistrations', eventId],
    queryFn: () => apiClient.getEventRegistrations(eventId!),
    enabled: !!eventId,
    staleTime: 10000,
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
    staleTime: 30000,
  });
}

// Hook for getting user's hosted events
export function useMyEvents() {
  return useQuery({
    queryKey: ['myEvents'],
    queryFn: () => apiClient.getMyEvents(),
    staleTime: 30000,
  });
}

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
