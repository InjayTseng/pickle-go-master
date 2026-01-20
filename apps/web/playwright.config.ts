import { defineConfig, devices } from '@playwright/test';

/**
 * Pickle Go E2E 測試配置
 * @see https://playwright.dev/docs/test-configuration
 */
export default defineConfig({
  // 測試目錄
  testDir: './e2e',

  // 每個測試的最大執行時間
  timeout: 30 * 1000,

  // expect 斷言的最大等待時間
  expect: {
    timeout: 5000,
  },

  // 完整的平行執行
  fullyParallel: true,

  // 禁止只執行單一測試 (CI 環境)
  forbidOnly: !!process.env.CI,

  // 失敗重試次數
  retries: process.env.CI ? 2 : 0,

  // 平行執行的 worker 數量
  workers: process.env.CI ? 1 : undefined,

  // 測試報告器
  reporter: [
    ['html', { open: 'never' }],
    ['json', { outputFile: 'test-results/results.json' }],
    process.env.CI ? ['github'] : ['list'],
  ],

  // 所有測試共用的設定
  use: {
    // 基礎 URL
    baseURL: process.env.PLAYWRIGHT_BASE_URL || 'http://localhost:3000',

    // 收集失敗測試的 trace
    trace: 'on-first-retry',

    // 截圖設定
    screenshot: 'only-on-failure',

    // 影片錄製
    video: 'on-first-retry',

    // 視窗大小
    viewport: { width: 1280, height: 720 },

    // 忽略 HTTPS 錯誤 (開發環境)
    ignoreHTTPSErrors: true,

    // 語言設定
    locale: 'zh-TW',
    timezoneId: 'Asia/Taipei',
  },

  // 瀏覽器設定
  projects: [
    // 桌面瀏覽器測試
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'firefox',
      use: { ...devices['Desktop Firefox'] },
    },
    {
      name: 'webkit',
      use: { ...devices['Desktop Safari'] },
    },

    // 行動裝置測試
    {
      name: 'Mobile Chrome',
      use: { ...devices['Pixel 5'] },
    },
    {
      name: 'Mobile Safari',
      use: { ...devices['iPhone 12'] },
    },
  ],

  // 在執行測試前啟動開發伺服器
  webServer: {
    command: 'pnpm dev',
    url: 'http://localhost:3000',
    reuseExistingServer: !process.env.CI,
    timeout: 120 * 1000,
  },

  // 輸出目錄
  outputDir: 'test-results/',
});
