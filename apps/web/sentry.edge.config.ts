/**
 * Sentry Edge Runtime 配置
 * 用於 Next.js Edge Functions 錯誤監控
 */
import * as Sentry from '@sentry/nextjs';

// 僅在有 DSN 時初始化 Sentry
const SENTRY_DSN = process.env.NEXT_PUBLIC_SENTRY_DSN;

if (SENTRY_DSN) {
  Sentry.init({
    dsn: SENTRY_DSN,

    // 環境設定
    environment: process.env.NODE_ENV,

    // 效能監控取樣率
    tracesSampleRate: process.env.NODE_ENV === 'production' ? 0.1 : 1.0,

    // 在傳送前處理事件
    beforeSend(event, hint) {
      if (process.env.NODE_ENV === 'development') {
        return null;
      }
      return event;
    },
  });
}
