'use client';

import Link from 'next/link';
import { useAuth } from '@/hooks/useAuth';
import { Button } from '@/components/ui/button';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { MapPin, Plus, User, LogOut, Menu } from 'lucide-react';
import { useState } from 'react';

export function Header() {
  const { user, isAuthenticated, logout, isLoading } = useAuth();
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);

  const getInitials = (name: string) => {
    return name
      .split(' ')
      .map((n) => n[0])
      .join('')
      .toUpperCase()
      .slice(0, 2);
  };

  return (
    <header className="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="container flex h-14 items-center">
        {/* Logo */}
        <Link href="/" className="flex items-center space-x-2">
          <MapPin className="h-6 w-6 text-pickle-600" />
          <span className="font-bold text-pickle-600">Pickle Go</span>
        </Link>

        {/* Desktop Navigation */}
        <nav className="hidden md:flex flex-1 items-center justify-center space-x-6">
          <Link
            href="/"
            className="text-sm font-medium text-muted-foreground hover:text-primary transition-colors"
          >
            Find Events
          </Link>
          {isAuthenticated && (
            <>
              <Link
                href="/events/new"
                className="text-sm font-medium text-muted-foreground hover:text-primary transition-colors"
              >
                Create Event
              </Link>
              <Link
                href="/my/events"
                className="text-sm font-medium text-muted-foreground hover:text-primary transition-colors"
              >
                My Events
              </Link>
              <Link
                href="/my/registrations"
                className="text-sm font-medium text-muted-foreground hover:text-primary transition-colors"
              >
                My Registrations
              </Link>
            </>
          )}
        </nav>

        {/* Desktop Auth */}
        <div className="hidden md:flex items-center space-x-4">
          {isLoading ? (
            <div className="h-8 w-8 animate-pulse rounded-full bg-muted" />
          ) : isAuthenticated && user ? (
            <div className="flex items-center space-x-4">
              <Link href="/events/new">
                <Button size="sm" className="gap-2">
                  <Plus className="h-4 w-4" />
                  Create
                </Button>
              </Link>
              <div className="relative group">
                <button className="flex items-center space-x-2">
                  <Avatar className="h-8 w-8">
                    {user.avatar_url ? (
                      <AvatarImage src={user.avatar_url} alt={user.display_name} />
                    ) : null}
                    <AvatarFallback>{getInitials(user.display_name)}</AvatarFallback>
                  </Avatar>
                </button>
                <div className="absolute right-0 mt-2 w-48 origin-top-right rounded-md bg-background shadow-lg ring-1 ring-black ring-opacity-5 opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all">
                  <div className="py-1">
                    <div className="px-4 py-2 text-sm text-muted-foreground border-b">
                      {user.display_name}
                    </div>
                    <Link
                      href="/my/events"
                      className="flex items-center px-4 py-2 text-sm hover:bg-muted"
                    >
                      <User className="mr-2 h-4 w-4" />
                      My Events
                    </Link>
                    <button
                      onClick={logout}
                      className="flex w-full items-center px-4 py-2 text-sm hover:bg-muted text-destructive"
                    >
                      <LogOut className="mr-2 h-4 w-4" />
                      Logout
                    </button>
                  </div>
                </div>
              </div>
            </div>
          ) : (
            <Link href="/login">
              <Button variant="line" size="sm">
                Login with LINE
              </Button>
            </Link>
          )}
        </div>

        {/* Mobile Menu Button */}
        <button
          className="md:hidden ml-auto"
          onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
        >
          <Menu className="h-6 w-6" />
        </button>
      </div>

      {/* Mobile Menu */}
      {isMobileMenuOpen && (
        <div className="md:hidden border-t bg-background">
          <nav className="container py-4 space-y-4">
            <Link
              href="/"
              className="block text-sm font-medium"
              onClick={() => setIsMobileMenuOpen(false)}
            >
              Find Events
            </Link>
            {isAuthenticated ? (
              <>
                <Link
                  href="/events/new"
                  className="block text-sm font-medium"
                  onClick={() => setIsMobileMenuOpen(false)}
                >
                  Create Event
                </Link>
                <Link
                  href="/my/events"
                  className="block text-sm font-medium"
                  onClick={() => setIsMobileMenuOpen(false)}
                >
                  My Events
                </Link>
                <Link
                  href="/my/registrations"
                  className="block text-sm font-medium"
                  onClick={() => setIsMobileMenuOpen(false)}
                >
                  My Registrations
                </Link>
                <button
                  onClick={() => {
                    logout();
                    setIsMobileMenuOpen(false);
                  }}
                  className="block text-sm font-medium text-destructive"
                >
                  Logout
                </button>
              </>
            ) : (
              <Link href="/login" onClick={() => setIsMobileMenuOpen(false)}>
                <Button variant="line" className="w-full">
                  Login with LINE
                </Button>
              </Link>
            )}
          </nav>
        </div>
      )}
    </header>
  );
}
