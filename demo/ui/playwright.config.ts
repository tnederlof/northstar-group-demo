import { defineConfig, devices } from '@playwright/test';

/**
 * Playwright configuration for scenario verification tests.
 * 
 * Environment variables:
 *   BASE_URL - Base URL for the scenario (e.g., http://healthy.localhost:8080)
 *   SCENARIO - Scenario path (e.g., platform/healthy)
 *   STAGE - Current stage (e.g., broken, healthy, fixed)
 *   DEMO_LOGIN_KEY - Authentication key for demo login
 */
export default defineConfig({
  testDir: './tests',
  
  // Fail fast in CI
  fullyParallel: false,
  
  // Retry failed tests in CI
  retries: process.env.CI ? 2 : 0,
  
  // Single worker for deterministic execution
  workers: 1,
  
  // Reporter configuration - always use list for inline output, never open browser
  reporter: process.env.CI 
    ? [['github'], ['html', { open: 'never' }]]
    : [['list'], ['html', { open: 'never' }]],
  
  // Global test timeout
  timeout: 30000,
  
  // Expect timeout for assertions
  expect: {
    timeout: 5000,
  },

  use: {
    // Base URL from environment (set by run-checks.sh with track-specific port)
    baseURL: process.env.BASE_URL,
    
    // Capture artifacts on failure
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'on-first-retry',
    
    // Extra HTTP headers if needed
    extraHTTPHeaders: {
      'Accept': 'text/html,application/xhtml+xml,application/json',
    },
  },

  // Output directory for artifacts
  outputDir: './test-results',

  // Projects for different test categories
  projects: [
    {
      name: 'chromium',
      use: {
        ...devices['Desktop Chrome'],
        // Fix DNS resolution for .localhost domains
        launchOptions: {
          args: [
            '--host-resolver-rules=MAP *.localhost 127.0.0.1',
          ],
        },
      },
    },
  ],

  // Web server configuration (not used - scenarios are already running)
  // webServer: undefined,
});
