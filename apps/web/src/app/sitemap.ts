import { MetadataRoute } from 'next';

const BASE_URL = process.env.NEXT_PUBLIC_BASE_URL || 'https://picklego.tw';
const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

interface Event {
  id: string;
  event_date: string;
  status: string;
  updated_at?: string;
}

interface EventListResponse {
  events: Event[];
  total: number;
  has_more: boolean;
}

/**
 * Fetch all public events from the API
 */
async function getPublicEvents(): Promise<Event[]> {
  try {
    const response = await fetch(`${API_URL}/events?status=open&limit=1000`, {
      next: { revalidate: 3600 }, // Revalidate every hour
    });

    if (!response.ok) {
      console.error('Failed to fetch events for sitemap:', response.statusText);
      return [];
    }

    const data: { success: boolean; data?: EventListResponse } = await response.json();
    return data.success && data.data ? data.data.events : [];
  } catch (error) {
    console.error('Error fetching events for sitemap:', error);
    return [];
  }
}

/**
 * Dynamic sitemap generation for SEO
 * @see https://nextjs.org/docs/app/api-reference/file-conventions/metadata/sitemap
 */
export default async function sitemap(): Promise<MetadataRoute.Sitemap> {
  // Static pages
  const staticPages: MetadataRoute.Sitemap = [
    {
      url: BASE_URL,
      lastModified: new Date(),
      changeFrequency: 'hourly',
      priority: 1,
    },
    {
      url: `${BASE_URL}/login`,
      lastModified: new Date(),
      changeFrequency: 'monthly',
      priority: 0.3,
    },
  ];

  // Fetch all public events
  const events = await getPublicEvents();

  // Generate event pages
  const eventPages: MetadataRoute.Sitemap = events.map((event) => ({
    url: `${BASE_URL}/events/${event.id}`,
    lastModified: event.updated_at ? new Date(event.updated_at) : new Date(),
    changeFrequency: 'daily' as const,
    priority: 0.8,
  }));

  return [...staticPages, ...eventPages];
}
