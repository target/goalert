import { devices } from '@playwright/test'
import type { PlaywrightTestConfig } from '@playwright/test'

const config: PlaywrightTestConfig = {
  testDir: './test/integration',
  globalSetup: require.resolve('./test/integration/setup/global-setup.ts'),
  retries: 3,
  use: {
    trace: 'on-first-retry',
    baseURL: 'http://localhost:6130',
    viewport: { width: 1440, height: 900 },
    timezoneId: 'America/Chicago',
    launchOptions: {
      // slowMo: 1000,
    },
    actionTimeout: 5000,
  },
  projects: [
    {
      name: 'chromium-wide',
      use: {
        ...devices['Desktop Chrome'],
        headless: true,
        viewport: { width: 1440, height: 900 },
      },
    },
    {
      name: 'chromium-mobile',
      use: {
        ...devices['Pixel 5'],
        headless: true,
        viewport: { width: 375, height: 667 },
      },
    },
  ],
  webServer: {
    command: 'make start-integration CI=1',
    url: 'http://localhost:6130',
    reuseExistingServer: true,
  },
}

module.exports = config
