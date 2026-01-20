import type { Metadata, Viewport } from 'next';
import { Inter } from 'next/font/google';
import './globals.css';
import { Providers } from './providers';
import { Header } from '@/components/layout/Header';
import { MobileNav } from '@/components/layout/MobileNav';

const inter = Inter({ subsets: ['latin'], variable: '--font-inter' });

const BASE_URL = process.env.NEXT_PUBLIC_BASE_URL || 'https://picklego.tw';

export const viewport: Viewport = {
  width: 'device-width',
  initialScale: 1,
  maximumScale: 5,
  themeColor: [
    { media: '(prefers-color-scheme: light)', color: '#ffffff' },
    { media: '(prefers-color-scheme: dark)', color: '#0a0a0a' },
  ],
};

export const metadata: Metadata = {
  metadataBase: new URL(BASE_URL),
  title: {
    default: 'Pickle Go - 匹克球揪團平台',
    template: '%s | Pickle Go',
  },
  description: '台灣最方便的匹克球揪團平台。30 秒內找到附近球局並報名，輕鬆揪團打球！',
  keywords: [
    '匹克球',
    'pickleball',
    '揪團',
    '運動',
    '球友',
    '台北匹克球',
    '匹克球活動',
    '找球局',
    '打球',
    '匹克球新手',
  ],
  authors: [{ name: 'Pickle Go Team' }],
  creator: 'Pickle Go',
  publisher: 'Pickle Go',
  formatDetection: {
    email: false,
    address: false,
    telephone: false,
  },
  alternates: {
    canonical: BASE_URL,
    languages: {
      'zh-TW': BASE_URL,
    },
  },
  openGraph: {
    type: 'website',
    locale: 'zh_TW',
    url: BASE_URL,
    siteName: 'Pickle Go',
    title: 'Pickle Go - 匹克球揪團平台',
    description: '台灣最方便的匹克球揪團平台。30 秒內找到附近球局並報名！',
    images: [
      {
        url: `${BASE_URL}/og-default.png`,
        width: 1200,
        height: 630,
        alt: 'Pickle Go - 匹克球揪團平台',
      },
    ],
  },
  twitter: {
    card: 'summary_large_image',
    title: 'Pickle Go - 匹克球揪團平台',
    description: '台灣最方便的匹克球揪團平台。30 秒內找到附近球局並報名！',
    images: [`${BASE_URL}/og-default.png`],
  },
  robots: {
    index: true,
    follow: true,
    googleBot: {
      index: true,
      follow: true,
      'max-video-preview': -1,
      'max-image-preview': 'large',
      'max-snippet': -1,
    },
  },
  verification: {
    // Add your verification codes here when available
    // google: 'your-google-verification-code',
  },
  category: 'sports',
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
