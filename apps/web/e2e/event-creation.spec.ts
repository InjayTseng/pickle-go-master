import { test, expect } from './fixtures/test-fixtures';

/**
 * 活動建立流程 E2E 測試
 */
test.describe('活動建立', () => {
  test.describe('未登入用戶', () => {
    test('應該重導向到登入頁面', async ({ page }) => {
      // 嘗試訪問建立活動頁面
      await page.goto('/events/new');

      // 等待導航完成
      await page.waitForLoadState('networkidle');

      // 應該被重導向到登入頁面
      expect(page.url()).toContain('/login');
    });
  });

  test.describe('已登入用戶', () => {
    test('應該顯示活動建立表單', async ({ authenticatedPage }) => {
      // 前往建立活動頁面
      await authenticatedPage.goto('/events/new');
      await authenticatedPage.waitForLoadState('networkidle');

      // 檢查表單元素存在
      await expect(authenticatedPage.getByLabel(/地點|場地/i)).toBeVisible();
      await expect(authenticatedPage.getByLabel(/日期/i)).toBeVisible();
      await expect(authenticatedPage.getByLabel(/時間|開始/i)).toBeVisible();
      await expect(authenticatedPage.getByLabel(/人數|名額/i)).toBeVisible();
      await expect(authenticatedPage.getByLabel(/程度|等級/i)).toBeVisible();
    });

    test('應該驗證必填欄位', async ({ authenticatedPage }) => {
      await authenticatedPage.goto('/events/new');
      await authenticatedPage.waitForLoadState('networkidle');

      // 直接點擊送出按鈕（不填任何欄位）
      const submitButton = authenticatedPage.getByRole('button', { name: /建立|送出|確認/i });
      await submitButton.click();

      // 等待驗證訊息
      await authenticatedPage.waitForTimeout(500);

      // 應該顯示錯誤訊息
      const errorMessages = authenticatedPage.locator('[role="alert"], .error, .text-destructive');
      const errorCount = await errorMessages.count();
      expect(errorCount).toBeGreaterThan(0);
    });

    test('應該可以填寫表單並建立活動', async ({ authenticatedPage }) => {
      // Mock API 回應
      await authenticatedPage.route('**/api/v1/events', async (route) => {
        if (route.request().method() === 'POST') {
          await route.fulfill({
            status: 201,
            contentType: 'application/json',
            body: JSON.stringify({
              success: true,
              data: {
                id: 'new-event-id',
                shortCode: 'abc123',
                title: '測試活動',
                locationName: '台北市大安運動中心',
                eventDate: '2024-12-20',
                startTime: '14:00',
                capacity: 8,
                skillLevel: 'intermediate',
              },
            }),
          });
        } else {
          await route.continue();
        }
      });

      await authenticatedPage.goto('/events/new');
      await authenticatedPage.waitForLoadState('networkidle');

      // 填寫地點 (假設有 places autocomplete)
      const locationInput = authenticatedPage.getByLabel(/地點|場地/i);
      await locationInput.fill('台北市大安運動中心');

      // 選擇日期
      const dateInput = authenticatedPage.getByLabel(/日期/i);
      await dateInput.fill('2024-12-20');

      // 選擇時間
      const timeInput = authenticatedPage.getByLabel(/開始時間|時間/i);
      await timeInput.fill('14:00');

      // 選擇人數
      const capacitySelect = authenticatedPage.getByLabel(/人數|名額/i);
      await capacitySelect.selectOption('8');

      // 選擇程度
      const skillSelect = authenticatedPage.getByLabel(/程度|等級/i);
      await skillSelect.selectOption('intermediate');

      // 送出表單
      const submitButton = authenticatedPage.getByRole('button', { name: /建立|送出|確認/i });
      await submitButton.click();

      // 等待導航或成功訊息
      await authenticatedPage.waitForURL(/\/events\/.*\/created|\/events\//, { timeout: 10000 }).catch(() => {});

      // 確認成功（URL 改變或顯示成功訊息）
      const successMessage = authenticatedPage.getByText(/成功|建立完成/i);
      const urlChanged = !authenticatedPage.url().includes('/events/new');

      const hasSuccess = await successMessage.isVisible().catch(() => false);
      expect(hasSuccess || urlChanged).toBeTruthy();
    });
  });
});
