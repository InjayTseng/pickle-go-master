import type { Metadata } from 'next';
import { Inter } from 'next/font/google';
import './globals.css';

const inter = Inter({ subsets: ['latin'], variable: '--font-inter' });

export const metadata: Metadata = {
  title: {
    default: 'Pickle Go - 找球局、揪球友',
    template: '%s | Pickle Go',
  },
  description: '台灣最方便的匹克球揪團平台，30 秒找到附近球局並報名',
  keywords: ['匹克球', 'pickleball', '揪團', '找球友', '運動'],
  openGraph: {
    type: 'website',
    locale: 'zh_TW',
    url: 'https://picklego.tw',
    siteName: 'Pickle Go',
    title: 'Pickle Go - 找球局、揪球友',
    description: '台灣最方便的匹克球揪團平台，30 秒找到附近球局並報名',
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
        {children}
      </body>
    </html>
  );
}
