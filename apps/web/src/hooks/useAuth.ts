'use client';

import { useAuthContext } from '@/contexts/AuthContext';

/**
 * Hook to access authentication state and methods
 */
export function useAuth() {
  return useAuthContext();
}
