'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { format, parseISO, isPast, isToday } from 'date-fns';
import { zhTW } from 'date-fns/locale';
import { useAuthContext } from '@/contexts/AuthContext';
import { useMyEvents } from '@/hooks/useEvents';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Spinner } from '@/components/ui/spinner';
import { Calendar, MapPin, Users, Plus, ChevronRight, CalendarX2 } from 'lucide-react';

export default function MyEventsPage() {
  const router = useRouter();
  const { user, isAuthenticated, isLoading: authLoading } = useAuthContext();
  const { data, isLoading: eventsLoading, error } = useMyEvents();

  // Redirect to login if not authenticated
  useEffect(() => {
    if (!authLoading && !isAuthenticated) {
      router.push('/login?returnUrl=/my/events');
    }
  }, [authLoading, isAuthenticated, router]);

  if (authLoading || eventsLoading) {
    return (
      <main className="container max-w-2xl mx-auto px-4 py-8">
        <div className="flex items-center justify-center h-64">
          <Spinner className="h-8 w-8" />
        </div>
      </main>
    );
  }

  if (!isAuthenticated) {
    return null; // Will redirect
  }

  const events = data?.events || [];

  // Separate upcoming and past events
  const upcoming = events.filter((e: any) => {
    const eventDate = parseISO(e.event_date);
    return (!isPast(eventDate) || isToday(eventDate)) && e.status !== 'cancelled';
  });
  const past = events.filter((e: any) => {
    const eventDate = parseISO(e.event_date);
    return (isPast(eventDate) && !isToday(eventDate)) || e.status === 'cancelled';
  });

  // Get status badge
  const getStatusBadge = (event: any) => {
    if (event.status === 'cancelled') {
      return <Badge variant="destructive">Cancelled</Badge>;
    }
    if (event.status === 'completed') {
      return <Badge variant="secondary">Completed</Badge>;
    }
    if (event.confirmed_count >= event.capacity) {
      return <Badge className="bg-red-500">Full</Badge>;
    }
    return (
      <Badge className="bg-green-500">
        {event.capacity - event.confirmed_count} spots left
      </Badge>
    );
  };

  const EventCard = ({ event }: { event: any }) => {
    const eventDate = parseISO(event.event_date);
    const formattedDate = format(eventDate, 'M/d (EEE)', { locale: zhTW });
    const isUpcoming = (!isPast(eventDate) || isToday(eventDate)) && event.status !== 'cancelled';

    return (
      <Link href={`/events/${event.id}`}>
        <Card className={`hover:shadow-md transition-shadow ${!isUpcoming ? 'opacity-75' : ''}`}>
          <CardContent className="p-4">
            <div className="flex items-start justify-between mb-2">
              {getStatusBadge(event)}
              <ChevronRight className="h-5 w-5 text-muted-foreground" />
            </div>

            <h3 className="font-semibold text-base mb-2">
              {event.title || event.location?.name}
            </h3>

            <div className="space-y-1 text-sm text-muted-foreground">
              <div className="flex items-center gap-2">
                <Calendar className="h-4 w-4" />
                <span className={isToday(eventDate) ? 'text-primary font-medium' : ''}>
                  {isToday(eventDate) ? 'Today' : formattedDate} {event.start_time}
                </span>
              </div>
              <div className="flex items-center gap-2">
                <MapPin className="h-4 w-4" />
                <span className="line-clamp-1">{event.location?.name}</span>
              </div>
              <div className="flex items-center gap-2">
                <Users className="h-4 w-4" />
                <span>
                  {event.confirmed_count || 0}/{event.capacity} registered
                </span>
              </div>
            </div>
          </CardContent>
        </Card>
      </Link>
    );
  };

  return (
    <main className="container max-w-2xl mx-auto px-4 py-8">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">My Events</h1>
        <Link href="/events/new">
          <Button size="sm" className="gap-2">
            <Plus className="h-4 w-4" />
            Create Event
          </Button>
        </Link>
      </div>

      {events.length === 0 ? (
        <div className="text-center py-12">
          <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-gray-100">
            <CalendarX2 className="h-8 w-8 text-gray-400" />
          </div>
          <h2 className="text-lg font-semibold mb-2">No events yet</h2>
          <p className="text-muted-foreground mb-6">
            Create your first event and invite friends to play!
          </p>
          <Link href="/events/new">
            <Button>Create Event</Button>
          </Link>
        </div>
      ) : (
        <div className="space-y-6">
          {/* Upcoming Events */}
          {upcoming.length > 0 && (
            <section>
              <h2 className="text-lg font-semibold mb-3 flex items-center gap-2">
                <Calendar className="h-5 w-5 text-primary" />
                Upcoming ({upcoming.length})
              </h2>
              <div className="space-y-3">
                {upcoming.map((event: any) => (
                  <EventCard key={event.id} event={event} />
                ))}
              </div>
            </section>
          )}

          {/* Past Events */}
          {past.length > 0 && (
            <section>
              <h2 className="text-lg font-semibold mb-3 flex items-center gap-2 text-muted-foreground">
                Past & Cancelled ({past.length})
              </h2>
              <div className="space-y-3">
                {past.map((event: any) => (
                  <EventCard key={event.id} event={event} />
                ))}
              </div>
            </section>
          )}
        </div>
      )}
    </main>
  );
}
