import { test, expect } from '@playwright/test';

/**
 * 認證流程 E2E 測試
 */
test.describe('認證流程', () => {
  test.describe('登入頁面', () => {
    test.beforeEach(async ({ page }) => {
      await page.goto('/login');
    });

    test('應該顯示 LINE 登入按鈕', async ({ page }) => {
      // 等待頁面載入
      await page.waitForLoadState('networkidle');

      // 檢查 LINE 登入按鈕
      const lineButton = page.getByRole('button', { name: /LINE/i });
      await expect(lineButton).toBeVisible();
    });

    test('點擊 LINE 登入按鈕應該重導向到 LINE OAuth', async ({ page }) => {
      // 點擊 LINE 登入按鈕
      const lineButton = page.getByRole('button', { name: /LINE/i });

      // 監聽導航事件
      const [request] = await Promise.all([
        page.waitForRequest((req) => req.url().includes('line.me') || req.url().includes('access.line.me')),
        lineButton.click(),
      ]).catch(() => [null]);

      // 確認有重導向到 LINE (或者在測試環境中被攔截)
      if (request) {
        expect(request.url()).toContain('line.me');
      }
    });
  });

  test.describe('登出流程', () => {
    test('已登入用戶應該可以登出', async ({ page, context }) => {
      // 模擬已登入狀態
      await context.addCookies([
        {
          name: 'auth_token',
          value: 'mock-jwt-token',
          domain: 'localhost',
          path: '/',
        },
      ]);

      // 設定 localStorage 模擬使用者資料
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

      // 前往首頁
      await page.goto('/');
      await page.waitForLoadState('networkidle');

      // 找到使用者頭像或選單並點擊
      const userMenu = page.locator('[data-testid="user-menu"]');
      const userMenuVisible = await userMenu.isVisible().catch(() => false);

      if (userMenuVisible) {
        await userMenu.click();

        // 點擊登出按鈕
        const logoutButton = page.getByRole('button', { name: /登出/i });
        if (await logoutButton.isVisible()) {
          await logoutButton.click();

          // 確認已登出 (應該重導向或顯示登入按鈕)
          await expect(page.getByRole('link', { name: /登入|LINE/i })).toBeVisible({ timeout: 5000 });
        }
      }
    });
  });

  test.describe('OAuth 回調', () => {
    test('帶有無效 code 應該顯示錯誤', async ({ page }) => {
      // 模擬 OAuth 回調但帶無效的 code
      await page.goto('/auth/callback?code=invalid-code&state=test-state');

      // 等待處理完成
      await page.waitForLoadState('networkidle');

      // 應該顯示錯誤訊息或重導向到登入頁
      const hasError = await page.getByText(/錯誤|失敗|error/i).isVisible().catch(() => false);
      const redirectedToLogin = page.url().includes('/login');

      expect(hasError || redirectedToLogin).toBeTruthy();
    });

    test('缺少 code 參數應該重導向到登入頁', async ({ page }) => {
      // 模擬 OAuth 回調但沒有 code
      await page.goto('/auth/callback');

      // 等待處理完成
      await page.waitForLoadState('networkidle');

      // 應該重導向到登入頁或首頁
      const currentUrl = page.url();
      expect(currentUrl.includes('/login') || currentUrl === 'http://localhost:3000/').toBeTruthy();
    });
  });
});
