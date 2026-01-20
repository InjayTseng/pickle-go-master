import { redirect, notFound } from 'next/navigation';
import { Metadata } from 'next';

// Server-side data fetching to resolve short code to event ID
async function getEventByShortCode(code: string) {
  const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

  try {
    // First, try to find the event by short code
    const response = await fetch(`${apiUrl}/events/by-code/${code}`, {
      cache: 'no-store', // Always fetch fresh for redirects
    });

    if (!response.ok) {
      return null;
    }

    const data = await response.json();
    return data.success ? data.data : null;
  } catch (error) {
    console.error('Error fetching event by short code:', error);
    return null;
  }
}

// Generate metadata for social sharing (important for Line preview)
export async function generateMetadata({ params }: { params: { code: string } }): Promise<Metadata> {
  const event = await getEventByShortCode(params.code);

  if (!event) {
    return {
      title: 'Pickle Go - Find Pickleball Games',
    };
  }

  return {
    title: `${event.location?.name || 'Pickle Go'} 匹克球揪團`,
    description: `${event.skill_level_label} | ${
      event.capacity - event.confirmed_count > 0
        ? `還缺 ${event.capacity - event.confirmed_count} 人`
        : '已滿團'
    }`,
  };
}

interface PageProps {
  params: {
    code: string;
  };
}

export default async function ShortCodeRedirectPage({ params }: PageProps) {
  const event = await getEventByShortCode(params.code);

  if (!event) {
    notFound();
  }

  // Redirect to the full event page
  redirect(`/events/${event.id}`);
}
