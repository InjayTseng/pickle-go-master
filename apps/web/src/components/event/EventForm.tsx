'use client';

import { useState, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import { format, addDays, startOfDay } from 'date-fns';
import { zhTW } from 'date-fns/locale';
import { Calendar, Clock, MapPin, Users, DollarSign, Target, FileText } from 'lucide-react';

import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { PlacesAutocomplete, PlaceResult } from '@/components/form/PlacesAutocomplete';
import { apiClient } from '@/lib/api-client';

// Skill level options
const SKILL_LEVELS = [
  { value: 'beginner', label: '新手友善 (2.0-2.5)' },
  { value: 'intermediate', label: '中階 (2.5-3.5)' },
  { value: 'advanced', label: '進階 (3.5-4.5)' },
  { value: 'expert', label: '高階 (4.5+)' },
  { value: 'any', label: '不限程度' },
];

// Capacity options (4-20)
const CAPACITY_OPTIONS = Array.from({ length: 17 }, (_, i) => i + 4);

// Time options (30-minute intervals)
const TIME_OPTIONS = Array.from({ length: 48 }, (_, i) => {
  const hours = Math.floor(i / 2);
  const minutes = (i % 2) * 30;
  return `${hours.toString().padStart(2, '0')}:${minutes.toString().padStart(2, '0')}`;
});

interface FormData {
  title: string;
  description: string;
  eventDate: string;
  startTime: string;
  endTime: string;
  location: {
    name: string;
    address: string;
    lat: number;
    lng: number;
    googlePlaceId: string;
  } | null;
  capacity: number;
  skillLevel: string;
  fee: number;
}

interface FormErrors {
  eventDate?: string;
  startTime?: string;
  location?: string;
  capacity?: string;
  skillLevel?: string;
  fee?: string;
  general?: string;
}

export function EventForm() {
  const router = useRouter();
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [errors, setErrors] = useState<FormErrors>({});

  // Initialize with default values
  const tomorrow = format(addDays(new Date(), 1), 'yyyy-MM-dd');

  const [formData, setFormData] = useState<FormData>({
    title: '',
    description: '',
    eventDate: tomorrow,
    startTime: '19:00',
    endTime: '21:00',
    location: null,
    capacity: 4,
    skillLevel: 'any',
    fee: 0,
  });

  // Validate form
  const validateForm = useCallback((): boolean => {
    const newErrors: FormErrors = {};

    // Date validation
    const selectedDate = new Date(formData.eventDate);
    const today = startOfDay(new Date());
    if (selectedDate < today) {
      newErrors.eventDate = '活動日期不可早於今天';
    }

    // Time validation
    if (!formData.startTime) {
      newErrors.startTime = '請選擇開始時間';
    }

    // Location validation
    if (!formData.location) {
      newErrors.location = '請選擇活動地點';
    }

    // Capacity validation
    if (formData.capacity < 4 || formData.capacity > 20) {
      newErrors.capacity = '人數需在 4-20 人之間';
    }

    // Skill level validation
    if (!formData.skillLevel) {
      newErrors.skillLevel = '請選擇程度要求';
    }

    // Fee validation
    if (formData.fee < 0 || formData.fee > 9999) {
      newErrors.fee = '費用需在 0-9999 元之間';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  }, [formData]);

  // Handle form submission
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validateForm()) {
      return;
    }

    if (!formData.location) {
      return;
    }

    setIsSubmitting(true);
    setErrors({});

    try {
      const response = await apiClient.createEvent({
        title: formData.title || undefined,
        description: formData.description || undefined,
        event_date: formData.eventDate,
        start_time: formData.startTime,
        end_time: formData.endTime || undefined,
        location: {
          name: formData.location.name,
          address: formData.location.address || undefined,
          lat: formData.location.lat,
          lng: formData.location.lng,
          google_place_id: formData.location.googlePlaceId || undefined,
        },
        capacity: formData.capacity,
        skill_level: formData.skillLevel,
        fee: formData.fee,
      });

      // Redirect to success page
      router.push(`/events/${response.id}/created?url=${encodeURIComponent(response.share_url)}`);
    } catch (error) {
      console.error('Failed to create event:', error);
      setErrors({
        general: error instanceof Error ? error.message : '建立活動失敗，請稍後再試',
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  // Handle place selection
  const handlePlaceSelect = useCallback((place: PlaceResult) => {
    setFormData((prev) => ({
      ...prev,
      location: {
        name: place.name,
        address: place.address,
        lat: place.lat,
        lng: place.lng,
        googlePlaceId: place.placeId,
      },
    }));
    // Clear location error
    setErrors((prev) => ({ ...prev, location: undefined }));
  }, []);

  // Get minimum date (today)
  const minDate = format(new Date(), 'yyyy-MM-dd');

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      {/* General Error */}
      {errors.general && (
        <div className="rounded-lg bg-destructive/10 p-4 text-destructive text-sm">
          {errors.general}
        </div>
      )}

      {/* Date and Time */}
      <Card>
        <CardHeader className="pb-4">
          <CardTitle className="flex items-center gap-2 text-lg">
            <Calendar className="h-5 w-5" />
            日期與時間
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* Date */}
          <div className="space-y-2">
            <Label htmlFor="eventDate">活動日期 *</Label>
            <Input
              id="eventDate"
              type="date"
              min={minDate}
              value={formData.eventDate}
              onChange={(e) =>
                setFormData((prev) => ({ ...prev, eventDate: e.target.value }))
              }
              className={errors.eventDate ? 'border-destructive' : ''}
            />
            {errors.eventDate && (
              <p className="text-sm text-destructive">{errors.eventDate}</p>
            )}
          </div>

          {/* Time */}
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="startTime">開始時間 *</Label>
              <Select
                value={formData.startTime}
                onValueChange={(value) =>
                  setFormData((prev) => ({ ...prev, startTime: value }))
                }
              >
                <SelectTrigger className={errors.startTime ? 'border-destructive' : ''}>
                  <SelectValue placeholder="選擇時間" />
                </SelectTrigger>
                <SelectContent>
                  {TIME_OPTIONS.map((time) => (
                    <SelectItem key={time} value={time}>
                      {time}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              {errors.startTime && (
                <p className="text-sm text-destructive">{errors.startTime}</p>
              )}
            </div>
            <div className="space-y-2">
              <Label htmlFor="endTime">結束時間</Label>
              <Select
                value={formData.endTime}
                onValueChange={(value) =>
                  setFormData((prev) => ({ ...prev, endTime: value }))
                }
              >
                <SelectTrigger>
                  <SelectValue placeholder="選擇時間" />
                </SelectTrigger>
                <SelectContent>
                  {TIME_OPTIONS.map((time) => (
                    <SelectItem key={time} value={time}>
                      {time}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Location */}
      <Card>
        <CardHeader className="pb-4">
          <CardTitle className="flex items-center gap-2 text-lg">
            <MapPin className="h-5 w-5" />
            活動地點
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="location">搜尋地點 *</Label>
            <PlacesAutocomplete
              value={formData.location?.name || ''}
              onPlaceSelect={handlePlaceSelect}
              placeholder="搜尋球場或地點..."
              className={errors.location ? 'border-destructive' : ''}
            />
            {errors.location && (
              <p className="text-sm text-destructive">{errors.location}</p>
            )}
          </div>
          {formData.location && (
            <div className="rounded-lg bg-muted p-3 text-sm">
              <p className="font-medium">{formData.location.name}</p>
              {formData.location.address && (
                <p className="text-muted-foreground">{formData.location.address}</p>
              )}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Event Details */}
      <Card>
        <CardHeader className="pb-4">
          <CardTitle className="flex items-center gap-2 text-lg">
            <Target className="h-5 w-5" />
            活動設定
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* Capacity */}
          <div className="space-y-2">
            <Label htmlFor="capacity" className="flex items-center gap-2">
              <Users className="h-4 w-4" />
              人數上限 *
            </Label>
            <Select
              value={formData.capacity.toString()}
              onValueChange={(value) =>
                setFormData((prev) => ({ ...prev, capacity: parseInt(value) }))
              }
            >
              <SelectTrigger className={errors.capacity ? 'border-destructive' : ''}>
                <SelectValue placeholder="選擇人數" />
              </SelectTrigger>
              <SelectContent>
                {CAPACITY_OPTIONS.map((num) => (
                  <SelectItem key={num} value={num.toString()}>
                    {num} 人
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            {errors.capacity && (
              <p className="text-sm text-destructive">{errors.capacity}</p>
            )}
          </div>

          {/* Skill Level */}
          <div className="space-y-2">
            <Label htmlFor="skillLevel">程度要求 *</Label>
            <Select
              value={formData.skillLevel}
              onValueChange={(value) =>
                setFormData((prev) => ({ ...prev, skillLevel: value }))
              }
            >
              <SelectTrigger className={errors.skillLevel ? 'border-destructive' : ''}>
                <SelectValue placeholder="選擇程度" />
              </SelectTrigger>
              <SelectContent>
                {SKILL_LEVELS.map((level) => (
                  <SelectItem key={level.value} value={level.value}>
                    {level.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            {errors.skillLevel && (
              <p className="text-sm text-destructive">{errors.skillLevel}</p>
            )}
          </div>

          {/* Fee */}
          <div className="space-y-2">
            <Label htmlFor="fee" className="flex items-center gap-2">
              <DollarSign className="h-4 w-4" />
              費用 (元)
            </Label>
            <Input
              id="fee"
              type="number"
              min="0"
              max="9999"
              value={formData.fee}
              onChange={(e) =>
                setFormData((prev) => ({
                  ...prev,
                  fee: Math.min(9999, Math.max(0, parseInt(e.target.value) || 0)),
                }))
              }
              className={errors.fee ? 'border-destructive' : ''}
            />
            {errors.fee && (
              <p className="text-sm text-destructive">{errors.fee}</p>
            )}
            <p className="text-xs text-muted-foreground">
              0 表示免費參加
            </p>
          </div>
        </CardContent>
      </Card>

      {/* Optional Info */}
      <Card>
        <CardHeader className="pb-4">
          <CardTitle className="flex items-center gap-2 text-lg">
            <FileText className="h-5 w-5" />
            其他資訊 (選填)
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* Title */}
          <div className="space-y-2">
            <Label htmlFor="title">活動標題</Label>
            <Input
              id="title"
              value={formData.title}
              onChange={(e) =>
                setFormData((prev) => ({ ...prev, title: e.target.value }))
              }
              placeholder="例如：週末輕鬆打"
              maxLength={200}
            />
          </div>

          {/* Description */}
          <div className="space-y-2">
            <Label htmlFor="description">備註說明</Label>
            <Textarea
              id="description"
              value={formData.description}
              onChange={(e) =>
                setFormData((prev) => ({ ...prev, description: e.target.value }))
              }
              placeholder="例如：歡迎新手、請自備球拍..."
              rows={3}
            />
          </div>
        </CardContent>
      </Card>

      {/* Submit Button */}
      <Button
        type="submit"
        size="lg"
        className="w-full"
        disabled={isSubmitting}
      >
        {isSubmitting ? '建立中...' : '建立活動'}
      </Button>
    </form>
  );
}
