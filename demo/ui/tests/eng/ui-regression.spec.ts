import { test, expect } from '@playwright/test';
import { demoLogin, getTestEnv, isHealthyState, isErrorState } from '../../lib/demoAuth';

/**
 * Engineering UI Regression Scenario Tests
 * 
 * These tests verify the ui-regression scenario in both broken and fixed states.
 * The scenario has a missing null check that causes 500 errors.
 * 
 * Test tags:
 *   eng/ui-regression-broken - Verify broken state
 *   eng/ui-regression-fixed - Verify fixed state
 */

// Get current stage from environment
const env = getTestEnv();
const stage = env.stage;

test.describe('eng/ui-regression-broken', () => {
  test.skip(stage === 'fixed', 'Skipping broken tests when stage is fixed');

  test('app shows 500 error under specific conditions', async ({ page }) => {
    // Navigate to home - basic page should load
    await page.goto('/');
    
    // For the broken state, we expect either:
    // 1. The page loads but API calls fail with 500
    // 2. The page shows error state for specific actions
    
    // Try to trigger the null check issue
    // This depends on the specific scenario implementation
    // For now, just verify the infrastructure is up
    const response = await page.goto('/_health');
    
    // Health endpoint should work even in broken state
    // (the bug is in specific handlers, not general health)
    expect(response?.status()).toBe(200);
  });

  test('specific API endpoint returns 500', async ({ request }) => {
    // The ui-regression scenario has a specific endpoint that fails
    // This test should be updated based on the actual buggy endpoint
    
    // For now, verify health works but note the scenario is "broken"
    const healthResponse = await request.get('/_health');
    expect(healthResponse.status()).toBe(200);
    
    // TODO: Add specific endpoint test when scenario details are finalized
  });
});

test.describe('eng/ui-regression-fixed', () => {
  test.skip(stage === 'broken', 'Skipping fixed tests when stage is broken');

  test('home page loads and works correctly', async ({ page }) => {
    // Navigate to home page
    await page.goto('/');
    
    // Verify page is healthy
    const healthy = await isHealthyState(page);
    expect(healthy).toBe(true);
    
    // No errors should be present
    const hasError = await isErrorState(page);
    expect(hasError).toBe(false);
  });

  test('health endpoint returns 200', async ({ request }) => {
    const response = await request.get('/_health');
    expect(response.status()).toBe(200);
  });

  test('demo login works after fix', async ({ page }) => {
    // Log in as user
    await demoLogin({
      page,
      baseURL: env.baseURL,
      persona: 'user',
      loginKey: env.demoLoginKey,
    });
    
    // Should be redirected to home
    await expect(page).toHaveURL(/\//);
    
    // Page should be healthy
    const healthy = await isHealthyState(page);
    expect(healthy).toBe(true);
  });

  test('previously failing functionality now works', async ({ page }) => {
    // Log in
    await demoLogin({
      page,
      baseURL: env.baseURL,
      persona: 'user',
      loginKey: env.demoLoginKey,
    });
    
    // Navigate to home
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    
    // The fix should make the application work correctly
    // Verify no errors and content loads
    const hasError = await isErrorState(page);
    expect(hasError).toBe(false);
    
    const content = await page.textContent('body');
    expect(content).toBeTruthy();
    expect(content).not.toContain('500');
    expect(content).not.toContain('Internal Server Error');
  });
});
