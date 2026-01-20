import Cookies from 'js-cookie';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: {
    code: string;
    message: string;
    details?: Record<string, unknown>;
  };
}

export interface User {
  id: string;
  display_name: string;
  avatar_url?: string;
  email?: string;
}

export interface AuthResponse {
  user: User;
  access_token: string;
  refresh_token?: string;
}

export interface Event {
  id: string;
  host: User;
  title?: string;
  description?: string;
  event_date: string;
  start_time: string;
  end_time?: string;
  location: {
    name: string;
    address?: string;
    lat: number;
    lng: number;
    google_place_id?: string;
  };
  capacity: number;
  confirmed_count: number;
  waitlist_count: number;
  skill_level: string;
  skill_level_label: string;
  fee: number;
  status: string;
}

export interface EventListResponse {
  events: Event[];
  total: number;
  has_more: boolean;
}

export interface RegistrationResponse {
  id: string;
  event_id: string;
  status: string;
  waitlist_position?: number;
  message: string;
}

class ApiClient {
  private baseUrl: string;

  constructor(baseUrl: string) {
    this.baseUrl = baseUrl;
  }

  private getHeaders(): HeadersInit {
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
    };

    const token = Cookies.get('access_token');
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }

    return headers;
  }

  private async handleResponse<T>(response: Response): Promise<T> {
    const data: ApiResponse<T> = await response.json();

    if (!response.ok || !data.success) {
      // Handle token expiration
      if (response.status === 401) {
        // Try to refresh token
        const refreshed = await this.refreshToken();
        if (!refreshed) {
          // Clear tokens and redirect to login
          Cookies.remove('access_token');
          Cookies.remove('refresh_token');
          if (typeof window !== 'undefined') {
            window.location.href = '/login';
          }
        }
      }

      throw new ApiError(
        data.error?.code || 'UNKNOWN_ERROR',
        data.error?.message || 'An error occurred',
        response.status
      );
    }

    return data.data as T;
  }

  private async refreshToken(): Promise<boolean> {
    const refreshToken = Cookies.get('refresh_token');
    if (!refreshToken) {
      return false;
    }

    try {
      const response = await fetch(`${this.baseUrl}/auth/refresh`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refresh_token: refreshToken }),
      });

      if (!response.ok) {
        return false;
      }

      const data: ApiResponse<AuthResponse> = await response.json();
      if (data.success && data.data) {
        Cookies.set('access_token', data.data.access_token, { expires: 7 });
        return true;
      }

      return false;
    } catch {
      return false;
    }
  }

  async get<T>(path: string): Promise<T> {
    const response = await fetch(`${this.baseUrl}${path}`, {
      method: 'GET',
      headers: this.getHeaders(),
    });
    return this.handleResponse<T>(response);
  }

  async post<T>(path: string, body?: unknown): Promise<T> {
    const response = await fetch(`${this.baseUrl}${path}`, {
      method: 'POST',
      headers: this.getHeaders(),
      body: body ? JSON.stringify(body) : undefined,
    });
    return this.handleResponse<T>(response);
  }

  async put<T>(path: string, body?: unknown): Promise<T> {
    const response = await fetch(`${this.baseUrl}${path}`, {
      method: 'PUT',
      headers: this.getHeaders(),
      body: body ? JSON.stringify(body) : undefined,
    });
    return this.handleResponse<T>(response);
  }

  async delete<T>(path: string): Promise<T> {
    const response = await fetch(`${this.baseUrl}${path}`, {
      method: 'DELETE',
      headers: this.getHeaders(),
    });
    return this.handleResponse<T>(response);
  }

  // Auth endpoints
  async lineCallback(code: string, state?: string): Promise<AuthResponse> {
    return this.post<AuthResponse>('/auth/line/callback', { code, state });
  }

  async refreshAccessToken(refreshToken: string): Promise<AuthResponse> {
    return this.post<AuthResponse>('/auth/refresh', { refresh_token: refreshToken });
  }

  async logout(): Promise<{ message: string }> {
    return this.post<{ message: string }>('/auth/logout');
  }

  // User endpoints
  async getCurrentUser(): Promise<User> {
    return this.get<User>('/users/me');
  }

  async getMyEvents(): Promise<{ events: Event[]; total: number }> {
    return this.get<{ events: Event[]; total: number }>('/users/me/events');
  }

  async getMyRegistrations(): Promise<{ registrations: unknown[]; total: number }> {
    return this.get<{ registrations: unknown[]; total: number }>('/users/me/registrations');
  }

  // Event endpoints
  async listEvents(params?: {
    lat?: number;
    lng?: number;
    radius?: number;
    skill_level?: string;
    status?: string;
    limit?: number;
    offset?: number;
  }): Promise<EventListResponse> {
    const searchParams = new URLSearchParams();
    if (params) {
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined) {
          searchParams.append(key, String(value));
        }
      });
    }
    const query = searchParams.toString();
    return this.get<EventListResponse>(`/events${query ? `?${query}` : ''}`);
  }

  async getEvent(id: string): Promise<Event> {
    return this.get<Event>(`/events/${id}`);
  }

  async createEvent(data: {
    title?: string;
    description?: string;
    event_date: string;
    start_time: string;
    end_time?: string;
    location: {
      name: string;
      address?: string;
      lat: number;
      lng: number;
      google_place_id?: string;
    };
    capacity: number;
    skill_level: string;
    fee?: number;
  }): Promise<{ id: string; share_url: string }> {
    return this.post<{ id: string; share_url: string }>('/events', data);
  }

  async updateEvent(
    id: string,
    data: Partial<{
      title: string;
      description: string;
      event_date: string;
      start_time: string;
      end_time: string;
      capacity: number;
      skill_level: string;
      fee: number;
      status: string;
    }>
  ): Promise<{ id: string; message: string }> {
    return this.put<{ id: string; message: string }>(`/events/${id}`, data);
  }

  async deleteEvent(id: string): Promise<{ message: string }> {
    return this.delete<{ message: string }>(`/events/${id}`);
  }

  // Registration endpoints
  async registerForEvent(eventId: string): Promise<RegistrationResponse> {
    return this.post<RegistrationResponse>(`/events/${eventId}/register`);
  }

  async cancelRegistration(eventId: string): Promise<{ message: string }> {
    return this.delete<{ message: string }>(`/events/${eventId}/register`);
  }

  async getEventRegistrations(eventId: string): Promise<{
    confirmed: unknown[];
    waitlist: unknown[];
    confirmed_count: number;
    waitlist_count: number;
  }> {
    return this.get(`/events/${eventId}/registrations`);
  }
}

export class ApiError extends Error {
  code: string;
  statusCode: number;

  constructor(code: string, message: string, statusCode: number) {
    super(message);
    this.code = code;
    this.statusCode = statusCode;
    this.name = 'ApiError';
  }
}

export const apiClient = new ApiClient(API_BASE_URL);
