'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { cn } from '@/lib/utils';
import { useAuth } from '@/hooks/useAuth';
import { MapPin, PlusCircle, Calendar, User } from 'lucide-react';

interface NavItem {
  href: string;
  label: string;
  icon: React.ComponentType<{ className?: string }>;
  requiresAuth?: boolean;
}

const navItems: NavItem[] = [
  {
    href: '/',
    label: 'Explore',
    icon: MapPin,
  },
  {
    href: '/events/new',
    label: 'Create',
    icon: PlusCircle,
    requiresAuth: true,
  },
  {
    href: '/my/registrations',
    label: 'My Events',
    icon: Calendar,
    requiresAuth: true,
  },
  {
    href: '/profile',
    label: 'Profile',
    icon: User,
    requiresAuth: true,
  },
];

export function MobileNav() {
  const pathname = usePathname();
  const { isAuthenticated } = useAuth();

  const filteredItems = navItems.filter(
    (item) => !item.requiresAuth || isAuthenticated
  );

  return (
    <nav className="md:hidden fixed bottom-0 left-0 right-0 z-50 bg-background border-t safe-area-inset">
      <div className="flex items-center justify-around h-16">
        {filteredItems.map((item) => {
          const isActive = pathname === item.href;
          const Icon = item.icon;

          return (
            <Link
              key={item.href}
              href={item.href}
              className={cn(
                'flex flex-col items-center justify-center flex-1 h-full space-y-1',
                'transition-colors',
                isActive
                  ? 'text-primary'
                  : 'text-muted-foreground hover:text-primary'
              )}
            >
              <Icon className="h-5 w-5" />
              <span className="text-xs">{item.label}</span>
            </Link>
          );
        })}
        {!isAuthenticated && (
          <Link
            href="/login"
            className={cn(
              'flex flex-col items-center justify-center flex-1 h-full space-y-1',
              'transition-colors text-muted-foreground hover:text-primary'
            )}
          >
            <User className="h-5 w-5" />
            <span className="text-xs">Login</span>
          </Link>
        )}
      </div>
    </nav>
  );
}
