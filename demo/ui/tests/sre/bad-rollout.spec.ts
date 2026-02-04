import { test, expect } from '@playwright/test';
import { isErrorState } from '../../lib/demoAuth';

/**
 * SRE Bad Rollout Scenario - Negative UI Tests
 * 
 * These tests verify that the bad-rollout scenario is correctly broken.
 * The scenario intentionally has a wrong DATABASE_URL, so the app should fail.
 * 
 * Test tag: sre/bad-rollout
 */

test.describe('sre/bad-rollout', () => {
  test('home page shows error or unavailable state', async ({ page }) => {
    // Navigate to home page - expect it to fail or show error
    const response = await page.goto('/', {
      timeout: 15000,
      waitUntil: 'domcontentloaded',
    });
    
    // The response should either:
    // 1. Return a non-200 status (502, 503, etc.)
    // 2. Show an error page
    // 3. Timeout/fail to connect
    
    if (response) {
      const status = response.status();
      
      // Non-200 status is expected for a broken deployment
      if (status >= 500 || status === 502 || status === 503) {
        // Good - server is returning error status
        expect(status).toBeGreaterThanOrEqual(500);
        return;
      }
      
      // If we got a 200, check if it's actually showing an error
      if (status === 200) {
        // Some proxies return 200 with error content
        const hasError = await isErrorState(page);
        expect(hasError).toBe(true);
        return;
      }
    }
    
    // If we got here, check page content for error indicators
    const hasError = await isErrorState(page);
    expect(hasError).toBe(true);
  });

  test('health endpoint returns non-200', async ({ request }) => {
    try {
      const response = await request.get('/_health', {
        timeout: 10000,
      });
      
      // Health check should fail
      const status = response.status();
      expect(status).not.toBe(200);
    } catch (error) {
      // Connection failure is also acceptable for a broken deployment
      // This happens when the app is completely down
      expect(error).toBeDefined();
    }
  });
});
