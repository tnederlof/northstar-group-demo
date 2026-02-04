import { Page } from '@playwright/test';

/**
 * Demo authentication helper for Fider application.
 * 
 * Uses the /__demo/login/:persona endpoint that is available when
 * the application is running in demo mode.
 */

export type Persona = 'admin' | 'user' | 'visitor';

interface DemoLoginOptions {
  /** The Playwright page instance */
  page: Page;
  /** Base URL of the application */
  baseURL: string;
  /** Persona to log in as */
  persona?: Persona;
  /** Demo login key (from DEMO_LOGIN_KEY env var) */
  loginKey?: string;
}

/**
 * Log in using the demo login endpoint.
 * 
 * @example
 * ```ts
 * await demoLogin({
 *   page,
 *   baseURL: process.env.BASE_URL,  // e.g., 'http://healthy.localhost:8080' for SRE
 *   persona: 'admin',
 *   loginKey: process.env.DEMO_LOGIN_KEY,
 * });
 * ```
 */
export async function demoLogin({
  page,
  baseURL,
  persona = 'admin',
  loginKey,
}: DemoLoginOptions): Promise<void> {
  const key = loginKey || process.env.DEMO_LOGIN_KEY || 'northstar-demo-key';
  
  const url = new URL(`/__demo/login/${persona}`, baseURL);
  url.searchParams.set('key', key);
  
  // Navigate to demo login endpoint
  const response = await page.goto(url.toString(), {
    waitUntil: 'networkidle',
  });
  
  // The demo login endpoint redirects to the home page on success
  if (!response) {
    throw new Error('No response from demo login endpoint');
  }
  
  // Wait for redirect to complete (demo login redirects to /)
  // Use a longer timeout and wait for network to settle
  await page.waitForLoadState('networkidle', { timeout: 15000 });
}

/**
 * Get environment variables for the current test run.
 */
export function getTestEnv() {
  return {
    baseURL: process.env.BASE_URL!,  // Required: set by run-checks.sh with track-specific port
    scenario: process.env.SCENARIO || '',
    stage: process.env.STAGE || '',
    demoLoginKey: process.env.DEMO_LOGIN_KEY || 'northstar-demo-key',
  };
}

/**
 * Check if the app is showing an error state.
 * Useful for negative tests (verifying broken state).
 */
export async function isErrorState(page: Page): Promise<boolean> {
  // Check for common error indicators
  const errorSelectors = [
    // HTTP error pages
    'text=500',
    'text=502',
    'text=503',
    'text=Internal Server Error',
    'text=Service Unavailable',
    // Application error states
    '[data-testid="error"]',
    '.error-page',
    '.error-message',
  ];
  
  for (const selector of errorSelectors) {
    try {
      const element = await page.$(selector);
      if (element) {
        return true;
      }
    } catch {
      // Selector not found, continue
    }
  }
  
  return false;
}

/**
 * Check if the app is in a healthy/working state.
 */
export async function isHealthyState(page: Page): Promise<boolean> {
  // Check for indicators of a working application
  try {
    // Wait for the page to have meaningful content
    await page.waitForSelector('body', { state: 'visible', timeout: 5000 });
    
    // Check that it's not an error state
    const hasError = await isErrorState(page);
    if (hasError) {
      return false;
    }
    
    // Check for Fider-specific healthy indicators
    const healthySelectors = [
      // Main navigation or header
      'nav',
      'header',
      // Fider-specific elements
      '.c-menu',
      '[data-testid="home"]',
      // Any post list (indicates app is working)
      '.c-post-list',
      '.p-home__ideas',
    ];
    
    for (const selector of healthySelectors) {
      const element = await page.$(selector);
      if (element) {
        return true;
      }
    }
    
    return false;
  } catch {
    return false;
  }
}
