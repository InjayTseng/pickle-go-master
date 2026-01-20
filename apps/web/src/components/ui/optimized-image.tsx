'use client';

import Image, { ImageProps } from 'next/image';
import { useState } from 'react';
import { cn } from '@/lib/utils';

interface OptimizedImageProps extends Omit<ImageProps, 'onError'> {
  fallbackSrc?: string;
  showPlaceholder?: boolean;
  containerClassName?: string;
}

/**
 * Optimized Image component with:
 * - Automatic WebP/AVIF optimization
 * - Lazy loading
 * - Error handling with fallback
 * - Loading placeholder
 * - Responsive sizing
 */
export function OptimizedImage({
  src,
  alt,
  fallbackSrc,
  showPlaceholder = true,
  containerClassName,
  className,
  ...props
}: OptimizedImageProps) {
  const [hasError, setHasError] = useState(false);
  const [isLoading, setIsLoading] = useState(true);

  const handleError = () => {
    setHasError(true);
    setIsLoading(false);
  };

  const handleLoad = () => {
    setIsLoading(false);
  };

  // Use fallback if there's an error and fallback is provided
  const imageSrc = hasError && fallbackSrc ? fallbackSrc : src;

  // If we have an error and no fallback, show placeholder
  if (hasError && !fallbackSrc) {
    return (
      <div
        className={cn(
          'flex items-center justify-center bg-muted text-muted-foreground',
          containerClassName
        )}
      >
        <svg
          className="h-1/3 w-1/3 opacity-50"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={1}
            d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"
          />
        </svg>
      </div>
    );
  }

  return (
    <div className={cn('relative overflow-hidden', containerClassName)}>
      {/* Loading placeholder */}
      {isLoading && showPlaceholder && (
        <div className="absolute inset-0 animate-pulse bg-muted" />
      )}

      <Image
        src={imageSrc}
        alt={alt}
        className={cn(
          'transition-opacity duration-300',
          isLoading ? 'opacity-0' : 'opacity-100',
          className
        )}
        onError={handleError}
        onLoad={handleLoad}
        loading="lazy"
        quality={80}
        {...props}
      />
    </div>
  );
}

/**
 * Responsive hero image optimized for above-the-fold content
 */
interface HeroImageProps {
  src: string;
  alt: string;
  priority?: boolean;
  className?: string;
}

export function HeroImage({
  src,
  alt,
  priority = true,
  className,
}: HeroImageProps) {
  return (
    <div className={cn('relative w-full', className)}>
      <Image
        src={src}
        alt={alt}
        fill
        priority={priority}
        quality={85}
        sizes="100vw"
        className="object-cover"
      />
    </div>
  );
}

/**
 * Event card thumbnail with fixed aspect ratio
 */
interface EventThumbnailProps {
  src?: string;
  alt: string;
  className?: string;
}

export function EventThumbnail({
  src,
  alt,
  className,
}: EventThumbnailProps) {
  const defaultImage = '/images/default-event.jpg';

  return (
    <OptimizedImage
      src={src || defaultImage}
      alt={alt}
      width={320}
      height={180}
      fallbackSrc={defaultImage}
      className={cn('aspect-video object-cover rounded-lg', className)}
    />
  );
}
