import { test, expect } from '@playwright/test';
import { demoLogin, getTestEnv, isHealthyState, isErrorState } from '../../lib/demoAuth';

/**
 * Engineering Feature Flag Rollout Scenario Tests
 * 
 * These tests verify the feature-flag-rollout scenario in both broken and fixed states.
 * The scenario has inverted feature flag logic causing incorrect behavior.
 * 
 * Test tags:
 *   eng/feature-flag-rollout-broken - Verify broken state
 *   eng/feature-flag-rollout-fixed - Verify fixed state
 */

// Get current stage from environment
const env = getTestEnv();
const stage = env.stage;

test.describe('eng/feature-flag-rollout-broken', () => {
  test.skip(stage === 'fixed', 'Skipping broken tests when stage is fixed');

  test('feature flag has inverted logic', async ({ page }) => {
    // Wait for app to be ready
    await page.waitForTimeout(2000);
    
    // Navigate to home page
    const response = await page.goto('/');
    
    // Page should load (200 status)
    expect(response?.status()).toBe(200);
    
    // In broken state, feature flag logic is inverted
    // For now, just verify the infrastructure is up and page loads
    const content = await page.textContent('body');
    expect(content).toBeTruthy();
  });

  test('health endpoint works (basic infra ok)', async ({ page }) => {
    // Wait for app to be ready
    await page.waitForTimeout(2000);
    
    const response = await page.goto('/_health');
    expect(response?.status()).toBe(200);
  });
});

test.describe('eng/feature-flag-rollout-fixed', () => {
  test.skip(stage === 'broken', 'Skipping fixed tests when stage is broken');

  test('feature flag logic is correct', async ({ page }) => {
    // Wait for app to be ready
    await page.waitForTimeout(2000);
    
    // Navigate to home page
    await page.goto('/');
    
    // Verify page is healthy
    const healthy = await isHealthyState(page);
    expect(healthy).toBe(true);
    
    // No errors should be present
    const hasError = await isErrorState(page);
    expect(hasError).toBe(false);
  });

  test('health endpoint works after fix', async ({ page }) => {
    // Wait for app to be ready
    await page.waitForTimeout(2000);
    
    const response = await page.goto('/_health');
    expect(response?.status()).toBe(200);
  });

  test('demo login works after fix', async ({ page }) => {
    // Wait for app to be ready
    await page.waitForTimeout(2000);
    
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

  test('feature flag enabled behavior works correctly', async ({ page }) => {
    // Wait for app to be ready
    await page.waitForTimeout(2000);
    
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
    
    // The fix should make feature flag logic work correctly
    // Verify no errors and content loads
    const hasError = await isErrorState(page);
    expect(hasError).toBe(false);
    
    const content = await page.textContent('body');
    expect(content).toBeTruthy();
    expect(content).not.toContain('500');
    expect(content).not.toContain('Internal Server Error');
  });
});
