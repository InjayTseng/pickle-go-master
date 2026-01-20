/**
 * Sentry 伺服器端配置
 * 用於 Next.js 伺服器端錯誤監控
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

    // 設定要忽略的錯誤
    ignoreErrors: [
      // 網路錯誤
      'ECONNREFUSED',
      'ETIMEDOUT',
      'ENOTFOUND',
      // 用戶取消
      'AbortError',
    ],

    // 在傳送前處理事件
    beforeSend(event, hint) {
      // 過濾開發環境的錯誤
      if (process.env.NODE_ENV === 'development') {
        console.error('[Sentry Server Dev]', hint.originalException || hint.syntheticException);
        return null;
      }

      return event;
    },
  });
}
