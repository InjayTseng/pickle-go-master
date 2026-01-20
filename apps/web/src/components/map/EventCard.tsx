'use client';

import React from 'react';
import Link from 'next/link';
import { format, parseISO } from 'date-fns';
import { zhTW } from 'date-fns/locale';
import { Event } from '@/lib/api-client';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { getEventPinColor, isEventFull, getAvailableSpots } from '@/hooks/useEvents';
import { MapPin, Calendar, Clock, Users, DollarSign } from 'lucide-react';

interface EventCardProps {
  event: Event;
  onClose?: () => void;
  compact?: boolean;
}

export function EventCard({ event, onClose, compact = false }: EventCardProps) {
  const color = getEventPinColor(event);
  const isFull = isEventFull(event);
  const availableSpots = getAvailableSpots(event);

  // Format date
  const eventDate = parseISO(event.event_date);
  const formattedDate = format(eventDate, 'M/d (EEE)', { locale: zhTW });

  // Get status badge
  const getStatusBadge = () => {
    if (event.status === 'cancelled') {
      return <Badge variant="destructive">已取消</Badge>;
    }
    if (event.status === 'completed') {
      return <Badge variant="secondary">已結束</Badge>;
    }
    if (isFull) {
      return <Badge variant="destructive">已滿</Badge>;
    }
    return (
      <Badge variant="default" className="bg-green-500 hover:bg-green-600">
        還缺 {availableSpots} 人
      </Badge>
    );
  };

  // Get skill level badge color
  const getSkillBadgeVariant = () => {
    switch (event.skill_level) {
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

  if (compact) {
    return (
      <Card className="w-72 shadow-xl border-0 bg-white/95 backdrop-blur-sm">
        <CardContent className="p-4">
          <div className="flex justify-between items-start mb-2">
            <div className="flex items-center gap-2">
              <Avatar className="h-6 w-6">
                <AvatarImage src={event.host.avatar_url} />
                <AvatarFallback>{event.host.display_name?.[0] || 'U'}</AvatarFallback>
              </Avatar>
              <span className="text-sm text-muted-foreground">{event.host.display_name}</span>
            </div>
            {getStatusBadge()}
          </div>

          <h3 className="font-semibold text-base mb-2 line-clamp-1">
            {event.title || event.location.name}
          </h3>

          <div className="space-y-1 text-sm text-muted-foreground mb-3">
            <div className="flex items-center gap-2">
              <Calendar className="h-4 w-4" />
              <span>{formattedDate} {event.start_time}</span>
            </div>
            <div className="flex items-center gap-2">
              <MapPin className="h-4 w-4" />
              <span className="line-clamp-1">{event.location.name}</span>
            </div>
            <div className="flex items-center gap-2">
              <Users className="h-4 w-4" />
              <span>{event.confirmed_count}/{event.capacity} 人</span>
              {event.fee > 0 && (
                <>
                  <span className="text-muted-foreground">|</span>
                  <DollarSign className="h-4 w-4" />
                  <span>${event.fee}</span>
                </>
              )}
            </div>
          </div>

          <div className="flex items-center justify-between">
            <Badge className={`${getSkillBadgeVariant()} border-0`}>
              {event.skill_level_label}
            </Badge>
            <Link href={`/events/${event.id}`}>
              <Button size="sm" variant="default">
                查看詳情
              </Button>
            </Link>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="w-full shadow-lg">
      <CardContent className="p-5">
        {/* Header */}
        <div className="flex justify-between items-start mb-3">
          <div className="flex items-center gap-3">
            <Avatar className="h-10 w-10">
              <AvatarImage src={event.host.avatar_url} />
              <AvatarFallback>{event.host.display_name?.[0] || 'U'}</AvatarFallback>
            </Avatar>
            <div>
              <p className="font-medium">{event.host.display_name}</p>
              <p className="text-sm text-muted-foreground">主辦人</p>
            </div>
          </div>
          {getStatusBadge()}
        </div>

        {/* Title */}
        {event.title && (
          <h3 className="font-semibold text-lg mb-3">{event.title}</h3>
        )}

        {/* Details */}
        <div className="space-y-2 text-sm mb-4">
          <div className="flex items-center gap-3">
            <Calendar className="h-5 w-5 text-muted-foreground" />
            <span className="font-medium">{formattedDate}</span>
            <Clock className="h-5 w-5 text-muted-foreground ml-2" />
            <span>{event.start_time}{event.end_time ? ` - ${event.end_time}` : ''}</span>
          </div>
          <div className="flex items-center gap-3">
            <MapPin className="h-5 w-5 text-muted-foreground" />
            <span>{event.location.name}</span>
          </div>
          {event.location.address && (
            <p className="text-muted-foreground ml-8">{event.location.address}</p>
          )}
          <div className="flex items-center gap-3">
            <Users className="h-5 w-5 text-muted-foreground" />
            <span>
              {event.confirmed_count}/{event.capacity} 人報名
              {event.waitlist_count > 0 && (
                <span className="text-muted-foreground"> ({event.waitlist_count} 人候補)</span>
              )}
            </span>
          </div>
          {event.fee > 0 && (
            <div className="flex items-center gap-3">
              <DollarSign className="h-5 w-5 text-muted-foreground" />
              <span>${event.fee}</span>
            </div>
          )}
        </div>

        {/* Skill level */}
        <div className="mb-4">
          <Badge className={`${getSkillBadgeVariant()} border-0`}>
            {event.skill_level_label}
          </Badge>
        </div>

        {/* Description */}
        {event.description && (
          <p className="text-sm text-muted-foreground mb-4">{event.description}</p>
        )}

        {/* Actions */}
        <div className="flex gap-2">
          <Link href={`/events/${event.id}`} className="flex-1">
            <Button className="w-full" variant="default">
              查看詳情
            </Button>
          </Link>
          {onClose && (
            <Button variant="outline" onClick={onClose}>
              關閉
            </Button>
          )}
        </div>
      </CardContent>
    </Card>
  );
}

export default EventCard;
