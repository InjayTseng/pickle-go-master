'use client';

import { useState, useCallback, useEffect } from 'react';
import { useSearchParams } from 'next/navigation';
import Link from 'next/link';
import { CheckCircle, Copy, Check, Share2, ExternalLink } from 'lucide-react';

import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Input } from '@/components/ui/input';

interface PageProps {
  params: {
    id: string;
  };
}

export default function EventCreatedPage({ params }: PageProps) {
  const searchParams = useSearchParams();
  const shareUrl = searchParams.get('url') || `https://picklego.tw/events/${params.id}`;
  const [copied, setCopied] = useState(false);
  const [canShare, setCanShare] = useState(false);

  // Check if Web Share API is available on mount
  useEffect(() => {
    setCanShare(typeof navigator !== 'undefined' && typeof navigator.share === 'function');
  }, []);

  // Reset copied state after 2 seconds
  useEffect(() => {
    if (copied) {
      const timer = setTimeout(() => setCopied(false), 2000);
      return () => clearTimeout(timer);
    }
  }, [copied]);

  // Copy link to clipboard
  const handleCopy = useCallback(async () => {
    try {
      await navigator.clipboard.writeText(shareUrl);
      setCopied(true);
    } catch (error) {
      console.error('Failed to copy:', error);
      // Fallback for older browsers
      const textArea = document.createElement('textarea');
      textArea.value = shareUrl;
      document.body.appendChild(textArea);
      textArea.select();
      document.execCommand('copy');
      document.body.removeChild(textArea);
      setCopied(true);
    }
  }, [shareUrl]);

  // Share using Web Share API
  const handleShare = useCallback(async () => {
    if (navigator.share) {
      try {
        await navigator.share({
          title: 'Pickle Go 揪團',
          text: '快來參加這個匹克球活動！',
          url: shareUrl,
        });
      } catch (error) {
        // User cancelled or share failed
        console.log('Share cancelled or failed:', error);
      }
    } else {
      // Fallback to copy
      handleCopy();
    }
  }, [shareUrl, handleCopy]);

  return (
    <div className="container max-w-md py-12 px-4 sm:px-6">
      <div className="text-center">
        {/* Success Icon */}
        <div className="mx-auto mb-6 flex h-16 w-16 items-center justify-center rounded-full bg-green-100">
          <CheckCircle className="h-10 w-10 text-green-600" />
        </div>

        {/* Success Message */}
        <h1 className="text-2xl font-bold">活動建立成功！</h1>
        <p className="mt-2 text-muted-foreground">
          複製連結分享給你的球友們吧
        </p>

        {/* Share Card */}
        <Card className="mt-8">
          <CardContent className="p-6 space-y-4">
            {/* URL Input with Copy Button */}
            <div className="flex gap-2">
              <Input
                readOnly
                value={shareUrl}
                className="bg-muted text-sm"
                onClick={(e) => (e.target as HTMLInputElement).select()}
              />
              <Button
                variant="outline"
                size="icon"
                onClick={handleCopy}
                className="shrink-0"
              >
                {copied ? (
                  <Check className="h-4 w-4 text-green-600" />
                ) : (
                  <Copy className="h-4 w-4" />
                )}
              </Button>
            </div>

            {/* Copy Feedback */}
            {copied && (
              <p className="text-sm text-green-600 font-medium">
                已複製到剪貼簿！
              </p>
            )}

            {/* Action Buttons */}
            <div className="flex flex-col gap-3 pt-2">
              <Button onClick={handleCopy} className="w-full" size="lg">
                <Copy className="mr-2 h-4 w-4" />
                複製連結
              </Button>

              {/* Show share button on mobile */}
              {canShare && (
                <Button
                  variant="outline"
                  onClick={handleShare}
                  className="w-full"
                  size="lg"
                >
                  <Share2 className="mr-2 h-4 w-4" />
                  分享
                </Button>
              )}
            </div>
          </CardContent>
        </Card>

        {/* Navigation Links */}
        <div className="mt-8 flex flex-col gap-3">
          <Link href={`/events/${params.id}`}>
            <Button variant="outline" className="w-full">
              <ExternalLink className="mr-2 h-4 w-4" />
              查看活動頁面
            </Button>
          </Link>

          <Link href="/">
            <Button variant="ghost" className="w-full">
              返回首頁
            </Button>
          </Link>
        </div>

        {/* Tips */}
        <div className="mt-8 rounded-lg bg-muted p-4 text-left">
          <h3 className="font-medium text-sm mb-2">分享小提示</h3>
          <ul className="text-sm text-muted-foreground space-y-1">
            <li>- 分享到 Line 群組會顯示活動預覽卡片</li>
            <li>- 報名者可直接用 Line 登入報名</li>
            <li>- 你可以在「我的活動」管理報名狀況</li>
          </ul>
        </div>
      </div>
    </div>
  );
}
