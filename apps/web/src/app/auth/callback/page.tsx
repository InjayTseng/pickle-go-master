'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/hooks/useAuth';
import { handleLineCallback } from '@/lib/auth';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Spinner } from '@/components/ui/spinner';
import { AlertCircle } from 'lucide-react';

export default function AuthCallbackPage() {
  const router = useRouter();
  const { login, isAuthenticated } = useAuth();
  const [error, setError] = useState<string | null>(null);
  const [isProcessing, setIsProcessing] = useState(true);

  useEffect(() => {
    const processCallback = async () => {
      const result = handleLineCallback();

      if ('error' in result) {
        setError(result.error);
        setIsProcessing(false);
        return;
      }

      try {
        await login(result.code, result.state);

        // Get redirect URL from session storage
        const redirectTo = sessionStorage.getItem('login_redirect') || '/';
        sessionStorage.removeItem('login_redirect');

        router.push(redirectTo);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Login failed');
        setIsProcessing(false);
      }
    };

    if (!isAuthenticated) {
      processCallback();
    } else {
      // Already authenticated, redirect
      const redirectTo = sessionStorage.getItem('login_redirect') || '/';
      sessionStorage.removeItem('login_redirect');
      router.push(redirectTo);
    }
  }, [login, router, isAuthenticated]);

  if (error) {
    return (
      <div className="flex min-h-[80vh] items-center justify-center px-4">
        <Card className="w-full max-w-md">
          <CardHeader className="text-center">
            <div className="flex justify-center mb-4">
              <AlertCircle className="h-12 w-12 text-destructive" />
            </div>
            <CardTitle>Login Failed</CardTitle>
            <CardDescription>{error}</CardDescription>
          </CardHeader>
          <CardContent>
            <Button
              className="w-full"
              onClick={() => router.push('/login')}
            >
              Try Again
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="flex min-h-[80vh] items-center justify-center px-4">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <div className="flex justify-center mb-4">
            <Spinner size="lg" />
          </div>
          <CardTitle>Signing in...</CardTitle>
          <CardDescription>
            Please wait while we complete your login
          </CardDescription>
        </CardHeader>
      </Card>
    </div>
  );
}
