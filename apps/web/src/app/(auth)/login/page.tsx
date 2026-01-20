'use client';

import { Suspense, useEffect } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import Link from 'next/link';
import { useAuth } from '@/hooks/useAuth';
import { getLineLoginURL } from '@/lib/auth';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { MapPin } from 'lucide-react';

function LoginContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { isAuthenticated, isLoading } = useAuth();

  const redirectTo = searchParams.get('redirect') || '/';

  useEffect(() => {
    if (isAuthenticated && !isLoading) {
      router.push(redirectTo);
    }
  }, [isAuthenticated, isLoading, redirectTo, router]);

  const handleLineLogin = () => {
    // Store redirect URL for after login
    if (typeof window !== 'undefined') {
      sessionStorage.setItem('login_redirect', redirectTo);
    }
    window.location.href = getLineLoginURL();
  };

  if (isLoading) {
    return (
      <div className="flex min-h-[80vh] items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
      </div>
    );
  }

  if (isAuthenticated) {
    return null;
  }

  return (
    <div className="flex min-h-[80vh] items-center justify-center px-4">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <div className="flex justify-center mb-4">
            <div className="flex items-center space-x-2">
              <MapPin className="h-8 w-8 text-pickle-600" />
              <span className="text-2xl font-bold text-pickle-600">Pickle Go</span>
            </div>
          </div>
          <CardTitle className="text-xl">Welcome Back</CardTitle>
          <CardDescription>
            Sign in to find and join pickleball games near you
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <Button
            variant="line"
            className="w-full h-12 text-base"
            onClick={handleLineLogin}
          >
            <svg
              className="mr-2 h-5 w-5"
              viewBox="0 0 24 24"
              fill="currentColor"
            >
              <path d="M19.365 9.863c.349 0 .63.285.63.631 0 .345-.281.63-.63.63H17.61v1.125h1.755c.349 0 .63.283.63.63 0 .344-.281.629-.63.629H17.61v1.125h1.755c.349 0 .63.285.63.63 0 .348-.281.631-.63.631h-2.386c-.345 0-.627-.283-.627-.631V9.862c0-.345.282-.63.63-.63h2.383zm-3.855 6.165c0 .345-.281.63-.628.63-.193 0-.377-.087-.498-.23l-2.437-3.259v2.86c0 .345-.281.629-.628.629-.345 0-.627-.284-.627-.629v-4.636c0-.345.282-.63.63-.63.189 0 .371.087.493.228l2.44 3.258V8.489c0-.345.281-.63.628-.63.346 0 .627.285.627.63v7.539zm-6.341-4.009c.345 0 .629.285.629.631 0 .345-.284.63-.63.63H7.04v1.125h2.131c.345 0 .627.283.627.63 0 .344-.282.629-.627.629H6.412c-.345 0-.63-.283-.63-.631V9.862c0-.345.285-.63.63-.63h2.759c.345 0 .63.285.63.631 0 .345-.285.63-.63.63H7.04v1.126h2.129zM4.89 15.391H2.647c-.345 0-.63-.283-.63-.631V8.489c0-.346.285-.63.63-.63.346 0 .63.284.63.63v5.642H4.89c.345 0 .63.283.63.63 0 .348-.285.63-.63.63M24 10.314C24 4.943 18.615.573 12 .573S0 4.943 0 10.314c0 4.811 4.27 8.842 10.035 9.608.391.082.923.258 1.058.59.12.301.079.766.038 1.08l-.164 1.02c-.045.301-.24 1.186 1.049.645 1.291-.539 6.916-4.078 9.436-6.975C23.176 14.393 24 12.458 24 10.314" />
            </svg>
            Login with LINE
          </Button>

          <div className="text-center text-sm text-muted-foreground">
            <p>By signing in, you agree to our</p>
            <p>
              <Link href="/terms" className="underline hover:text-primary">
                Terms of Service
              </Link>
              {' and '}
              <Link href="/privacy" className="underline hover:text-primary">
                Privacy Policy
              </Link>
            </p>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

function LoginFallback() {
  return (
    <div className="flex min-h-[80vh] items-center justify-center">
      <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
    </div>
  );
}

export default function LoginPage() {
  return (
    <Suspense fallback={<LoginFallback />}>
      <LoginContent />
    </Suspense>
  );
}
