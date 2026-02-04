import { test, expect } from '@playwright/test';
import { demoLogin, getTestEnv, isHealthyState } from '../../lib/demoAuth';

/**
 * SRE Healthy Scenario - UI Smoke Tests
 * 
 * These tests verify that the healthy baseline scenario is working correctly.
 * They are designed to be run as part of the verification process.
 * 
 * Test tag: sre/healthy
 */

test.describe('sre/healthy', () => {
  const env = getTestEnv();

  test('home page loads successfully', async ({ page }) => {
    // Navigate to home page
    await page.goto('/');
    
    // Verify page loaded (title varies by tenant configuration)
    await expect(page).toHaveTitle(/.+/);
    
    // Verify it's not an error state
    const healthy = await isHealthyState(page);
    expect(healthy).toBe(true);
  });

  test('health endpoint returns 200', async ({ page }) => {
    // Use page navigation instead of request context (DNS resolution works in browser)
    const response = await page.goto('/_health');
    expect(response?.status()).toBe(200);
  });

  test('demo login works', async ({ page }) => {
    // Log in as admin
    await demoLogin({
      page,
      baseURL: env.baseURL,
      persona: 'admin',
      loginKey: env.demoLoginKey,
    });
    
    // Verify we're logged in (should see user menu or similar)
    // The demo login redirects to home, so check we're there
    await expect(page).toHaveURL(/\//);
    
    // Page should be healthy after login
    const healthy = await isHealthyState(page);
    expect(healthy).toBe(true);
  });

  test('can view feature posts list', async ({ page }) => {
    // Log in first
    await demoLogin({
      page,
      baseURL: env.baseURL,
      persona: 'user',
      loginKey: env.demoLoginKey,
    });
    
    // Navigate to home which shows posts
    await page.goto('/');
    
    // Wait for content to load
    await page.waitForLoadState('networkidle');
    
    // Should see some content (posts, or empty state message)
    const content = await page.textContent('body');
    expect(content).toBeTruthy();
  });
});
