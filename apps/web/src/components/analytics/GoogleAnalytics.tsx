'use client';

import Script from 'next/script';
import { usePathname, useSearchParams } from 'next/navigation';
import { useEffect, Suspense } from 'react';

// GA Measurement ID
const GA_MEASUREMENT_ID = process.env.NEXT_PUBLIC_GA_MEASUREMENT_ID;

/**
 * Google Analytics 4 事件追蹤函數
 * 用於追蹤自訂事件
 */
export const trackEvent = (
  eventName: string,
  eventParams?: Record<string, string | number | boolean>
) => {
  if (typeof window !== 'undefined' && window.gtag && GA_MEASUREMENT_ID) {
    window.gtag('event', eventName, eventParams);
  }
};

/**
 * 追蹤頁面瀏覽
 */
export const trackPageView = (url: string) => {
  if (typeof window !== 'undefined' && window.gtag && GA_MEASUREMENT_ID) {
    window.gtag('config', GA_MEASUREMENT_ID, {
      page_path: url,
    });
  }
};

/**
 * 追蹤活動相關事件
 */
export const trackEventActions = {
  // 瀏覽活動
  viewEvent: (eventId: string, eventTitle: string) => {
    trackEvent('view_event', {
      event_id: eventId,
      event_title: eventTitle,
    });
  },

  // 報名活動
  registerEvent: (eventId: string, eventTitle: string) => {
    trackEvent('register_event', {
      event_id: eventId,
      event_title: eventTitle,
    });
  },

  // 取消報名
  cancelRegistration: (eventId: string, eventTitle: string) => {
    trackEvent('cancel_registration', {
      event_id: eventId,
      event_title: eventTitle,
    });
  },

  // 建立活動
  createEvent: (eventId: string) => {
    trackEvent('create_event', {
      event_id: eventId,
    });
  },

  // 分享活動
  shareEvent: (eventId: string, method: string) => {
    trackEvent('share', {
      content_type: 'event',
      item_id: eventId,
      method: method,
    });
  },

  // 使用地圖
  useMap: (action: string) => {
    trackEvent('use_map', {
      action: action,
    });
  },

  // 登入
  login: (method: string) => {
    trackEvent('login', {
      method: method,
    });
  },

  // 登出
  logout: () => {
    trackEvent('logout');
  },
};

/**
 * 頁面追蹤元件（內部使用）
 */
function PageViewTracker() {
  const pathname = usePathname();
  const searchParams = useSearchParams();

  useEffect(() => {
    if (pathname && GA_MEASUREMENT_ID) {
      const url = pathname + (searchParams?.toString() ? `?${searchParams.toString()}` : '');
      trackPageView(url);
    }
  }, [pathname, searchParams]);

  return null;
}

/**
 * Google Analytics 4 元件
 * 在 layout 中引入此元件以啟用 GA4 追蹤
 */
export function GoogleAnalytics() {
  // 如果沒有設定 GA ID，不渲染任何內容
  if (!GA_MEASUREMENT_ID) {
    return null;
  }

  return (
    <>
      {/* Google Analytics Script */}
      <Script
        strategy="afterInteractive"
        src={`https://www.googletagmanager.com/gtag/js?id=${GA_MEASUREMENT_ID}`}
      />
      <Script
        id="google-analytics"
        strategy="afterInteractive"
        dangerouslySetInnerHTML={{
          __html: `
            window.dataLayer = window.dataLayer || [];
            function gtag(){dataLayer.push(arguments);}
            gtag('js', new Date());
            gtag('config', '${GA_MEASUREMENT_ID}', {
              page_path: window.location.pathname,
              cookie_flags: 'SameSite=None;Secure',
            });
          `,
        }}
      />
      {/* 頁面追蹤 */}
      <Suspense fallback={null}>
        <PageViewTracker />
      </Suspense>
    </>
  );
}

/**
 * 擴展 Window 介面以支援 gtag
 */
declare global {
  interface Window {
    gtag: (
      command: 'config' | 'event' | 'js' | 'set',
      targetIdOrEventName: string | Date,
      config?: Record<string, unknown>
    ) => void;
    dataLayer: unknown[];
  }
}
