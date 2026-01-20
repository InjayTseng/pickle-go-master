'use client';

import { useState, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import { format } from 'date-fns';
import { zhTW } from 'date-fns/locale';
import {
  Calendar,
  Clock,
  MapPin,
  Users,
  DollarSign,
  Target,
  Share2,
  Copy,
  Check,
  ArrowLeft,
  ExternalLink,
} from 'lucide-react';

import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { useAuthContext } from '@/contexts/AuthContext';
import { apiClient, Event, RegistrationResponse } from '@/lib/api-client';
import { getLineLoginURL } from '@/lib/auth';

interface EventDetailProps {
  event: Event;
}

export function EventDetail({ event }: EventDetailProps) {
  const router = useRouter();
  const { user, isAuthenticated, isLoading: authLoading } = useAuthContext();
  const [isRegistering, setIsRegistering] = useState(false);
  const [registrationResult, setRegistrationResult] = useState<RegistrationResponse | null>(null);
  const [copied, setCopied] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Parse event date
  const eventDate = new Date(event.event_date);
  const dayOfWeek = format(eventDate, 'EEEE', { locale: zhTW });
  const formattedDate = format(eventDate, 'yyyy/MM/dd', { locale: zhTW });

  // Calculate spots
  const spotsRemaining = event.capacity - event.confirmed_count;
  const isFull = spotsRemaining <= 0;
  const isHost = user?.id === event.host.id;
  const isCancelled = event.status === 'cancelled';

  // Determine event status badge
  const getStatusBadge = () => {
    if (isCancelled) {
      return <Badge variant="destructive">已取消</Badge>;
    }
    if (isFull) {
      return <Badge variant="warning">已滿團</Badge>;
    }
    return <Badge variant="success">招募中</Badge>;
  };

  // Handle registration
  const handleRegister = useCallback(async () => {
    if (!isAuthenticated) {
      // Save intended destination and redirect to login
      sessionStorage.setItem('redirectAfterLogin', `/events/${event.id}`);
      window.location.href = getLineLoginURL();
      return;
    }

    setIsRegistering(true);
    setError(null);

    try {
      const result = await apiClient.registerForEvent(event.id);
      setRegistrationResult(result);
    } catch (err) {
      setError(err instanceof Error ? err.message : '報名失敗，請稍後再試');
    } finally {
      setIsRegistering(false);
    }
  }, [isAuthenticated, event.id]);

  // Handle share
  const handleShare = useCallback(async () => {
    const shareUrl = window.location.href;

    if (navigator.share) {
      try {
        await navigator.share({
          title: `${event.location.name} 匹克球揪團`,
          text: `${event.skill_level_label} | ${isFull ? '已滿團' : `還缺 ${spotsRemaining} 人`}`,
          url: shareUrl,
        });
      } catch (error) {
        // User cancelled
      }
    } else {
      // Copy to clipboard
      try {
        await navigator.clipboard.writeText(shareUrl);
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
      } catch (error) {
        console.error('Failed to copy:', error);
      }
    }
  }, [event, isFull, spotsRemaining]);

  // Open in Google Maps
  const openMaps = useCallback(() => {
    const url = `https://www.google.com/maps/search/?api=1&query=${encodeURIComponent(
      event.location.address || event.location.name
    )}`;
    window.open(url, '_blank');
  }, [event.location]);

  return (
    <div className="container max-w-2xl py-6 px-4 sm:px-6">
      {/* Back Button */}
      <Button
        variant="ghost"
        size="sm"
        onClick={() => router.back()}
        className="mb-4 -ml-2"
      >
        <ArrowLeft className="mr-2 h-4 w-4" />
        返回
      </Button>

      {/* Header */}
      <div className="mb-6">
        <div className="flex items-center justify-between mb-2">
          <h1 className="text-2xl font-bold">
            {event.title || event.location.name}
          </h1>
          {getStatusBadge()}
        </div>

        {/* Host Info */}
        <div className="flex items-center gap-2 text-muted-foreground">
          <Avatar className="h-6 w-6">
            <AvatarImage src={event.host.avatar_url} />
            <AvatarFallback>{event.host.display_name[0]}</AvatarFallback>
          </Avatar>
          <span className="text-sm">{event.host.display_name} 發起</span>
        </div>
      </div>

      {/* Main Info Card */}
      <Card className="mb-6">
        <CardContent className="p-6 space-y-4">
          {/* Date & Time */}
          <div className="flex items-start gap-4">
            <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10">
              <Calendar className="h-5 w-5 text-primary" />
            </div>
            <div>
              <p className="font-medium">{formattedDate} ({dayOfWeek})</p>
              <p className="text-sm text-muted-foreground">
                {event.start_time}
                {event.end_time && ` - ${event.end_time}`}
              </p>
            </div>
          </div>

          {/* Location */}
          <div className="flex items-start gap-4">
            <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10">
              <MapPin className="h-5 w-5 text-primary" />
            </div>
            <div className="flex-1">
              <p className="font-medium">{event.location.name}</p>
              {event.location.address && (
                <p className="text-sm text-muted-foreground">
                  {event.location.address}
                </p>
              )}
              <Button
                variant="link"
                size="sm"
                className="h-auto p-0 text-primary"
                onClick={openMaps}
              >
                <ExternalLink className="mr-1 h-3 w-3" />
                在地圖上查看
              </Button>
            </div>
          </div>

          {/* Capacity */}
          <div className="flex items-start gap-4">
            <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10">
              <Users className="h-5 w-5 text-primary" />
            </div>
            <div>
              <p className="font-medium">
                {event.confirmed_count} / {event.capacity} 人
              </p>
              <p className="text-sm text-muted-foreground">
                {isFull
                  ? `${event.waitlist_count} 人候補中`
                  : `還缺 ${spotsRemaining} 人`}
              </p>
            </div>
          </div>

          {/* Skill Level */}
          <div className="flex items-start gap-4">
            <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10">
              <Target className="h-5 w-5 text-primary" />
            </div>
            <div>
              <p className="font-medium">{event.skill_level_label}</p>
              <p className="text-sm text-muted-foreground">程度要求</p>
            </div>
          </div>

          {/* Fee */}
          <div className="flex items-start gap-4">
            <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10">
              <DollarSign className="h-5 w-5 text-primary" />
            </div>
            <div>
              <p className="font-medium">
                {event.fee > 0 ? `NT$ ${event.fee}` : '免費'}
              </p>
              <p className="text-sm text-muted-foreground">參加費用</p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Description */}
      {event.description && (
        <Card className="mb-6">
          <CardHeader className="pb-2">
            <CardTitle className="text-base">活動說明</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-muted-foreground whitespace-pre-wrap">
              {event.description}
            </p>
          </CardContent>
        </Card>
      )}

      {/* Error Message */}
      {error && (
        <div className="mb-6 rounded-lg bg-destructive/10 p-4 text-destructive text-sm">
          {error}
        </div>
      )}

      {/* Registration Result */}
      {registrationResult && (
        <div className="mb-6 rounded-lg bg-green-50 p-4 text-green-800">
          <p className="font-medium">{registrationResult.message}</p>
          {registrationResult.status === 'waitlist' && (
            <p className="text-sm mt-1">
              你目前是候補第 {registrationResult.waitlist_position} 位
            </p>
          )}
        </div>
      )}

      {/* Action Buttons */}
      <div className="fixed bottom-16 left-0 right-0 border-t bg-background p-4 md:static md:border-0 md:p-0">
        <div className="container max-w-2xl flex gap-3">
          {/* Share Button */}
          <Button
            variant="outline"
            size="lg"
            className="shrink-0"
            onClick={handleShare}
          >
            {copied ? (
              <Check className="h-5 w-5 text-green-600" />
            ) : (
              <Share2 className="h-5 w-5" />
            )}
          </Button>

          {/* Register Button */}
          {!isCancelled && !isHost && !registrationResult && (
            <Button
              size="lg"
              className="flex-1"
              onClick={handleRegister}
              disabled={isRegistering || authLoading}
            >
              {isRegistering ? (
                '處理中...'
              ) : isFull ? (
                '排候補'
              ) : isAuthenticated ? (
                '+1 參加'
              ) : (
                'Line 登入報名'
              )}
            </Button>
          )}

          {/* Already Registered */}
          {registrationResult && (
            <Button size="lg" className="flex-1" variant="secondary" disabled>
              {registrationResult.status === 'confirmed' ? '已報名' : '已排候補'}
            </Button>
          )}

          {/* Host Actions */}
          {isHost && (
            <Button
              size="lg"
              className="flex-1"
              variant="outline"
              onClick={() => router.push(`/events/${event.id}/edit`)}
            >
              管理活動
            </Button>
          )}

          {/* Cancelled Event */}
          {isCancelled && !isHost && (
            <Button size="lg" className="flex-1" variant="secondary" disabled>
              活動已取消
            </Button>
          )}
        </div>
      </div>

      {/* Bottom Spacer for Fixed Button */}
      <div className="h-20 md:hidden" />
    </div>
  );
}
