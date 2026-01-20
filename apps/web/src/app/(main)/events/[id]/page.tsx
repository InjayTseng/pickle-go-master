import { Metadata } from 'next';
import { notFound } from 'next/navigation';
import { format } from 'date-fns';
import { zhTW } from 'date-fns/locale';

import { EventDetail } from '@/components/event/EventDetail';

// Server-side data fetching
async function getEvent(id: string) {
  const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

  try {
    const response = await fetch(`${apiUrl}/events/${id}`, {
      next: { revalidate: 60 }, // Revalidate every 60 seconds
    });

    if (!response.ok) {
      if (response.status === 404) {
        return null;
      }
      throw new Error('Failed to fetch event');
    }

    const data = await response.json();
    return data.success ? data.data : null;
  } catch (error) {
    console.error('Error fetching event:', error);
    return null;
  }
}

// Dynamic metadata for OG tags
export async function generateMetadata({ params }: { params: { id: string } }): Promise<Metadata> {
  const event = await getEvent(params.id);

  if (!event) {
    return {
      title: '活動不存在',
    };
  }

  // Parse date for display
  const eventDate = new Date(event.event_date);
  const dayOfWeek = format(eventDate, 'EEEE', { locale: zhTW });
  const formattedDate = format(eventDate, 'MM/dd', { locale: zhTW });

  // Calculate spots remaining
  const spotsRemaining = event.capacity - event.confirmed_count;
  const spotsText = spotsRemaining > 0 ? `還缺 ${spotsRemaining} 人` : '已滿團';

  // Build title: "01/25 (六) 20:00 @ 內湖運動中心"
  const title = `${formattedDate} (${dayOfWeek.charAt(0)}) ${event.start_time} @ ${event.location.name}`;

  // Build description: "新手友善 | 還缺 3 人"
  const description = `${event.skill_level_label} | ${spotsText}`;

  // Full title for tab
  const fullTitle = `${event.location.name} 匹克球揪團 ${formattedDate}`;

  // Base URL for images
  const baseUrl = process.env.NEXT_PUBLIC_BASE_URL || 'https://picklego.tw';

  // Dynamic OG image URL
  const ogImageParams = new URLSearchParams({
    title: title,
    location: event.location.name,
    date: `${formattedDate} (${dayOfWeek.charAt(0)})`,
    spots: spotsText,
  });
  const ogImageUrl = `${baseUrl}/api/og?${ogImageParams.toString()}`;

  return {
    title: fullTitle,
    description: description,
    openGraph: {
      type: 'website',
      locale: 'zh_TW',
      url: `${baseUrl}/events/${params.id}`,
      siteName: 'Pickle Go',
      title: title,
      description: description,
      images: [
        {
          url: ogImageUrl,
          width: 1200,
          height: 630,
          alt: title,
        },
      ],
    },
    twitter: {
      card: 'summary_large_image',
      title: title,
      description: description,
      images: [ogImageUrl],
    },
    // Additional meta for Line
    other: {
      'og:site_name': 'Pickle Go',
    },
  };
}

interface PageProps {
  params: {
    id: string;
  };
}

export default async function EventPage({ params }: PageProps) {
  const event = await getEvent(params.id);

  if (!event) {
    notFound();
  }

  return <EventDetail event={event} />;
}
