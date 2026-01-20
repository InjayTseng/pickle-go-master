'use client';

import * as React from 'react';
import Image from 'next/image';
import * as AvatarPrimitive from '@radix-ui/react-avatar';

import { cn } from '@/lib/utils';

interface OptimizedAvatarProps {
  src?: string | null;
  alt?: string;
  fallback?: string;
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl';
  className?: string;
}

const sizeMap = {
  xs: { container: 'h-5 w-5', image: 20, text: 'text-xs' },
  sm: { container: 'h-6 w-6', image: 24, text: 'text-xs' },
  md: { container: 'h-8 w-8', image: 32, text: 'text-sm' },
  lg: { container: 'h-10 w-10', image: 40, text: 'text-base' },
  xl: { container: 'h-12 w-12', image: 48, text: 'text-lg' },
};

/**
 * Optimized Avatar component using Next.js Image for better performance
 * - Automatic lazy loading
 * - Image optimization (WebP, AVIF)
 * - Proper sizing
 * - Blur placeholder
 */
export function OptimizedAvatar({
  src,
  alt = 'Avatar',
  fallback,
  size = 'md',
  className,
}: OptimizedAvatarProps) {
  const [hasError, setHasError] = React.useState(false);
  const { container, image, text } = sizeMap[size];

  // Generate fallback initials from alt text if no fallback provided
  const displayFallback = fallback || (alt ? alt.charAt(0).toUpperCase() : 'U');

  // Check if we should show the image
  const showImage = src && !hasError;

  return (
    <div
      className={cn(
        'relative flex shrink-0 overflow-hidden rounded-full',
        container,
        className
      )}
    >
      {showImage ? (
        <Image
          src={src}
          alt={alt}
          width={image}
          height={image}
          className="aspect-square h-full w-full object-cover"
          onError={() => setHasError(true)}
          loading="lazy"
          // Use blur placeholder for better UX
          placeholder="empty"
          // Quality optimization
          quality={75}
          // Unoptimized for external URLs (Line CDN, etc.)
          unoptimized={
            src.includes('profile.line-scdn.net') ||
            src.includes('lh3.googleusercontent.com')
          }
        />
      ) : (
        <div
          className={cn(
            'flex h-full w-full items-center justify-center rounded-full bg-muted text-muted-foreground font-medium',
            text
          )}
        >
          {displayFallback}
        </div>
      )}
    </div>
  );
}

// Re-export original Avatar components for backwards compatibility
const Avatar = React.forwardRef<
  React.ElementRef<typeof AvatarPrimitive.Root>,
  React.ComponentPropsWithoutRef<typeof AvatarPrimitive.Root>
>(({ className, ...props }, ref) => (
  <AvatarPrimitive.Root
    ref={ref}
    className={cn(
      'relative flex h-10 w-10 shrink-0 overflow-hidden rounded-full',
      className
    )}
    {...props}
  />
));
Avatar.displayName = AvatarPrimitive.Root.displayName;

const AvatarImage = React.forwardRef<
  React.ElementRef<typeof AvatarPrimitive.Image>,
  React.ComponentPropsWithoutRef<typeof AvatarPrimitive.Image>
>(({ className, ...props }, ref) => (
  <AvatarPrimitive.Image
    ref={ref}
    className={cn('aspect-square h-full w-full', className)}
    {...props}
  />
));
AvatarImage.displayName = AvatarPrimitive.Image.displayName;

const AvatarFallback = React.forwardRef<
  React.ElementRef<typeof AvatarPrimitive.Fallback>,
  React.ComponentPropsWithoutRef<typeof AvatarPrimitive.Fallback>
>(({ className, ...props }, ref) => (
  <AvatarPrimitive.Fallback
    ref={ref}
    className={cn(
      'flex h-full w-full items-center justify-center rounded-full bg-muted',
      className
    )}
    {...props}
  />
));
AvatarFallback.displayName = AvatarPrimitive.Fallback.displayName;

export { Avatar, AvatarImage, AvatarFallback };
