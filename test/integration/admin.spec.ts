import { test, expect } from '@playwright/test'
import { adminSessionFile } from './lib'
test.describe.configure({ mode: 'parallel' })
test.use({ storageState: adminSessionFile })

test('Admin', async ({ page }) => {
  const testMetricsOverview = async (text: string): Promise<void> => {
    await expect(
      page.locator('.MuiCard-root', {
        has: page.locator('div').filter({ hasText: text }),
      }),
    ).toBeVisible()
  }
  await page.goto('./admin/service-metrics')

  testMetricsOverview('3Total Services')
  testMetricsOverview('6Services Missing Integrations')
  testMetricsOverview('0Services Missing Notifications')
  testMetricsOverview('0Services Reaching Alert Limit')
})
