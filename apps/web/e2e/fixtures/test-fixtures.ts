import { test as base, expect, Page } from '@playwright/test';

/**
 * 測試用的假使用者資料
 */
export interface TestUser {
  id: string;
  lineUserId: string;
  displayName: string;
  avatarUrl: string;
}

/**
 * 測試用的假活動資料
 */
export interface TestEvent {
  id: string;
  title: string;
  locationName: string;
  eventDate: string;
  startTime: string;
  capacity: number;
  skillLevel: string;
}

/**
 * 擴展的測試 fixtures
 */
export const test = base.extend<{
  // 已登入的頁面
  authenticatedPage: Page;
  // 測試用戶
  testUser: TestUser;
}>({
  // 提供已模擬登入的頁面
  authenticatedPage: async ({ page }, use) => {
    // 設定模擬的 auth token
    await page.context().addCookies([
      {
        name: 'auth_token',
        value: 'mock-jwt-token-for-testing',
        domain: 'localhost',
        path: '/',
        httpOnly: true,
        secure: false,
        sameSite: 'Lax',
      },
    ]);

    // 模擬 localStorage 中的使用者資料
    await page.addInitScript(() => {
      window.localStorage.setItem(
        'user',
        JSON.stringify({
          id: 'test-user-id',
          displayName: '測試用戶',
          avatarUrl: 'https://via.placeholder.com/150',
        })
      );
    });

    await use(page);
  },

  // 測試用戶資料
  testUser: async ({}, use) => {
    const user: TestUser = {
      id: 'test-user-id',
      lineUserId: 'U1234567890abcdef',
      displayName: '測試用戶',
      avatarUrl: 'https://via.placeholder.com/150',
    };
    await use(user);
  },
});

export { expect };

/**
 * 輔助函數：等待頁面載入完成
 */
export async function waitForPageLoad(page: Page) {
  await page.waitForLoadState('networkidle');
}

/**
 * 輔助函數：等待 Toast 訊息出現
 */
export async function waitForToast(page: Page, text: string) {
  await expect(page.getByText(text)).toBeVisible({ timeout: 10000 });
}

/**
 * 輔助函數：截圖並儲存
 */
export async function takeScreenshot(page: Page, name: string) {
  await page.screenshot({ path: `test-results/screenshots/${name}.png` });
}
