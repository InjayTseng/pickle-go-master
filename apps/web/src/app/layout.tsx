import type { Metadata } from 'next';
import { Inter } from 'next/font/google';
import './globals.css';
import { Providers } from './providers';
import { Header } from '@/components/layout/Header';
import { MobileNav } from '@/components/layout/MobileNav';

const inter = Inter({ subsets: ['latin'], variable: '--font-inter' });

export const metadata: Metadata = {
  title: {
    default: 'Pickle Go - Find and Join Pickleball Games',
    template: '%s | Pickle Go',
  },
  description: 'The easiest way to find and join pickleball games near you. Create or join events in 30 seconds.',
  keywords: ['pickleball', 'sports', 'events', 'games', 'community'],
  openGraph: {
    type: 'website',
    locale: 'zh_TW',
    url: 'https://picklego.tw',
    siteName: 'Pickle Go',
    title: 'Pickle Go - Find and Join Pickleball Games',
    description: 'The easiest way to find and join pickleball games near you. Create or join events in 30 seconds.',
  },
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="zh-TW">
      <body className={`${inter.variable} font-sans antialiased`}>
        <Providers>
          <div className="relative flex min-h-screen flex-col">
            <Header />
            <main className="flex-1 pb-16 md:pb-0">{children}</main>
            <MobileNav />
          </div>
        </Providers>
      </body>
    </html>
  );
}
