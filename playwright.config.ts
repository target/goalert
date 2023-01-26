import { devices } from '@playwright/test'
import { scanUniqueFlagCombos } from './test/integration/setup/scan-flags'

const dbURL =
  process.env.DB_URL ||
  process.env.GOALERT_DB_URL ||
  'postgres://goalert@localhost:5432/goalert_integration'

const wsEnv = {
  GOALERT_DB_URL: dbURL,
  GOALERT_ENGINE_CYCLE_TIME: '50ms',
  GOALERT_STRICT_EXPERIMENTAL: '1',
  GOALERT_LOG_ERRORS_ONLY: '1',
}

const config = {
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
        viewport: { width: 1440, height: 900 },
      },
    },
    {
      name: 'chromium-mobile',
      use: {
        ...devices['Pixel 5'],
        viewport: { width: 375, height: 667 },
      },
    },
  ],
  webServer: [
    {
      command: 'make start-integration CI=1',
      port: 6125,
      reuseExistingServer: true,
    },
    {
      command:
        './bin/goalert -l=localhost:6130 --public-url=http://localhost:6130',
      env: wsEnv,
      url: 'http://localhost:6130/health',
    },

    // generate a web server for each unique flag combination
    ...scanUniqueFlagCombos().map((flagStr, i) => ({
      command: `./bin/goalert -l=localhost:${
        i + 6131
      } --public-url=http://localhost:${i + 6131} --experimental=${flagStr}`,
      env: wsEnv,
      url: `http://localhost:${i + 6131}/health`,
    })),
  ],
}

export default config
