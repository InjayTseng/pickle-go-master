'use client';

import React, { createContext, useContext, useEffect, useState, useCallback } from 'react';
import Cookies from 'js-cookie';
import { apiClient, User, AuthResponse } from '@/lib/api-client';

interface AuthContextType {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (code: string, state?: string) => Promise<void>;
  logout: () => Promise<void>;
  refreshUser: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

// Secure Cookie options helper
const getCookieOptions = (days: number): Cookies.CookieAttributes => {
  const isProduction = process.env.NODE_ENV === 'production';
  return {
    expires: days,
    secure: isProduction, // Only transmit over HTTPS in production
    sameSite: 'strict',   // Prevent CSRF attacks
    path: '/',
  };
};

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  // Initialize auth state from stored token
  useEffect(() => {
    const initAuth = async () => {
      const token = Cookies.get('access_token');
      if (!token) {
        setIsLoading(false);
        return;
      }

      try {
        const userData = await apiClient.getCurrentUser();
        setUser(userData);
      } catch (error) {
        // Token might be invalid, try to refresh
        const refreshToken = Cookies.get('refresh_token');
        if (refreshToken) {
          try {
            const response = await apiClient.refreshAccessToken(refreshToken);
            Cookies.set('access_token', response.access_token, getCookieOptions(7));
            setUser(response.user);
          } catch {
            // Refresh failed, clear tokens
            Cookies.remove('access_token');
            Cookies.remove('refresh_token');
          }
        } else {
          Cookies.remove('access_token');
        }
      } finally {
        setIsLoading(false);
      }
    };

    initAuth();
  }, []);

  const login = useCallback(async (code: string, state?: string) => {
    setIsLoading(true);
    try {
      const response: AuthResponse = await apiClient.lineCallback(code, state);

      // Store tokens with secure options
      Cookies.set('access_token', response.access_token, getCookieOptions(7));
      if (response.refresh_token) {
        Cookies.set('refresh_token', response.refresh_token, getCookieOptions(30));
      }

      setUser(response.user);
    } finally {
      setIsLoading(false);
    }
  }, []);

  const logout = useCallback(async () => {
    try {
      await apiClient.logout();
    } catch {
      // Ignore logout API errors
    } finally {
      Cookies.remove('access_token');
      Cookies.remove('refresh_token');
      setUser(null);
    }
  }, []);

  const refreshUser = useCallback(async () => {
    const token = Cookies.get('access_token');
    if (!token) return;

    try {
      const userData = await apiClient.getCurrentUser();
      setUser(userData);
    } catch {
      // Ignore refresh errors
    }
  }, []);

  return (
    <AuthContext.Provider
      value={{
        user,
        isAuthenticated: !!user,
        isLoading,
        login,
        logout,
        refreshUser,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuthContext() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuthContext must be used within an AuthProvider');
  }
  return context;
}
