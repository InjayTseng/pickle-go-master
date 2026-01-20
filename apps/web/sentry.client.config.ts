/**
 * Sentry 客戶端配置
 * 用於瀏覽器端錯誤監控
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
    // 在生產環境使用較低的取樣率以減少成本
    tracesSampleRate: process.env.NODE_ENV === 'production' ? 0.1 : 1.0,

    // 錯誤取樣率
    sampleRate: 1.0,

    // 設定要忽略的錯誤
    ignoreErrors: [
      // 瀏覽器擴充功能錯誤
      /^chrome-extension:\/\//,
      /^moz-extension:\/\//,
      // 網路錯誤
      'Network Error',
      'Failed to fetch',
      'Load failed',
      // 用戶取消操作
      'AbortError',
      // React 開發模式錯誤
      /Minified React error/,
    ],

    // 設定要忽略的 URL
    denyUrls: [
      // Chrome 擴充功能
      /extensions\//i,
      /^chrome:\/\//i,
      // Facebook floc
      /connect\.facebook\.net/i,
      // 廣告相關
      /googlesyndication\.com/i,
      /googleadservices\.com/i,
    ],

    // 在傳送前處理事件
    beforeSend(event, hint) {
      // 過濾開發環境的錯誤
      if (process.env.NODE_ENV === 'development') {
        // 在開發環境中，可以選擇只記錄到 console
        console.error('[Sentry Dev]', hint.originalException || hint.syntheticException);
        return null; // 不傳送到 Sentry
      }

      // 過濾非關鍵錯誤
      const exception = event.exception?.values?.[0];
      if (exception) {
        // 過濾 ChunkLoadError（通常是部署後的快取問題）
        if (exception.type === 'ChunkLoadError') {
          return null;
        }
      }

      return event;
    },

    // 要追蹤的目標
    tracePropagationTargets: [
      'localhost',
      /^https:\/\/api\.picklego\.tw/,
    ],

    // 整合設定
    integrations: [
      // 啟用瀏覽器追蹤
      Sentry.browserTracingIntegration(),
      // 啟用 Replay（錯誤回放）
      Sentry.replayIntegration({
        // 只在錯誤發生時記錄
        maskAllText: true,
        blockAllMedia: true,
      }),
    ],

    // Replay 取樣率
    replaysSessionSampleRate: 0.1, // 10% 的正常會話
    replaysOnErrorSampleRate: 1.0, // 100% 的錯誤會話
  });
}
