import { test, expect } from '@playwright/test'
import { adminSessionFile } from './lib'
import Chance from 'chance'
import { createService } from './lib/service'
const c = new Chance()
test.describe.configure({ mode: 'parallel' })
test.use({ storageState: adminSessionFile })

test('Admin', async ({ page }) => {
  const name = 'pw-service ' + c.name()
  const description = c.sentence()
  await createService(page, name, description)

  await page.goto('./admin/service-metrics')

  // wait for services to finish loading
  await expect(page.locator('.MuiDataGrid-overlay')).toHaveCount(0)

  const totalServicesLocator = page
    .locator('.MuiCardHeader-content', { hasText: 'Total Services' })
    .locator('.MuiCardHeader-title')

  // This will retry until the locator's text matches the regex (a number greater than 0)
  // or the timeout is reached.
  await expect(totalServicesLocator).toHaveText(/^[1-9]\d*$/)

  // get services missing integrations count
  const missingIntsLocator = page
    .locator('.MuiCardHeader-content', {
      hasText: 'Services With No Integrations',
    })
    .locator('.MuiCardHeader-title')
  await expect(missingIntsLocator).toHaveText(/^[1-9]\d*$/)

  // get services missing notifications count
  const missingNotifLocator = page
    .locator('.MuiCardHeader-content', {
      hasText: 'Services With Empty Escalation Policies',
    })
    .locator('.MuiCardHeader-title')
  await expect(missingNotifLocator).toHaveText('0')

  // get services reaching alert limit
  const alertLimitLocator = page
    .locator('.MuiCardHeader-content', {
      hasText: 'Services Reaching Alert Limit',
    })
    .locator('.MuiCardHeader-title')
  await expect(alertLimitLocator).toHaveText('0')
})
