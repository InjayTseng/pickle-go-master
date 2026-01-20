import { test, expect } from './fixtures/test-fixtures';

/**
 * 活動報名流程 E2E 測試
 */
test.describe('活動報名', () => {
  // 模擬活動資料
  const mockEvent = {
    id: 'test-event-id',
    shortCode: 'abc123',
    title: '週末匹克球活動',
    description: '歡迎一起來打球！',
    eventDate: '2024-12-20',
    startTime: '14:00',
    endTime: '16:00',
    locationName: '台北市大安運動中心',
    locationAddress: '台北市大安區...',
    latitude: 25.0330,
    longitude: 121.5430,
    capacity: 8,
    confirmedCount: 4,
    waitlistCount: 0,
    skillLevel: 'intermediate',
    fee: 100,
    status: 'open',
    host: {
      id: 'host-user-id',
      displayName: '主辦人',
      avatarUrl: 'https://via.placeholder.com/150',
    },
    registrations: [],
  };

  test.describe('活動詳情頁面', () => {
    test.beforeEach(async ({ page }) => {
      // Mock 活動 API
      await page.route('**/api/v1/events/*', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ success: true, data: mockEvent }),
        });
      });
    });

    test('應該顯示活動詳情', async ({ page }) => {
      await page.goto(`/events/${mockEvent.id}`);
      await page.waitForLoadState('networkidle');

      // 檢查活動標題
      await expect(page.getByText(mockEvent.title)).toBeVisible();

      // 檢查地點
      await expect(page.getByText(mockEvent.locationName)).toBeVisible();

      // 檢查報名狀態
      await expect(page.getByText(/4.*\/.*8|報名/)).toBeVisible();
    });

    test('未登入用戶點擊報名應該提示登入', async ({ page }) => {
      await page.goto(`/events/${mockEvent.id}`);
      await page.waitForLoadState('networkidle');

      // 點擊報名按鈕
      const registerButton = page.getByRole('button', { name: /報名|參加/i });
      if (await registerButton.isVisible()) {
        await registerButton.click();

        // 應該顯示登入提示或重導向
        const loginPrompt = page.getByText(/登入|LINE/i);
        const redirected = page.url().includes('/login');

        expect(await loginPrompt.isVisible() || redirected).toBeTruthy();
      }
    });
  });

  test.describe('已登入用戶報名', () => {
    test('應該可以成功報名活動', async ({ authenticatedPage }) => {
      // Mock 活動 API
      await authenticatedPage.route('**/api/v1/events/test-event-id', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ success: true, data: mockEvent }),
        });
      });

      // Mock 報名 API
      await authenticatedPage.route('**/api/v1/events/*/register', async (route) => {
        if (route.request().method() === 'POST') {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({
              success: true,
              data: {
                id: 'registration-id',
                eventId: mockEvent.id,
                userId: 'test-user-id',
                status: 'confirmed',
                registeredAt: new Date().toISOString(),
              },
            }),
          });
        } else {
          await route.continue();
        }
      });

      await authenticatedPage.goto(`/events/${mockEvent.id}`);
      await authenticatedPage.waitForLoadState('networkidle');

      // 點擊報名按鈕
      const registerButton = authenticatedPage.getByRole('button', { name: /報名|參加/i });
      await registerButton.click();

      // 等待回應
      await authenticatedPage.waitForTimeout(1000);

      // 確認成功（按鈕文字改變或顯示成功訊息）
      const successIndicator = authenticatedPage.getByText(/已報名|取消報名|成功/i);
      await expect(successIndicator).toBeVisible({ timeout: 5000 });
    });

    test('已報名的活動應該顯示取消選項', async ({ authenticatedPage }) => {
      // 修改 mockEvent 以包含使用者的報名
      const registeredEvent = {
        ...mockEvent,
        registrations: [
          {
            id: 'registration-id',
            userId: 'test-user-id',
            status: 'confirmed',
          },
        ],
      };

      // Mock 活動 API
      await authenticatedPage.route('**/api/v1/events/test-event-id', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ success: true, data: registeredEvent }),
        });
      });

      await authenticatedPage.goto(`/events/${mockEvent.id}`);
      await authenticatedPage.waitForLoadState('networkidle');

      // 應該顯示「已報名」或「取消報名」按鈕
      const cancelButton = authenticatedPage.getByRole('button', { name: /已報名|取消/i });
      await expect(cancelButton).toBeVisible();
    });
  });

  test.describe('短網址分享', () => {
    test('短網址應該正確重導向到活動頁面', async ({ page }) => {
      // Mock 短網址 API
      await page.route('**/api/v1/events/by-code/abc123', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ success: true, data: mockEvent }),
        });
      });

      await page.goto(`/g/${mockEvent.shortCode}`);
      await page.waitForLoadState('networkidle');

      // 應該顯示活動詳情
      await expect(page.getByText(mockEvent.title)).toBeVisible();
    });

    test('無效的短網址應該顯示 404', async ({ page }) => {
      // Mock 404 回應
      await page.route('**/api/v1/events/by-code/invalid', async (route) => {
        await route.fulfill({
          status: 404,
          contentType: 'application/json',
          body: JSON.stringify({ success: false, error: 'Event not found' }),
        });
      });

      await page.goto('/g/invalid');
      await page.waitForLoadState('networkidle');

      // 應該顯示錯誤訊息
      const notFound = page.getByText(/找不到|不存在|404/i);
      await expect(notFound).toBeVisible();
    });
  });
});
