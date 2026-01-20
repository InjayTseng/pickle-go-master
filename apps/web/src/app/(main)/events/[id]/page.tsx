import { Metadata } from 'next';
import { notFound } from 'next/navigation';
import { format } from 'date-fns';
import { zhTW } from 'date-fns/locale';

import { EventDetail } from '@/components/event/EventDetail';
import { EventJsonLd } from '@/components/seo/EventJsonLd';

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

// Dynamic metadata for OG tags and SEO
export async function generateMetadata({ params }: { params: { id: string } }): Promise<Metadata> {
  const event = await getEvent(params.id);

  if (!event) {
    return {
      title: '活動不存在',
      description: '找不到此活動，可能已被刪除或從未存在。',
      robots: {
        index: false,
        follow: false,
      },
    };
  }

  // Parse date for display
  const eventDate = new Date(event.event_date);
  const dayOfWeek = format(eventDate, 'EEEE', { locale: zhTW });
  const formattedDate = format(eventDate, 'MM/dd', { locale: zhTW });
  const formattedFullDate = format(eventDate, 'yyyy/MM/dd', { locale: zhTW });

  // Calculate spots remaining
  const spotsRemaining = event.capacity - event.confirmed_count;
  const spotsText = spotsRemaining > 0 ? `還缺 ${spotsRemaining} 人` : '已滿團';

  // Build title: "01/25 (六) 20:00 @ 內湖運動中心"
  const ogTitle = `${formattedDate} (${dayOfWeek.charAt(0)}) ${event.start_time} @ ${event.location.name}`;

  // Build description for SEO: more detailed
  const seoDescription = `${event.location.name}匹克球揪團 - ${formattedFullDate} ${event.start_time}。程度：${event.skill_level_label}，${spotsText}。${event.fee > 0 ? `費用 NT$${event.fee}` : '免費參加'}。`;

  // Short description for social sharing
  const socialDescription = `${event.skill_level_label} | ${spotsText}`;

  // Full title for browser tab (SEO optimized)
  const seoTitle = `${event.location.name} 匹克球揪團 ${formattedDate} - ${event.skill_level_label}`;

  // Base URL for images
  const baseUrl = process.env.NEXT_PUBLIC_BASE_URL || 'https://picklego.tw';
  const eventUrl = `${baseUrl}/events/${params.id}`;

  // Dynamic OG image URL
  const ogImageParams = new URLSearchParams({
    title: ogTitle,
    location: event.location.name,
    date: `${formattedDate} (${dayOfWeek.charAt(0)})`,
    spots: spotsText,
  });
  const ogImageUrl = `${baseUrl}/api/og?${ogImageParams.toString()}`;

  // Determine if event should be indexed (past/cancelled events have lower priority)
  const shouldIndex = event.status !== 'cancelled' && event.status !== 'completed';

  return {
    title: seoTitle,
    description: seoDescription,
    keywords: [
      '匹克球',
      event.location.name,
      '匹克球揪團',
      event.skill_level_label,
      '打球',
      '球局',
    ],
    alternates: {
      canonical: eventUrl,
    },
    robots: {
      index: shouldIndex,
      follow: true,
      googleBot: {
        index: shouldIndex,
        follow: true,
        'max-image-preview': 'large',
      },
    },
    openGraph: {
      type: 'website',
      locale: 'zh_TW',
      url: eventUrl,
      siteName: 'Pickle Go',
      title: ogTitle,
      description: socialDescription,
      images: [
        {
          url: ogImageUrl,
          width: 1200,
          height: 630,
          alt: ogTitle,
        },
      ],
    },
    twitter: {
      card: 'summary_large_image',
      title: ogTitle,
      description: socialDescription,
      images: [ogImageUrl],
    },
    // Additional meta for Line and other platforms
    other: {
      'og:site_name': 'Pickle Go',
      'og:locale': 'zh_TW',
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

  const baseUrl = process.env.NEXT_PUBLIC_BASE_URL || 'https://picklego.tw';

  return (
    <>
      <EventJsonLd event={event} baseUrl={baseUrl} />
      <EventDetail event={event} />
    </>
  );
}
