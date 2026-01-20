'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { format, parseISO, isPast, isToday } from 'date-fns';
import { zhTW } from 'date-fns/locale';
import { useAuthContext } from '@/contexts/AuthContext';
import { useMyRegistrations } from '@/hooks/useEvents';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Spinner } from '@/components/ui/spinner';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Calendar, MapPin, Clock, Users, ChevronRight, CalendarX2 } from 'lucide-react';

interface RegistrationWithEvent {
  id: string;
  event_id: string;
  status: string;
  waitlist_position?: number;
  registered_at: string;
  event: {
    id: string;
    title?: string;
    event_date: string;
    start_time: string;
    location: string;
    skill_level: string;
    status: string;
  };
}

export default function MyRegistrationsPage() {
  const router = useRouter();
  const { user, isAuthenticated, isLoading: authLoading } = useAuthContext();
  const { data, isLoading: registrationsLoading, error } = useMyRegistrations();

  // Redirect to login if not authenticated
  useEffect(() => {
    if (!authLoading && !isAuthenticated) {
      router.push('/login?returnUrl=/my/registrations');
    }
  }, [authLoading, isAuthenticated, router]);

  if (authLoading || registrationsLoading) {
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

  const registrations = (data?.registrations || []) as RegistrationWithEvent[];

  // Separate upcoming and past registrations
  const now = new Date();
  const upcoming = registrations.filter((r) => {
    const eventDate = parseISO(r.event.event_date);
    return !isPast(eventDate) || isToday(eventDate);
  });
  const past = registrations.filter((r) => {
    const eventDate = parseISO(r.event.event_date);
    return isPast(eventDate) && !isToday(eventDate);
  });

  // Get status badge
  const getStatusBadge = (status: string, waitlistPosition?: number) => {
    switch (status) {
      case 'confirmed':
        return <Badge className="bg-green-500">Confirmed</Badge>;
      case 'waitlist':
        return (
          <Badge variant="secondary">
            Waitlist #{waitlistPosition}
          </Badge>
        );
      case 'cancelled':
        return <Badge variant="destructive">Cancelled</Badge>;
      default:
        return null;
    }
  };

  // Get skill level badge color
  const getSkillBadgeClass = (skillLevel: string) => {
    switch (skillLevel) {
      case 'beginner':
        return 'bg-emerald-100 text-emerald-800';
      case 'intermediate':
        return 'bg-blue-100 text-blue-800';
      case 'advanced':
        return 'bg-purple-100 text-purple-800';
      case 'expert':
        return 'bg-red-100 text-red-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  const RegistrationCard = ({ registration }: { registration: RegistrationWithEvent }) => {
    const eventDate = parseISO(registration.event.event_date);
    const formattedDate = format(eventDate, 'M/d (EEE)', { locale: zhTW });
    const isUpcoming = !isPast(eventDate) || isToday(eventDate);

    return (
      <Link href={`/events/${registration.event_id}`}>
        <Card className={`hover:shadow-md transition-shadow ${!isUpcoming ? 'opacity-75' : ''}`}>
          <CardContent className="p-4">
            <div className="flex items-start justify-between mb-2">
              <div className="flex items-center gap-2">
                {getStatusBadge(registration.status, registration.waitlist_position)}
                {registration.event.status === 'cancelled' && (
                  <Badge variant="destructive">Event Cancelled</Badge>
                )}
              </div>
              <ChevronRight className="h-5 w-5 text-muted-foreground" />
            </div>

            <h3 className="font-semibold text-base mb-2">
              {registration.event.title || registration.event.location}
            </h3>

            <div className="space-y-1 text-sm text-muted-foreground">
              <div className="flex items-center gap-2">
                <Calendar className="h-4 w-4" />
                <span className={isToday(eventDate) ? 'text-primary font-medium' : ''}>
                  {isToday(eventDate) ? 'Today' : formattedDate} {registration.event.start_time}
                </span>
              </div>
              <div className="flex items-center gap-2">
                <MapPin className="h-4 w-4" />
                <span className="line-clamp-1">{registration.event.location}</span>
              </div>
            </div>

            <div className="mt-3 flex items-center gap-2">
              <Badge className={`${getSkillBadgeClass(registration.event.skill_level)} border-0 text-xs`}>
                {registration.event.skill_level}
              </Badge>
            </div>
          </CardContent>
        </Card>
      </Link>
    );
  };

  return (
    <main className="container max-w-2xl mx-auto px-4 py-8">
      <h1 className="text-2xl font-bold mb-6">My Registrations</h1>

      {registrations.length === 0 ? (
        <div className="text-center py-12">
          <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-gray-100">
            <CalendarX2 className="h-8 w-8 text-gray-400" />
          </div>
          <h2 className="text-lg font-semibold mb-2">No registrations yet</h2>
          <p className="text-muted-foreground mb-6">
            Find events near you and start playing!
          </p>
          <Link href="/">
            <Button>Browse Events</Button>
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
                {upcoming.map((registration) => (
                  <RegistrationCard key={registration.id} registration={registration} />
                ))}
              </div>
            </section>
          )}

          {/* Past Events */}
          {past.length > 0 && (
            <section>
              <h2 className="text-lg font-semibold mb-3 flex items-center gap-2 text-muted-foreground">
                <Clock className="h-5 w-5" />
                Past ({past.length})
              </h2>
              <div className="space-y-3">
                {past.map((registration) => (
                  <RegistrationCard key={registration.id} registration={registration} />
                ))}
              </div>
            </section>
          )}
        </div>
      )}
    </main>
  );
}
