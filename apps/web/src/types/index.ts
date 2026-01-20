// User types
export interface User {
  id: string;
  lineUserId: string;
  displayName: string;
  avatarUrl?: string;
  email?: string;
  createdAt: string;
  updatedAt: string;
}

// Event types
export type SkillLevel = 'beginner' | 'intermediate' | 'advanced' | 'expert' | 'any';

export interface Location {
  name: string;
  address?: string;
  lat: number;
  lng: number;
  googlePlaceId?: string;
}

export interface Event {
  id: string;
  host: {
    id: string;
    displayName: string;
    avatarUrl?: string;
  };
  title?: string;
  description?: string;
  eventDate: string;
  startTime: string;
  endTime?: string;
  location: Location;
  capacity: number;
  confirmedCount: number;
  waitlistCount: number;
  skillLevel: SkillLevel;
  fee: number;
  status: 'open' | 'full' | 'cancelled' | 'completed';
  createdAt: string;
  updatedAt: string;
}

// Registration types
export type RegistrationStatus = 'confirmed' | 'waitlist' | 'cancelled';

export interface Registration {
  id: string;
  eventId: string;
  userId: string;
  status: RegistrationStatus;
  waitlistPosition?: number;
  registeredAt: string;
  confirmedAt?: string;
  cancelledAt?: string;
}

// API response types
export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: {
    code: string;
    message: string;
    details?: Record<string, unknown>;
  };
}

export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  hasMore: boolean;
}

// Skill level labels
export const skillLevelLabels: Record<SkillLevel, string> = {
  beginner: '新手友善 (2.0-2.5)',
  intermediate: '中階 (2.5-3.5)',
  advanced: '進階 (3.5-4.5)',
  expert: '高階 (4.5+)',
  any: '不限程度',
};
