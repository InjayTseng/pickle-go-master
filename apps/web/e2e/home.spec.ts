import { test, expect } from '@playwright/test';

/**
 * 首頁 E2E 測試
 */
test.describe('首頁', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('應該正確顯示首頁標題', async ({ page }) => {
    // 等待頁面載入
    await page.waitForLoadState('networkidle');

    // 檢查標題
    await expect(page).toHaveTitle(/Pickle Go/);
  });

  test('應該顯示導航列', async ({ page }) => {
    // 檢查 Header 存在
    const header = page.locator('header');
    await expect(header).toBeVisible();

    // 檢查 Logo 或品牌名稱
    await expect(page.getByRole('link', { name: /pickle/i })).toBeVisible();
  });

  test('應該顯示地圖元件', async ({ page }) => {
    // 等待地圖載入
    await page.waitForTimeout(2000);

    // 檢查地圖容器存在
    const mapContainer = page.locator('[data-testid="event-map"]');
    // 如果沒有 data-testid，檢查 Google Maps 相關元素
    const googleMap = page.locator('.gm-style');

    // 至少其中一個應該存在
    const mapVisible = await mapContainer.isVisible().catch(() => false);
    const googleMapVisible = await googleMap.isVisible().catch(() => false);

    expect(mapVisible || googleMapVisible).toBeTruthy();
  });

  test('應該在行動裝置顯示底部導航', async ({ page }) => {
    // 設定行動裝置視窗大小
    await page.setViewportSize({ width: 375, height: 667 });

    // 等待響應式調整
    await page.waitForTimeout(500);

    // 檢查行動裝置導航存在
    const mobileNav = page.locator('nav').last();
    await expect(mobileNav).toBeVisible();
  });

  test('未登入時應該顯示登入按鈕', async ({ page }) => {
    // 清除任何已存在的 cookies
    await page.context().clearCookies();

    // 重新載入頁面
    await page.reload();
    await page.waitForLoadState('networkidle');

    // 檢查登入按鈕或連結
    const loginButton = page.getByRole('link', { name: /登入|LINE/i });
    const loginExists = await loginButton.isVisible().catch(() => false);

    // 登入按鈕應該存在（可能在 header 或其他位置）
    expect(loginExists).toBeTruthy();
  });
});
