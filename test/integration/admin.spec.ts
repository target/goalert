import { test, expect } from '@playwright/test'
import { adminSessionFile } from './lib'
import Chance from 'chance'
import { createService } from './lib/service'
const c = new Chance()
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

  const name = 'pw-service ' + c.name()
  const description = c.sentence()
  await createService(page, name, description)

  await page.goto('./admin/service-metrics')

  await testMetricsOverview('1Total Services')
  await testMetricsOverview('1Services Missing Integrations')
  await testMetricsOverview('0Services Missing Notifications')
  await testMetricsOverview('0Services Reaching Alert Limit')
})
