import { Event } from '@/lib/api-client';

interface EventJsonLdProps {
  event: Event;
  baseUrl: string;
}

/**
 * Schema.org Event structured data component
 * @see https://schema.org/Event
 * @see https://developers.google.com/search/docs/appearance/structured-data/event
 */
export function EventJsonLd({ event, baseUrl }: EventJsonLdProps) {
  // Build ISO 8601 datetime for start date
  const startDateTime = `${event.event_date}T${event.start_time}:00+08:00`;

  // Build ISO 8601 datetime for end date if available
  const endDateTime = event.end_time
    ? `${event.event_date}T${event.end_time}:00+08:00`
    : undefined;

  // Determine event status
  const eventStatus = (() => {
    switch (event.status) {
      case 'cancelled':
        return 'https://schema.org/EventCancelled';
      case 'completed':
        return 'https://schema.org/EventMovedOnline'; // No exact match, use past event indication
      default:
        return 'https://schema.org/EventScheduled';
    }
  })();

  // Determine event attendance mode (always offline for pickleball)
  const eventAttendanceMode = 'https://schema.org/OfflineEventAttendanceMode';

  // Calculate available spots
  const spotsRemaining = event.capacity - event.confirmed_count;

  // Build structured data
  const structuredData = {
    '@context': 'https://schema.org',
    '@type': 'SportsEvent',
    name: event.title || `${event.location.name} 匹克球揪團`,
    description: event.description || `${event.skill_level_label} 匹克球活動，還缺 ${spotsRemaining > 0 ? spotsRemaining : 0} 人`,
    startDate: startDateTime,
    ...(endDateTime && { endDate: endDateTime }),
    eventStatus,
    eventAttendanceMode,
    location: {
      '@type': 'Place',
      name: event.location.name,
      address: {
        '@type': 'PostalAddress',
        streetAddress: event.location.address || event.location.name,
        addressLocality: '台北市',
        addressRegion: '台灣',
        addressCountry: 'TW',
      },
      geo: {
        '@type': 'GeoCoordinates',
        latitude: event.location.lat,
        longitude: event.location.lng,
      },
    },
    organizer: {
      '@type': 'Person',
      name: event.host.display_name,
      ...(event.host.avatar_url && { image: event.host.avatar_url }),
    },
    offers: {
      '@type': 'Offer',
      price: event.fee,
      priceCurrency: 'TWD',
      availability: spotsRemaining > 0
        ? 'https://schema.org/InStock'
        : 'https://schema.org/SoldOut',
      url: `${baseUrl}/events/${event.id}`,
      validFrom: event.created_at || new Date().toISOString(),
    },
    maximumAttendeeCapacity: event.capacity,
    remainingAttendeeCapacity: spotsRemaining > 0 ? spotsRemaining : 0,
    sport: 'Pickleball',
    url: `${baseUrl}/events/${event.id}`,
    image: `${baseUrl}/api/og?title=${encodeURIComponent(event.location.name)}&date=${encodeURIComponent(event.event_date)}`,
  };

  return (
    <script
      type="application/ld+json"
      dangerouslySetInnerHTML={{ __html: JSON.stringify(structuredData) }}
    />
  );
}
