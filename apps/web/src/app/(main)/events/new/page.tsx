'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import Script from 'next/script';
import { ArrowLeft } from 'lucide-react';

import { Button } from '@/components/ui/button';
import { EventForm } from '@/components/event/EventForm';
import { useAuthContext } from '@/contexts/AuthContext';

export default function CreateEventPage() {
  const router = useRouter();
  const { isAuthenticated, isLoading } = useAuthContext();

  // Redirect to login if not authenticated
  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      // Store the intended destination
      sessionStorage.setItem('redirectAfterLogin', '/events/new');
      router.push('/login');
    }
  }, [isAuthenticated, isLoading, router]);

  // Show loading while checking auth
  if (isLoading) {
    return (
      <div className="flex min-h-[50vh] items-center justify-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    );
  }

  // Don't render form if not authenticated
  if (!isAuthenticated) {
    return null;
  }

  return (
    <>
      {/* Load Google Maps API */}
      <Script
        src={`https://maps.googleapis.com/maps/api/js?key=${process.env.NEXT_PUBLIC_GOOGLE_MAPS_API_KEY}&libraries=places`}
        strategy="beforeInteractive"
      />

      <div className="container max-w-2xl py-6 px-4 sm:px-6">
        {/* Header */}
        <div className="mb-6">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => router.back()}
            className="mb-4 -ml-2"
          >
            <ArrowLeft className="mr-2 h-4 w-4" />
            返回
          </Button>
          <h1 className="text-2xl font-bold">建立活動</h1>
          <p className="mt-1 text-muted-foreground">
            填寫以下資訊來建立你的揪團活動
          </p>
        </div>

        {/* Form */}
        <EventForm />
      </div>
    </>
  );
}
