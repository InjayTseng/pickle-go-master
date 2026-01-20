// Shared types between frontend and backend
// These types are used for API contracts

export type SkillLevel = 'beginner' | 'intermediate' | 'advanced' | 'expert' | 'any';

export type EventStatus = 'open' | 'full' | 'cancelled' | 'completed';

export type RegistrationStatus = 'confirmed' | 'waitlist' | 'cancelled';

export interface Coordinates {
  lat: number;
  lng: number;
}

export interface Location extends Coordinates {
  name: string;
  address?: string;
  googlePlaceId?: string;
}

// API Error codes
export const ErrorCodes = {
  // Authentication
  UNAUTHORIZED: 'UNAUTHORIZED',
  INVALID_TOKEN: 'INVALID_TOKEN',
  TOKEN_EXPIRED: 'TOKEN_EXPIRED',

  // Event
  EVENT_NOT_FOUND: 'EVENT_NOT_FOUND',
  EVENT_FULL: 'EVENT_FULL',
  EVENT_CANCELLED: 'EVENT_CANCELLED',
  EVENT_EXPIRED: 'EVENT_EXPIRED',

  // Registration
  ALREADY_REGISTERED: 'ALREADY_REGISTERED',
  NOT_REGISTERED: 'NOT_REGISTERED',
  CANNOT_CANCEL: 'CANNOT_CANCEL',

  // General
  VALIDATION_ERROR: 'VALIDATION_ERROR',
  INTERNAL_ERROR: 'INTERNAL_ERROR',
} as const;

export type ErrorCode = typeof ErrorCodes[keyof typeof ErrorCodes];
